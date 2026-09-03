package factory

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cherry/api/internal/activate"
	"github.com/cherry/api/internal/gdpr"
	"github.com/cherry/api/internal/llm"
	"github.com/cherry/api/internal/maestro"
	"github.com/cherry/api/internal/opencode"
	"github.com/cherry/api/internal/store"
)

type MaestroRunner interface {
	RunDir(ctx context.Context, maestroDir, localURL string) maestro.Report
}

type Activator interface {
	Start(ctx context.Context, id, projectRoot string) (activate.Snapshot, error)
	Stop(id string) activate.Snapshot
	Snapshot(id string) activate.Snapshot
}

type Service struct {
	Store      store.Store
	Root       string
	StepDelay  time.Duration
	AutoRun    bool
	LLM        *llm.Service
	OpenCode   opencode.Runner
	Activator  Activator
	MaestroRun MaestroRunner
	mu         sync.Mutex
	reports    map[string]maestro.Report
}

func New(st store.Store, root string) *Service {
	if strings.TrimSpace(root) == "" {
		root = filepath.Join("var", "projects")
	}
	if abs, err := filepath.Abs(root); err == nil {
		root = abs
	}
	return &Service{Store: st, Root: root, StepDelay: 380 * time.Millisecond, AutoRun: true, reports: map[string]maestro.Report{}}
}

func (s *Service) Create(ctx context.Context, userID, name, brief, stackRaw, backendRaw string) (store.Project, error) {
	name = strings.TrimSpace(name)
	brief = strings.TrimSpace(brief)
	if name == "" || len(brief) < 8 {
		return store.Project{}, store.ErrValidation
	}
	stack, err := parseStack(stackRaw)
	if err != nil {
		return store.Project{}, err
	}
	backend, err := parseBackendTarget(backendRaw)
	if err != nil {
		return store.Project{}, err
	}
	if backend != store.TargetLocal {
		kind := store.ConnectionKind(backend)
		conn, err := s.Store.GetConnection(ctx, userID, kind)
		if err != nil || conn.Status != store.ConnConnected {
			return store.Project{}, fmt.Errorf("%w: %s bağlı değil. Bağlantılar menüsünden ekle.", store.ErrValidation, backend)
		}
	}
	now := time.Now().UTC()
	id := store.NewID()
	project := store.Project{
		ID:        id,
		UserID:    userID,
		Name:      name,
		Brief:     brief,
		Stack:     stack,
		Status:    store.StatusQueued,
		RootPath:  filepath.Join(s.Root, userID, id),
		Backend:   backend,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := os.MkdirAll(project.RootPath, 0o755); err != nil {
		return store.Project{}, err
	}
	created, err := s.Store.CreateProject(ctx, project)
	if err != nil {
		return store.Project{}, err
	}
	if err := s.log(ctx, created.ID, "Kuyruğa alındı. Ajan arka planda yazacak."); err != nil {
		return store.Project{}, err
	}
	if err := s.logRole(ctx, created.ID, created.Brief, store.RoleUser); err != nil {
		return store.Project{}, err
	}
	if s.AutoRun {
		go s.run(created.ID)
	}
	return created, nil
}

func (s *Service) Get(ctx context.Context, userID, id string) (store.Project, error) {
	project, err := s.Store.GetProject(ctx, id)
	if err != nil {
		return store.Project{}, err
	}
	if project.UserID != userID {
		return store.Project{}, store.ErrNotFound
	}
	return *project, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]store.Project, error) {
	return s.Store.Projects(ctx, userID)
}

func (s *Service) Logs(ctx context.Context, userID, id string) ([]store.JobLog, error) {
	if _, err := s.Get(ctx, userID, id); err != nil {
		return nil, err
	}
	return s.Store.ListLogs(ctx, id)
}

func (s *Service) ZipPath(ctx context.Context, userID, id string) (string, error) {
	project, err := s.Get(ctx, userID, id)
	if err != nil {
		return "", err
	}
	path := filepath.Join(project.RootPath, "cherry.zip")
	if _, err := os.Stat(path); err != nil {
		return "", store.ErrNotFound
	}
	return path, nil
}

