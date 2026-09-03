package colabbridge

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/cherry/api/internal/sidecar"
)

// Tunnel starts a public HTTPS front for a loopback URL and later stops it.
type Tunnel interface {
	Start(ctx context.Context, localURL string) (publicURL string, err error)
	Stop()
}

// LookCloudflared returns the sidecar hit or a missing error. Never invents a URL.
func LookCloudflared() (sidecar.Hit, error) {
	return sidecar.Look("cloudflared")
}

// Cloudflared runs `cloudflared tunnel --url` and parses the trycloudflare URL from logs.
type Cloudflared struct {
	Bin     string
	Timeout time.Duration

	mu     sync.Mutex
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

func (c *Cloudflared) Start(ctx context.Context, localURL string) (string, error) {
	if c.Bin == "" {
		return "", wrap("cloudflared yok — vendor/bin veya PATH. Sahte tünel yok.")
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 40 * time.Second
	}
	runCtx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(runCtx, c.Bin, "tunnel", "--url", localURL, "--no-autoupdate", "--metrics", "127.0.0.1:0")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return "", wrap(err.Error())
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return "", wrap(err.Error())
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return "", wrap("cloudflared başlamadı: " + err.Error())
	}
	c.mu.Lock()
	c.cmd = cmd
	c.cancel = cancel
	c.mu.Unlock()

	found := make(chan string, 1)
	go readTunnelLogs(io.MultiReader(stdout, stderr), found)
	waitErr := make(chan error, 1)
	go func() {
		waitErr <- cmd.Wait()
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		c.Stop()
		return "", wrap("tünel iptal")
	case err := <-waitErr:
		if err == nil {
			err = fmt.Errorf("süreç bitti")
		}
		c.Stop()
		return "", wrap("cloudflared çıktı: " + err.Error())
	case url := <-found:
		if url == "" {
			c.Stop()
			return "", wrap("trycloudflare URL yok")
		}
		return url, nil
	case <-timer.C:
		c.Stop()
		return "", wrap("tünel URL zaman aşımı — cloudflared log’unda trycloudflare yok")
	}
}

func (c *Cloudflared) Stop() {
	c.mu.Lock()
	cmd := c.cmd
	cancel := c.cancel
	c.cmd = nil
	c.cancel = nil
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if cmd == nil || cmd.Process == nil {
		return
	}
	if runtime.GOOS == "windows" {
		_ = cmd.Process.Kill()
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		_ = cmd.Process.Kill()
	}
}

func readTunnelLogs(r io.Reader, found chan<- string) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var buf string
	for scanner.Scan() {
		line := scanner.Text()
		buf += line + "\n"
		if url := ParsePublicURL(buf); url != "" {
			select {
			case found <- url:
			default:
			}
			return
		}
		if len(buf) > 64*1024 {
			buf = buf[len(buf)/2:]
		}
	}
}

// StaticTunnel is a test double. Production must use Cloudflared.
type StaticTunnel struct {
	URL string
	Err error
}

func (s StaticTunnel) Start(context.Context, string) (string, error) {
	return s.URL, s.Err
}

func (s StaticTunnel) Stop() {}
