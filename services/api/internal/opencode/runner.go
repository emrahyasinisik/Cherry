package opencode

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/icerde/api/internal/gdpr"
)

type Status string

const (
	StatusRan     Status = "ran"
	StatusMissing Status = "missing"
	StatusFailed  Status = "failed"
)

func (s Status) Label() string {
	switch s {
	case StatusRan:
		return "yazdı"
	case StatusMissing:
		return "CLI yok"
	case StatusFailed:
		return "hata"
	default:
		return string(s)
	}
}

type Request struct {
	Dir    string
	Prompt string
	Title  string
	Model  string
}

type Result struct {
	Status  Status
	Bin     string
	Version string
	Output  string
	Err     string
}

type Runner interface {
	Probe() (bin string, version string, ok bool)
	Run(ctx context.Context, req Request) (Result, error)
}

type CLI struct {
	Bin     string
	Timeout time.Duration
	Require bool
}

func NewCLI() *CLI {
	timeout := 8 * time.Minute
	if raw := strings.TrimSpace(os.Getenv("ICERDE_OPENCODE_TIMEOUT_SEC")); raw != "" {
		if n, err := time.ParseDuration(raw + "s"); err == nil && n > 0 {
			timeout = n
		}
	}
	return &CLI{
		Bin:     strings.TrimSpace(os.Getenv("ICERDE_OPENCODE_BIN")),
		Timeout: timeout,
		Require: envTruthy("ICERDE_OPENCODE_REQUIRE"),
	}
}

func (c *CLI) Probe() (string, string, bool) {
	bin, err := c.lookup()
	if err != nil {
		return "", "", false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "--version")
	out, _ := cmd.CombinedOutput()
	return bin, strings.TrimSpace(string(out)), true
}

func (c *CLI) lookup() (string, error) {
	if c.Bin != "" {
		return c.Bin, nil
	}
	return exec.LookPath("opencode")
}

func (c *CLI) Run(ctx context.Context, req Request) (Result, error) {
	bin, err := c.lookup()
	if err != nil {
		res := Result{Status: StatusMissing, Err: "opencode CLI PATH'te yok"}
		if c.Require {
			return res, fmt.Errorf("opencode required: %w", err)
		}
		return res, nil
	}
	if err := WriteConfig(req.Dir); err != nil {
		return Result{Status: StatusFailed, Bin: bin, Err: err.Error()}, err
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 8 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := []string{"run", "--dir", req.Dir, "--auto", "--title", titleOr(req.Title, "icerde")}
	if model := strings.TrimSpace(firstNonEmpty(req.Model, os.Getenv("ICERDE_OPENCODE_MODEL"), os.Getenv("ICERDE_LLM_MODEL"))); model != "" {
		args = append(args, "--model", model)
	}
	cmd := exec.CommandContext(runCtx, bin, args...)
	cmd.Dir = req.Dir
	cmd.Stdin = strings.NewReader(req.Prompt)
	cmd.Env = withLLMKey(os.Environ())
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	runErr := cmd.Run()
	raw := buf.String()
	safe, _ := gdpr.Scan(raw)
	res := Result{Status: StatusRan, Bin: bin, Output: safe, Version: versionOf(bin)}
	if runErr != nil {
		res.Status = StatusFailed
		res.Err = runErr.Error()
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			res.Err = "opencode zaman aşımı"
		}
		if c.Require {
			return res, fmt.Errorf("opencode: %s", res.Err)
		}
		return res, nil
	}
	return res, nil
}

func versionOf(bin string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, bin, "--version").CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func titleOr(title, fallback string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return fallback
	}
	if len(title) > 60 {
		return title[:60]
	}
	return title
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func withLLMKey(env []string) []string {
	key := strings.TrimSpace(os.Getenv("ICERDE_LLM_API_KEY"))
	if key == "" {
		return env
	}
	if os.Getenv("OPENAI_API_KEY") == "" {
		env = append(env, "OPENAI_API_KEY="+key)
	}
	return env
}

func envTruthy(key string) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func WriteLog(dir, body string) error {
	path := filepath.Join(dir, "llm", "opencode.log")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}

func LogBody(res Result) string {
	var b strings.Builder
	b.WriteString("status: ")
	b.WriteString(string(res.Status))
	b.WriteString("\nbin: ")
	b.WriteString(res.Bin)
	b.WriteString("\nversion: ")
	b.WriteString(res.Version)
	if res.Err != "" {
		b.WriteString("\nerror: ")
		b.WriteString(res.Err)
	}
	b.WriteString("\n\n")
	b.WriteString(res.Output)
	b.WriteString("\n")
	return b.String()
}

// Fake is the test double. It records the prompt and optionally writes a marker file.
type Fake struct {
	Ran    int
	Prompt string
	Dir    string
	Result Result
}

func (f *Fake) Probe() (string, string, bool) {
	return "fake-opencode", "fake", true
}

func (f *Fake) Run(_ context.Context, req Request) (Result, error) {
	f.Ran++
	f.Prompt = req.Prompt
	f.Dir = req.Dir
	res := f.Result
	if res.Status == "" {
		res.Status = StatusRan
	}
	if res.Bin == "" {
		res.Bin = "fake-opencode"
	}
	if res.Output == "" {
		res.Output = "OpenCode fake: scaffold kept, marker written."
	}
	if err := os.MkdirAll(filepath.Join(req.Dir, "llm"), 0o755); err != nil {
		return res, err
	}
	marker := filepath.Join(req.Dir, "llm", "opencode.ran")
	if err := os.WriteFile(marker, []byte("ok\n"), 0o644); err != nil {
		return res, err
	}
	return res, nil
}
