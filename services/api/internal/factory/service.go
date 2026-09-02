package factory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/icerde/api/internal/store"
)

type Service struct {
	Store     store.Store
	Root      string
	StepDelay time.Duration
	AutoRun   bool
}

func New(st store.Store, root string) *Service {
	if strings.TrimSpace(root) == "" {
		root = filepath.Join("var", "projects")
	}
	return &Service{Store: st, Root: root, StepDelay: 380 * time.Millisecond, AutoRun: true}
}

func (s *Service) Create(ctx context.Context, userID, name, brief, stackRaw string) (store.Project, error) {
	name = strings.TrimSpace(name)
	brief = strings.TrimSpace(brief)
	if name == "" || len(brief) < 8 {
		return store.Project{}, store.ErrValidation
	}
	stack, err := parseStack(stackRaw)
	if err != nil {
		return store.Project{}, err
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
	path := filepath.Join(project.RootPath, "icerde.zip")
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
	if err := s.log(ctx, project.ID, "Dosya ağacı hazır: "+project.RootPath); err != nil {
		return err
	}
	s.pause()
	if err := s.setStatus(ctx, project.ID, store.StatusTesting); err != nil {
		return err
	}
	if err := s.log(ctx, project.ID, "Test aşaması. Maestro ekranı açılabilir — cihaz yok, akışlar SKIPPED (sahte geçiş yok)."); err != nil {
		return err
	}
	s.pause()
	zipPath := filepath.Join(project.RootPath, "icerde.zip")
	if err := zipProject(project.RootPath, zipPath); err != nil {
		return err
	}
	if err := s.setStatus(ctx, project.ID, store.StatusReady); err != nil {
		return err
	}
	return s.log(ctx, project.ID, "Hazır. Zip ve Maestro YAML müşteri dosyalarında. OpenCode henüz bağlı değil (stub).")
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
	return s.Store.AppendLog(ctx, store.JobLog{
		ProjectID: projectID,
		At:        time.Now().UTC(),
		Message:   message,
	})
}

func (s *Service) pause() {
	if s.StepDelay <= 0 {
		return
	}
	time.Sleep(s.StepDelay)
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
		if rel == "icerde.zip" {
			continue
		}
		out = append(out, DiskFile{Path: rel, Kind: fileKind(rel)})
	}
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
			Note:   "Emülatör yok. SKIPPED — geçti sayılmaz. Dilim 6’da maestro mcp.",
		})
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
