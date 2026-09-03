package colabbridge

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cherry/api/internal/store"
)

var ErrBridge = errors.New("colab köprüsü")

func wrap(msg string) error {
	return fmt.Errorf("%w: %s", ErrBridge, msg)
}

type Status string

const (
	StatusIdle     Status = "IDLE"
	StatusStarting Status = "STARTING"
	StatusRunning  Status = "RUNNING"
	StatusStopping Status = "STOPPING"
	StatusFailed   Status = "FAILED"
)

const maxCheckpoint = 256 << 20

type Snapshot struct {
	Status      Status
	PublicURL   string
	LocalURL    string
	Token       string
	TokenHint   string
	Cloudflared string
	StartedAt   string
	Note        string
	UserID      string
}

type Packer func(ctx context.Context, userID string) (jsonBody, jsonlBody string, err error)

type Registrar func(ctx context.Context, userID, slot, name, note, checkpointRef string) error

type Service struct {
	Addr     string
	CheckDir string
	Pack     Packer
	Register Registrar
	NewTunnel func(bin string) Tunnel

	mu      sync.Mutex
	snap    Snapshot
	token   string
	srv     *http.Server
	ln      net.Listener
	tunnel  Tunnel
	started time.Time
}

func New(addr, checkDir string) *Service {
	if strings.TrimSpace(addr) == "" {
		addr = "127.0.0.1:43149"
	}
	return &Service{
		Addr:     addr,
		CheckDir: checkDir,
		snap: Snapshot{
			Status:      StatusIdle,
			Cloudflared: cloudflaredSource(),
			Note:        idleNote(),
		},
	}
}

func idleNote() string {
	return "Tünel kapalı. Colab stüdyoya ulaşamaz; paketi elle yükle."
}

func cloudflaredSource() string {
	hit, err := LookCloudflared()
	if err != nil {
		return "missing"
	}
	return hit.Source
}

func (s *Service) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.snap
	out.Cloudflared = cloudflaredSource()
	out.Token = s.token
	out.TokenHint = tokenHint(s.token)
	return out
}

func (s *Service) Start(userID string) Snapshot {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		s.mu.Lock()
		s.failLocked(wrap("oturum gerekli"))
		out := s.snap
		s.mu.Unlock()
		return out
	}
	s.mu.Lock()
	if s.snap.Status == StatusRunning && s.snap.UserID == userID && s.snap.PublicURL != "" {
		out := s.snap
		out.Token = s.token
		out.TokenHint = tokenHint(s.token)
		s.mu.Unlock()
		return out
	}
	if s.snap.Status == StatusStarting {
		out := s.snap
		out.Token = s.token
		out.TokenHint = tokenHint(s.token)
		s.mu.Unlock()
		return out
	}
	s.mu.Unlock()
	_ = s.Stop()

	hit, err := LookCloudflared()
	if err != nil {
		s.mu.Lock()
		s.token = ""
		s.snap = Snapshot{
			Status:      StatusFailed,
			Cloudflared: "missing",
			Note:        "cloudflared yok — vendor/bin veya PATH. Sahte tünel yok.",
		}
		out := s.snap
		s.mu.Unlock()
		return out
	}

	token := store.NewID() + store.NewID()
	s.mu.Lock()
	s.token = token
	s.snap = Snapshot{
		Status:      StatusStarting,
		TokenHint:   tokenHint(token),
		Cloudflared: hit.Source,
		UserID:      userID,
		Note:        "Tünel açılıyor…",
	}
	s.mu.Unlock()

	go s.open(userID, token, hit.Path)
	return s.Snapshot()
}

