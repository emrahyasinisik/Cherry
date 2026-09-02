package activate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Status string

const (
	StatusIdle     Status = "IDLE"
	StatusStarting Status = "STARTING"
	StatusRunning  Status = "RUNNING"
	StatusStopping Status = "STOPPING"
	StatusFailed   Status = "FAILED"
)

func (s Status) Label() string {
	switch s {
	case StatusIdle:
		return "kapalı"
	case StatusStarting:
		return "kalkıyor"
	case StatusRunning:
		return "çalışıyor"
	case StatusStopping:
		return "durduruluyor"
	case StatusFailed:
		return "hata"
	default:
		return string(s)
	}
}

type Snapshot struct {
	Status Status
	URL    string
	Port   int
	PID    int
	Note   string
}

type run struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	snap   Snapshot
}

type Service struct {
	mu    sync.Mutex
	procs map[string]*run
}

func New() *Service {
	return &Service{procs: map[string]*run{}}
}

func (s *Service) Snapshot(id string) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.procs[id]
	if !ok {
		return Snapshot{Status: StatusIdle, Note: "Yerel API kapalı."}
	}
	return item.snap
}

func (s *Service) Start(ctx context.Context, id, projectRoot string) (Snapshot, error) {
	s.mu.Lock()
	if existing, ok := s.procs[id]; ok && existing.snap.Status == StatusRunning {
		snap := existing.snap
		s.mu.Unlock()
		return snap, nil
	}
	s.mu.Unlock()

	backend := filepath.Join(projectRoot, "backend")
	if _, err := os.Stat(filepath.Join(backend, "main.go")); err != nil {
		return Snapshot{Status: StatusFailed, Note: "backend/main.go yok"}, err
	}
	if _, err := exec.LookPath("go"); err != nil {
		return Snapshot{Status: StatusFailed, Note: "go yok — müşteri API’si ayağa kalkamadı"}, err
	}
	port, err := freePort()
	if err != nil {
		return Snapshot{Status: StatusFailed, Note: err.Error()}, err
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	runCtx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(runCtx, "go", "run", "main.go")
	cmd.Dir = backend
	cmd.Env = append(os.Environ(), "ICERDE_CUSTOMER_ADDR="+addr)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	item := &run{
		cmd:    cmd,
		cancel: cancel,
		snap: Snapshot{
			Status: StatusStarting,
			URL:    "http://" + addr,
			Port:   port,
			Note:   "Kalkıyor…",
		},
	}
	s.mu.Lock()
	s.procs[id] = item
	s.mu.Unlock()
	if err := cmd.Start(); err != nil {
		cancel()
		item.snap.Status = StatusFailed
		item.snap.Note = err.Error()
		return item.snap, err
	}
	item.snap.PID = cmd.Process.Pid
	health := "http://" + addr + "/health"
	if err := waitHealth(ctx, health, 12*time.Second); err != nil {
		_ = s.Stop(id)
		fail := Snapshot{Status: StatusFailed, Port: port, URL: "http://" + addr, Note: "localhost health yok: " + err.Error()}
		s.mu.Lock()
		s.procs[id] = &run{snap: fail}
		s.mu.Unlock()
		return fail, err
	}
	s.mu.Lock()
	item.snap.Status = StatusRunning
	item.snap.Note = "Müşteri API " + item.snap.URL + " — barındırma yok."
	snap := item.snap
	s.mu.Unlock()
	go func() {
		_ = cmd.Wait()
		s.mu.Lock()
		if current, ok := s.procs[id]; ok && current.cmd == cmd && current.snap.Status == StatusRunning {
			current.snap.Status = StatusIdle
			current.snap.Note = "Süreç bitti."
			current.snap.PID = 0
		}
		s.mu.Unlock()
	}()
	return snap, nil
}

func (s *Service) Stop(id string) Snapshot {
	s.mu.Lock()
	item, ok := s.procs[id]
	if !ok {
		s.mu.Unlock()
		return Snapshot{Status: StatusIdle, Note: "Yerel API kapalı."}
	}
	item.snap.Status = StatusStopping
	cmd := item.cmd
	cancel := item.cancel
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if cmd != nil && cmd.Process != nil {
		killProcess(cmd)
		_ = cmd.Wait()
	}
	idle := Snapshot{Status: StatusIdle, Note: "Yerel API durdu."}
	s.mu.Lock()
	s.procs[id] = &run{snap: idle}
	s.mu.Unlock()
	return idle
}

func freePort() (int, error) {
	for port := 47000; port <= 47999; port++ {
		ln, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
		if err != nil {
			continue
		}
		addr := ln.Addr().(*net.TCPAddr)
		_ = ln.Close()
		return addr.Port, nil
	}
	return 0, errors.New("47000–47999 arası boş port yok")
}

func waitHealth(ctx context.Context, url string, d time.Duration) error {
	deadline := time.Now().Add(d)
	client := &http.Client{Timeout: 400 * time.Millisecond}
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		res, err := client.Do(req)
		if err == nil {
			_ = res.Body.Close()
			if res.StatusCode < 500 {
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return errors.New("timeout")
}

func killProcess(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	if runtime.GOOS == "windows" {
		_ = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
}
