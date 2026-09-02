package llm

import (
	"os"
	"path/filepath"
	"strings"
)

func ColabDir() string {
	if value := strings.TrimSpace(os.Getenv("CHERRY_COLAB_DIR")); value != "" {
		if abs, err := filepath.Abs(value); err == nil {
			return abs
		}
		return value
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	candidates := []string{
		filepath.Join(wd, "..", "..", "colab"),
		filepath.Join(wd, "colab"),
	}
	for _, candidate := range candidates {
		info, statErr := os.Stat(candidate)
		if statErr != nil || !info.IsDir() {
			continue
		}
		abs, absErr := filepath.Abs(candidate)
		if absErr != nil {
			return candidate
		}
		return abs
	}
	return ""
}
