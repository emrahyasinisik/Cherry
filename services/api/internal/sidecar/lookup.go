package sidecar

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Hit struct {
	Path   string
	Source string
}

func Look(name string) (Hit, error) {
	envKey := envKeyFor(name)
	if envKey != "" {
		if p := strings.TrimSpace(os.Getenv(envKey)); p != "" {
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return Hit{Path: p, Source: "env"}, nil
			}
		}
	}
	for _, dir := range bundledDirs() {
		for _, candidate := range names(name) {
			p := filepath.Join(dir, candidate)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return Hit{Path: p, Source: "bundled"}, nil
			}
		}
	}
	if p, err := exec.LookPath(name); err == nil {
		return Hit{Path: p, Source: "path"}, nil
	}
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath(name + ".exe"); err == nil {
			return Hit{Path: p, Source: "path"}, nil
		}
	}
	return Hit{}, os.ErrNotExist
}

func envKeyFor(name string) string {
	switch strings.ToLower(name) {
	case "opencode":
		return "CHERRY_OPENCODE_BIN"
	case "maestro":
		return "CHERRY_MAESTRO_BIN"
	default:
		return ""
	}
}

func names(name string) []string {
	if runtime.GOOS == "windows" {
		return []string{name + ".exe", name}
	}
	return []string{name}
}

func bundledDirs() []string {
	var out []string
	if dir := strings.TrimSpace(os.Getenv("CHERRY_SIDECAR_DIR")); dir != "" {
		out = append(out, dir)
		return unique(out)
	}
	if exe, err := os.Executable(); err == nil {
		base := filepath.Dir(exe)
		out = append(out,
			filepath.Join(base, "resources", "bin"),
			filepath.Join(base, "..", "resources", "bin"),
			filepath.Join(base, "bin"),
		)
	}
	cwd, _ := os.Getwd()
	dir := cwd
	for i := 0; i < 8 && dir != ""; i++ {
		out = append(out, filepath.Join(dir, "vendor", "bin"))
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return unique(out)
}

func unique(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, p := range in {
		clean, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	return out
}