func (s *Service) run(id string) {
	ctx := context.Background()
	project, err := s.Store.GetProject(ctx, id)
	if err != nil {
		return
	}
	if err := s.pipeline(ctx, *project); err != nil {
		_ = s.fail(ctx, id, err)
	}
}

func (s *Service) pipeline(ctx context.Context, project store.Project) error {
	s.pause()
	if err := s.setStatus(ctx, project.ID, store.StatusWriting); err != nil {
		return err
	}
	label, err := stackLabel(project.Stack)
	if err != nil {
		return err
	}
	if err := s.log(ctx, project.ID, "Yığın: "+label+". frontend/, backend/, maestro/ yazılıyor."); err != nil {
		return err
	}
	s.pause()
	if err := writeTree(project); err != nil {
		return err
	}
	if err := initCustomerRepo(project.RootPath); err != nil {
		return err
	}
	if err := s.log(ctx, project.ID, "Dosya ağacı hazır: "+project.RootPath); err != nil {
		return err
	}
	if err := opencode.WriteConfig(project.RootPath); err != nil {
		return err
	}
	briefSafe, _ := gdpr.Redact(project.Brief)
	rule, err := stackSourceRule(project.Stack)
	if err != nil {
		return err
	}
	if err := opencode.WriteAgents(project.RootPath, project.Name, label, briefSafe, rule); err != nil {
		return err
	}
	s.pause()
	if err := s.runLLM(ctx, project); err != nil {
		return err
	}
	s.pause()
	if err := s.runOpenCode(ctx, project); err != nil {
		return err
	}
	s.pause()
	if err := s.setStatus(ctx, project.ID, store.StatusTesting); err != nil {
		return err
	}
	if err := s.log(ctx, project.ID, "Test aşaması. Yerel API + Maestro."); err != nil {
		return err
	}
	if err := s.runLocalTest(ctx, project); err != nil {
		return err
	}
	s.pause()
	zipPath := filepath.Join(project.RootPath, "cherry.zip")
	if err := zipProject(project.RootPath, zipPath); err != nil {
		return err
	}
	if err := s.setStatus(ctx, project.ID, store.StatusReady); err != nil {
		return err
	}
	return s.log(ctx, project.ID, "Hazır. Zip seçilen dilin kaynağı (HTML site değil). Yazıcı: OpenCode. Yerel API durdu.")
}