func (s *Service) open(userID, token, bin string) {
	localURL, err := s.ensureListener()
	if err != nil {
		s.mu.Lock()
		s.failLocked(err)
		s.mu.Unlock()
		return
	}
	factory := s.NewTunnel
	if factory == nil {
		factory = func(path string) Tunnel {
			return &Cloudflared{Bin: path}
		}
	}
	tun := factory(bin)
	if tun == nil {
		s.mu.Lock()
		s.failLocked(wrap("tünel başlatıcı yok"))
		s.mu.Unlock()
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	publicURL, err := tun.Start(ctx, localURL)
	if err != nil {
		tun.Stop()
		_ = s.closeListener()
		s.mu.Lock()
		s.failLocked(err)
		s.mu.Unlock()
		return
	}
	if ParsePublicURL(publicURL) == "" && !strings.HasPrefix(publicURL, "http://127.0.0.1") {
		tun.Stop()
		_ = s.closeListener()
		s.mu.Lock()
		s.failLocked(wrap("geçersiz tünel URL — sahte adres yok"))
		s.mu.Unlock()
		return
	}

	s.mu.Lock()
	if s.token != token {
		s.mu.Unlock()
		tun.Stop()
		return
	}
	s.tunnel = tun
	s.started = time.Now().UTC()
	s.snap = Snapshot{
		Status:      StatusRunning,
		PublicURL:   strings.TrimRight(publicURL, "/"),
		LocalURL:    localURL,
		TokenHint:   tokenHint(token),
		Cloudflared: cloudflaredSource(),
		StartedAt:   s.started.Format(time.RFC3339),
		Note:        "Colab bu URL + token ile paketi çeker ve adapter gönderir. GraphQL public değil.",
		UserID:      userID,
	}
	s.mu.Unlock()
}

func (s *Service) Stop() Snapshot {
	s.mu.Lock()
	tun := s.tunnel
	s.tunnel = nil
	s.token = ""
	if s.snap.Status == StatusRunning || s.snap.Status == StatusStarting {
		s.snap.Status = StatusStopping
		s.snap.Note = "Tünel kapanıyor…"
	}
	s.mu.Unlock()
	if tun != nil {
		tun.Stop()
	}
	_ = s.closeListener()
	s.mu.Lock()
	s.snap = Snapshot{
		Status:      StatusIdle,
		Cloudflared: cloudflaredSource(),
		Note:        idleNote(),
	}
	out := s.snap
	s.mu.Unlock()
	return out
}

func (s *Service) failLocked(err error) {
	note := "Tünel açılamadı."
	if err != nil {
		note = strings.TrimPrefix(err.Error(), ErrBridge.Error()+": ")
	}
	s.token = ""
	s.tunnel = nil
	s.snap = Snapshot{
		Status:      StatusFailed,
		Cloudflared: cloudflaredSource(),
		Note:        note,
	}
}

func (s *Service) ensureListener() (string, error) {
	s.mu.Lock()
	if s.srv != nil && s.ln != nil {
		url := "http://" + s.ln.Addr().String()
		s.mu.Unlock()
		return url, nil
	}
	s.mu.Unlock()

	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return "", wrap("köprü dinlenemedi: " + err.Error())
	}
	host, _, splitErr := net.SplitHostPort(ln.Addr().String())
	if splitErr == nil && host != "127.0.0.1" && host != "::1" && host != "localhost" {
		_ = ln.Close()
		return "", wrap("köprü yalnızca 127.0.0.1 — GraphQL tünellenmez")
	}
	srv := &http.Server{
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	s.mu.Lock()
	s.ln = ln
	s.srv = srv
	s.mu.Unlock()
	go func() {
		_ = srv.Serve(ln)
	}()
	return "http://" + ln.Addr().String(), nil
}

func (s *Service) closeListener() error {
	s.mu.Lock()
	srv := s.srv
	s.srv = nil
	s.ln = nil
	s.mu.Unlock()
	if srv == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/pack.json", s.handlePackJSON)
	mux.HandleFunc("/pack.jsonl", s.handlePackJSONL)
	mux.HandleFunc("/checkpoint", s.handleCheckpoint)
	return mux
}

func (s *Service) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	snap := s.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"service":  "cherry-colab-bridge",
		"status":   snap.Status,
		"tunnel":   snap.Status == StatusRunning && snap.PublicURL != "",
		"inference": false,
	})
}

