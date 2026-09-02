package connect

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CLIGit struct{}

func (CLIGit) Push(dir, repo, token string) error {
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("proje dizini yok")
	}
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("proje dizini yok")
	}
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git yok")
	}
	url := "https://x-access-token:" + token + "@github.com/" + repo + ".git"
	if err := runGit(dir, "add", "-A"); err != nil {
		return err
	}
	_ = runGit(dir, "-c", "user.email=icerde@local", "-c", "user.name=Icerde", "commit", "-m", "İçerde teslim")
	if err := runGit(dir, "push", url, "HEAD:main"); err != nil {
		if err2 := runGit(dir, "push", url, "HEAD:master"); err2 != nil {
			return fmt.Errorf("git push başarısız: %v", err)
		}
	}
	return nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return err
		}
		if strings.Contains(strings.ToLower(msg), "token") {
			return fmt.Errorf("git reddetti")
		}
		if len(msg) > 220 {
			msg = msg[:220] + "…"
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}