func (s *Service) runLocalTest(ctx context.Context, project store.Project) error {
	url := ""
	if s.Activator != nil {
		snap, err := s.Activator.Start(ctx, project.ID, project.RootPath)
		if err != nil {
			if logErr := s.log(ctx, project.ID, "Yerel aktif olmadı: "+snap.Note); logErr != nil {
				return logErr
			}
		} else {
			url = snap.URL
			if logErr := s.log(ctx, project.ID, "Yerel aktif: "+snap.Note); logErr != nil {
				return logErr
			}
		}
		defer func() { _ = s.Activator.Stop(project.ID) }()
	} else if logErr := s.log(ctx, project.ID, "Yerel aktif bağlı değil — Maestro yine SKIPPED olabilir."); logErr != nil {
		return logErr
	}
	if s.MaestroRun == nil {
		return s.log(ctx, project.ID, "Maestro koşucu yok. Akışlar SKIPPED.")
	}
	report := s.MaestroRun.RunDir(ctx, filepath.Join(project.RootPath, "maestro"), url)
	s.mu.Lock()
	if s.reports == nil {
		s.reports = map[string]maestro.Report{}
	}
	s.reports[project.ID] = report
	s.mu.Unlock()
	for _, flow := range report.Flows {
		msg := flow.ID + ".yaml → " + string(flow.Result) + ". " + flow.Note
		if err := s.log(ctx, project.ID, msg); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Activate(ctx context.Context, userID, id string) (store.Project, error) {
	project, err := s.Get(ctx, userID, id)
	if err != nil {
		return store.Project{}, err
	}
	if s.Activator == nil {
		return project, fmt.Errorf("%w: yerel aktif yok", store.ErrNotFound)
	}
	snap, err := s.Activator.Start(ctx, project.ID, project.RootPath)
	note := snap.Note
	if err != nil {
		_ = s.log(ctx, project.ID, "Yerel aktif hata: "+note)
		return project, err
	}
	_ = s.log(ctx, project.ID, "Yerel aktif: "+note)
	return s.Get(ctx, userID, id)
}

func (s *Service) Deactivate(ctx context.Context, userID, id string) (store.Project, error) {
	project, err := s.Get(ctx, userID, id)
	if err != nil {
		return store.Project{}, err
	}
	if s.Activator == nil {
		return project, nil
	}
	snap := s.Activator.Stop(project.ID)
	_ = s.log(ctx, project.ID, snap.Note)
	return s.Get(ctx, userID, id)
}

func (s *Service) RunMaestro(ctx context.Context, userID, id string) (store.Project, error) {
	project, err := s.Get(ctx, userID, id)
	if err != nil {
		return store.Project{}, err
	}
	if s.MaestroRun == nil {
		_ = s.log(ctx, project.ID, "Maestro koşucu yok.")
		return project, nil
	}
	url := ""
	if s.Activator != nil {
		snap := s.Activator.Snapshot(project.ID)
		if snap.Status == activate.StatusRunning {
			url = snap.URL
		}
	}
	report := s.MaestroRun.RunDir(ctx, filepath.Join(project.RootPath, "maestro"), url)
	s.mu.Lock()
	if s.reports == nil {
		s.reports = map[string]maestro.Report{}
	}
	s.reports[project.ID] = report
	s.mu.Unlock()
	for _, flow := range report.Flows {
		_ = s.log(ctx, project.ID, flow.ID+".yaml → "+string(flow.Result)+". "+flow.Note)
	}
	return s.Get(ctx, userID, id)
}

func (s *Service) ActivateSnap(id string) activate.Snapshot {
	if s.Activator == nil {
		return activate.Snapshot{Status: activate.StatusIdle, Note: "Yerel aktif bağlı değil."}
	}
	return s.Activator.Snapshot(id)
}

func (s *Service) runLLM(ctx context.Context, project store.Project) error {
	if s.LLM == nil {
		return s.log(ctx, project.ID, "LLM bağlı değil — yalnızca stub dosyalar.")
	}
	if err := s.LLM.SetMcpRoot(ctx, project.RootPath); err != nil {
		return err
	}
	readme, err := s.LLM.ReadProjectFile(ctx, "README.md")
	if err != nil {
		readme = ""
	}
	prompt := "Amaç: codegen. Yalnızca proje kökündeki dosyalar.\nYığın: " + string(project.Stack) + "\nAd: " + project.Name + "\nBrif: " + project.Brief + "\nClean Architecture katmanlarını koru. Uygulamayı seçilen dilde yaz; HTML site değil.\nREADME:\n" + readme
	out, err := s.LLM.Complete(ctx, llm.CompleteInput{
		UserID:     project.UserID,
		ProjectID:  project.ID,
		Purpose:    "codegen",
		LegalBasis: "contract",
		Prompt:     prompt,
	})
	if err != nil {
		return err
	}
	planPath := filepath.Join(project.RootPath, "llm", "plan.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		return err
	}
	body := "# LLM " + string(out.Slot) + " " + out.Version.Name + "\n\nkanal: " + out.Channel + "\nredaksiyon giriş/çıkış: " + fmt.Sprintf("%d/%d", out.InputN, out.OutputN) + "\n\n" + out.Text + "\n"
	if err := os.WriteFile(planPath, []byte(body), 0o644); err != nil {
		return err
	}
	return s.log(ctx, project.ID, "GDPR → işçi "+string(out.Slot)+" "+out.Version.Name+" ("+out.Channel+") redact="+fmt.Sprintf("%d/%d", out.InputN, out.OutputN)+". Plan: llm/plan.md")
}

func (s *Service) runOpenCode(ctx context.Context, project store.Project) error {
	if s.OpenCode == nil {
		return s.log(ctx, project.ID, "OpenCode bağlı değil — iskelet kaldı.")
	}
	if err := s.log(ctx, project.ID, "OpenCode yazılıyor (GDPR’li brif, proje kökü)."); err != nil {
		return err
	}
	plan := ""
	if s.LLM != nil {
		text, err := s.LLM.ReadProjectFile(ctx, "llm/plan.md")
		if err == nil {
			plan = text
		}
	}
	label, err := stackLabel(project.Stack)
	if err != nil {
		return err
	}
	briefSafe, _ := gdpr.Redact(project.Brief)
	nameSafe, _ := gdpr.Redact(project.Name)
	rule, err := stackSourceRule(project.Stack)
	if err != nil {
		return err
	}
	beLabel, err := backendLabel(project.Backend)
	if err != nil {
		beLabel = "Yerel API"
	}
	prompt := strings.Join([]string{
		"Cherry müşteri uygulamasını bu dizinde yaz. Kök dışına çıkma.",
		"Yığın: " + label,
		rule,
		"Backend hedefi: " + beLabel + ". Token’ı koda gömme.",
		"Ad: " + nameSafe,
		"Brif: " + briefSafe,
		"Asıl klasörler: frontend/ (seçilen dil, Clean Architecture), backend/, maestro/.",
		"Mevcut domain/data/presentation katmanlarını koru ve genişlet. Tek dosyaya yığma.",
		"preview/ HTML makettir; uygulamayı HTML ile yazma. Teslim zip HTML site olmasın.",
		"Barındırma yok. Teslim dosya. Maestro YAML yaz; emülatör yoksa test çalıştırma.",
		"Cherry platform GraphQL’ine dokunma.",
		"Plan:",
		plan,
	}, "\n")
	req := opencode.Request{
		Dir:      project.RootPath,
		Prompt:   prompt,
		Title:    "cherry-" + project.ID,
		Continue: false,
	}
	if s.LLM != nil {
		req.BaseURL, req.APIKey = s.LLM.OpenCodeEndpoint()
	}
	res, err := s.OpenCode.Run(ctx, req)
	if writeErr := opencode.WriteLog(project.RootPath, opencode.LogBody(res)); writeErr != nil {
		return writeErr
	}
	if err != nil {
		return err
	}
	return s.recordOpenCode(ctx, project.ID, res)
}

func (s *Service) SendMessage(ctx context.Context, userID, id, body string) (store.Project, error) {
	project, err := s.Get(ctx, userID, id)
	if err != nil {
		return store.Project{}, err
	}
	body = strings.TrimSpace(body)
	if len(body) < 1 {
		return store.Project{}, store.ErrValidation
	}
	switch project.Status {
	case store.StatusQueued, store.StatusWriting:
		return store.Project{}, store.ErrBusy
	case store.StatusTesting, store.StatusReady, store.StatusFailed:
		// follow-up chat allowed
	default:
		return store.Project{}, fmt.Errorf("unhandled status: %s", project.Status)
	}
	safe, _ := gdpr.Redact(body)
	if err := s.logRole(ctx, project.ID, body, store.RoleUser); err != nil {
		return store.Project{}, err
	}
	if s.LLM != nil {
		if err := s.LLM.SetMcpRoot(ctx, project.RootPath); err != nil {
			return store.Project{}, err
		}
		_, _ = s.LLM.Complete(ctx, llm.CompleteInput{
			UserID:     project.UserID,
			ProjectID:  project.ID,
			Purpose:    "chat",
			LegalBasis: "contract",
			Prompt:     "Amaç: sohbet düzeltmesi. Yalnızca proje kökü.\nMesaj: " + safe,
		})
	}
	if s.OpenCode == nil {
		if err := s.logRole(ctx, project.ID, "Ajan bu stüdyoda bağlı değil. OpenCode penceresi açılmaz; iskelet duruyor.", store.RoleAgent); err != nil {
			return store.Project{}, err
		}
		return s.Get(ctx, userID, id)
	}
	if err := s.setStatus(ctx, project.ID, store.StatusWriting); err != nil {
		return store.Project{}, err
	}
	rule, err := stackSourceRule(project.Stack)
	if err != nil {
		return store.Project{}, err
	}
	prompt := strings.Join([]string{
		"Cherry sohbeti. OpenCode TUI açma. Yalnızca bu dizin.",
		"Kullanıcı: " + safe,
		rule,
		"Gerekirse frontend/, backend/, maestro/ güncelle. Clean Architecture katmanlarını bozma. HTML site yazma.",
	}, "\n")
	ocReq := opencode.Request{
		Dir:      project.RootPath,
		Prompt:   prompt,
		Title:    "cherry-" + project.ID,
		Continue: true,
	}
	if s.LLM != nil {
		ocReq.BaseURL, ocReq.APIKey = s.LLM.OpenCodeEndpoint()
	}
	res, err := s.OpenCode.Run(ctx, ocReq)
	if writeErr := opencode.WriteLog(project.RootPath, opencode.LogBody(res)); writeErr != nil {
		_ = s.setStatus(ctx, project.ID, store.StatusFailed)
		return store.Project{}, writeErr
	}
	if err != nil {
		_ = s.setStatus(ctx, project.ID, store.StatusFailed)
		return store.Project{}, err
	}
	if recErr := s.recordOpenCode(ctx, project.ID, res); recErr != nil {
		return store.Project{}, recErr
	}
	next := store.StatusReady
	if res.Status == opencode.StatusFailed {
		next = store.StatusFailed
	}
	if err := s.setStatus(ctx, project.ID, next); err != nil {
		return store.Project{}, err
	}
	return s.Get(ctx, userID, id)
}

func (s *Service) recordOpenCode(ctx context.Context, projectID string, res opencode.Result) error {
	switch res.Status {
	case opencode.StatusRan:
		reply := clipChat(res.Output, 800)
		if reply == "" {
			reply = "Dosyaları güncelledim. OpenCode bu pencerede çalıştı, ayrı uygulama açılmadı."
		}
		if err := s.logRole(ctx, projectID, reply, store.RoleAgent); err != nil {
			return err
		}
		return s.log(ctx, projectID, "OpenCode "+res.Status.Label()+" ("+res.Bin+"). Günlük: llm/opencode.log")
	case opencode.StatusMissing:
		if err := s.logRole(ctx, projectID, "Ajan bu bilgisayarda henüz kurulu değil. İskelet duruyor — sahte yazım yok. OpenCode penceresi açılmaz.", store.RoleAgent); err != nil {
			return err
		}
		return s.log(ctx, projectID, "OpenCode CLI yok — iskelet kaldı, sahte yazım yok.")
	case opencode.StatusFailed:
		if err := s.logRole(ctx, projectID, "Yazamadı: "+res.Err+". İskelet duruyor.", store.RoleAgent); err != nil {
			return err
		}
		return s.log(ctx, projectID, "OpenCode hata: "+res.Err+" — iskelet duruyor. llm/opencode.log")
	default:
		return s.log(ctx, projectID, "OpenCode durum: "+string(res.Status))
	}
}

func clipChat(s string, n int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func (s *Service) setStatus(ctx context.Context, id string, status store.ProjectStatus) error {
	project, err := s.Store.GetProject(ctx, id)
	if err != nil {
		return err
	}
	project.Status = status
	project.UpdatedAt = time.Now().UTC()
	return s.Store.UpdateProject(ctx, *project)
}

func (s *Service) fail(ctx context.Context, id string, cause error) error {
	_ = s.setStatus(ctx, id, store.StatusFailed)
	return s.log(ctx, id, "İş durdu: "+cause.Error())
}

func (s *Service) log(ctx context.Context, projectID, message string) error {
	return s.logRole(ctx, projectID, message, store.RoleSystem)
}

func (s *Service) logRole(ctx context.Context, projectID, message string, role store.ChatRole) error {
	if role == "" {
		role = store.RoleSystem
	}
	return s.Store.AppendLog(ctx, store.JobLog{
		ProjectID: projectID,
		At:        time.Now().UTC(),
		Message:   message,
		Role:      role,
	})
}

func (s *Service) pause() {
	if s.StepDelay <= 0 {
		return
	}
	time.Sleep(s.StepDelay)
}

func initCustomerRepo(root string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return nil
	}
	if _, err := os.Stat(filepath.Join(root, ".git")); err == nil {
		return nil
	}
	cmd := exec.Command("git", "init")
	cmd.Dir = root
	return cmd.Run()
}

func fileKind(rel string) string {
	switch {
	case rel == "README.md":
		return "readme"
	case strings.HasPrefix(rel, "frontend/"):
		return "frontend"
	case strings.HasPrefix(rel, "backend/"):
		return "backend"
	case strings.HasPrefix(rel, "maestro/"):
		return "maestro"
	case strings.HasPrefix(rel, "llm/"):
		return "llm"
	case rel == "opencode.json" || rel == "AGENTS.md":
		return "opencode"
	case strings.HasPrefix(rel, "preview/"):
		return "preview"
	default:
		return "other"
	}
}

func (s *Service) Files(ctx context.Context, userID, id string) ([]DiskFile, error) {
	project, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	paths, err := listFiles(project.RootPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []DiskFile{}, nil
		}
		return nil, err
	}
	out := make([]DiskFile, 0, len(paths))
	for _, rel := range paths {
		if skipDeliveryRel(rel) {
			continue
		}
		out = append(out, DiskFile{Path: rel, Kind: fileKind(rel)})
	}
	sort.Slice(out, func(i, j int) bool {
		ri, rj := rankDelivery(out[i].Path), rankDelivery(out[j].Path)
		if ri != rj {
			return ri < rj
		}
		return out[i].Path < out[j].Path
	})
	return out, nil
}

func (s *Service) Maestro(ctx context.Context, userID, id string) (MaestroStudio, error) {
	project, err := s.Get(ctx, userID, id)
	if err != nil {
		return MaestroStudio{}, err
	}
	studio := MaestroStudio{
		DeviceStatus: "none",
		Ready:        project.Status == store.StatusTesting || project.Status == store.StatusReady,
	}
	previewDir := filepath.Join(project.RootPath, "preview")
	entries, _ := os.ReadDir(previewDir)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(previewDir, entry.Name()))
		if err != nil {
			return MaestroStudio{}, err
		}
		idName := strings.TrimSuffix(entry.Name(), ".html")
		title := "Ekran"
		if idName == "login" {
			title = "Giriş"
		}
		if idName == "home" {
			title = "Ana ekran"
		}
		studio.Screens = append(studio.Screens, DesignScreen{
			ID:   idName,
			Name: title,
			HTML: string(body),
		})
	}
	maestroDir := filepath.Join(project.RootPath, "maestro")
	flows, _ := os.ReadDir(maestroDir)
	for _, entry := range flows {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(maestroDir, entry.Name()))
		if err != nil {
			return MaestroStudio{}, err
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		studio.Flows = append(studio.Flows, Flow{
			ID:     name,
			Name:   name + ".yaml",
			YAML:   string(body),
			Result: store.MaestroSkipped,
			Note:   "Henüz koşulmadı. Cihaz yoksa SKIPPED — geçti sayılmaz.",
		})
	}
	s.mu.Lock()
	report, ok := s.reports[project.ID]
	s.mu.Unlock()
	if ok {
		studio.DeviceStatus = report.DeviceStatus
		byID := map[string]maestro.FlowResult{}
		for _, item := range report.Flows {
			byID[item.ID] = item
		}
		for i, flow := range studio.Flows {
			if hit, exists := byID[flow.ID]; exists {
				studio.Flows[i].Result = hit.Result
				studio.Flows[i].Note = hit.Note
			}
		}
	}
	return studio, nil
}

type DiskFile struct {
	Path string
	Kind string
}

type MaestroStudio struct {
	DeviceStatus string
	Ready        bool
	Screens      []DesignScreen
	Flows        []Flow
}

type DesignScreen struct {
	ID   string
	Name string
	HTML string
}

type Flow struct {
	ID     string
	Name   string
	YAML   string
	Result store.MaestroResult
	Note   string
}

func (s *Service) RunSync(ctx context.Context, id string) error {
	project, err := s.Store.GetProject(ctx, id)
	if err != nil {
		return fmt.Errorf("project: %w", err)
	}
	return s.pipeline(ctx, *project)
}