func (s *Service) handlePackJSON(w http.ResponseWriter, r *http.Request) {
	s.servePack(w, r, false)
}

func (s *Service) handlePackJSONL(w http.ResponseWriter, r *http.Request) {
	s.servePack(w, r, true)
}

func (s *Service) servePack(w http.ResponseWriter, r *http.Request, jsonl bool) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := s.authorize(r)
	if err != nil {
		http.Error(w, "köprü token gerekli", http.StatusUnauthorized)
		return
	}
	if s.Pack == nil {
		http.Error(w, "paket yok", http.StatusServiceUnavailable)
		return
	}
	jsonBody, jsonlBody, err := s.Pack(r.Context(), userID)
	if err != nil {
		http.Error(w, "paket hazırlanamadı", http.StatusInternalServerError)
		return
	}
	if jsonl {
		w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="cherry-training-pack.jsonl"`)
		_, _ = io.WriteString(w, jsonlBody)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="cherry-training-pack.json"`)
	_, _ = io.WriteString(w, jsonBody)
}

func (s *Service) handleCheckpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := s.authorize(r)
	if err != nil {
		http.Error(w, "köprü token gerekli", http.StatusUnauthorized)
		return
	}
	if s.Register == nil {
		http.Error(w, "kayıt yok", http.StatusServiceUnavailable)
		return
	}
	if err := r.ParseMultipartForm(maxCheckpoint); err != nil {
		http.Error(w, "zip gerekli", http.StatusBadRequest)
		return
	}
	slot := strings.TrimSpace(r.FormValue("slot"))
	name := strings.TrimSpace(r.FormValue("name"))
	note := strings.TrimSpace(r.FormValue("note"))
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file alanı zip olmalı", http.StatusBadRequest)
		return
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, maxCheckpoint+1))
	if err != nil || len(raw) == 0 {
		http.Error(w, "zip okunamadı", http.StatusBadRequest)
		return
	}
	if len(raw) > maxCheckpoint {
		http.Error(w, "zip çok büyük", http.StatusRequestEntityTooLarge)
		return
	}
	if len(raw) < 4 || string(raw[:2]) != "PK" {
		http.Error(w, "dosya zip değil", http.StatusBadRequest)
		return
	}
	filename := filepath.Base(header.Filename)
	if filename == "" || filename == "." || !strings.HasSuffix(strings.ToLower(filename), ".zip") {
		filename = "cherry_adapter.zip"
	}
	if err := os.MkdirAll(s.CheckDir, 0o755); err != nil {
		http.Error(w, "kayıt klasörü yok", http.StatusInternalServerError)
		return
	}
	safe := fmt.Sprintf("%s-%s-%s", time.Now().UTC().Format("20060102T150405"), slot, filename)
	dest := filepath.Join(s.CheckDir, safe)
	if err := os.WriteFile(dest, raw, 0o644); err != nil {
		http.Error(w, "zip yazılamadı", http.StatusInternalServerError)
		return
	}
	if name == "" {
		name = "v-colab"
	}
	if err := s.Register(r.Context(), userID, slot, name, note, safe); err != nil {
		http.Error(w, "sürüm kaydedilemedi", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":            true,
		"checkpointRef": safe,
		"slot":          slot,
		"name":          name,
		"inference":     false,
	})
}

func (s *Service) authorize(r *http.Request) (string, error) {
	raw := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if raw == "" {
		raw = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.token == "" || s.snap.Status != StatusRunning || s.snap.UserID == "" {
		return "", wrap("tünel kapalı")
	}
	if subtle.ConstantTimeCompare([]byte(raw), []byte(s.token)) != 1 {
		return "", wrap("token")
	}
	return s.snap.UserID, nil
}

func tokenHint(token string) string {
	if len(token) < 4 {
		return ""
	}
	return token[len(token)-4:]
}
