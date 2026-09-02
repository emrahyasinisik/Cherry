package opencode

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cherry/api/internal/gdpr"
	"github.com/cherry/api/internal/sidecar"
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
	Dir      string
	Prompt   string
	Title    string
	Model    string
	Continue bool
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
	if raw := strings.TrimSpace(os.Getenv("CHERRY_OPENCODE_TIMEOUT_SEC")); raw != "" {
		if n, err := time.ParseDuration(raw + "s"); err == nil && n > 0 {
			timeout = n
		}
	}
	return &CLI{
		Bin:     strings.TrimSpace(os.Getenv("CHERRY_OPENCODE_BIN")),
		Timeout: timeout,
		Require: envTruthy("CHERRY_OPENCODE_REQUIRE"),
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
	hit, err := sidecar.Look("opencode")
	if err != nil {
		return "", err
	}
	return hit.Path, nil
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
	dir, err := absDir(req.Dir)
	if err != nil {
		return Result{Status: StatusFailed, Bin: bin, Err: err.Error()}, err
	}
	if err := WriteConfig(dir); err != nil {
		return Result{Status: StatusFailed, Bin: bin, Err: err.Error()}, err
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 8 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := []string{"run", "--dir", dir, "--auto", "--title", titleOr(req.Title, "cherry")}
	if req.Continue {
		args = append(args, "--continue")
	}
	if model := strings.TrimSpace(firstNonEmpty(req.Model, os.Getenv("CHERRY_OPENCODE_MODEL"), os.Getenv("CHERRY_LLM_MODEL"))); model != "" {
		args = append(args, "--model", model)
	}
	cmd := exec.CommandContext(runCtx, bin, args...)
	cmd.Dir = dir
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
		res.Err = failErr(runErr, raw)
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

func absDir(dir string) (string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", fmt.Errorf("proje dizini boş")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !st.IsDir() {
		return "", fmt.Errorf("proje dizini klasör değil: %s", abs)
	}
	return abs, nil
}

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func failErr(runErr error, raw string) string {
	text := strings.TrimSpace(ansiEscape.ReplaceAllString(raw, ""))
	if text == "" && runErr != nil {
		return runErr.Error()
	}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		low := strings.ToLower(line)
		if strings.Contains(low, "error:") || strings.Contains(low, "failed to") {
			return clipErr(line, 240)
		}
	}
	if text != "" {
		return clipErr(text, 240)
	}
	if runErr != nil {
		return runErr.Error()
	}
	return "opencode başarısız"
}

func clipErr(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
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
	key := strings.TrimSpace(os.Getenv("CHERRY_LLM_API_KEY"))
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
	Ran      int
	Prompt   string
	Dir      string
	Continue bool
	Result   Result
}

func (f *Fake) Probe() (string, string, bool) {
	return "fake-opencode", "fake", true
}

func (f *Fake) Run(_ context.Context, req Request) (Result, error) {
	f.Ran++
	f.Prompt = req.Prompt
	f.Dir = req.Dir
	f.Continue = req.Continue
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
