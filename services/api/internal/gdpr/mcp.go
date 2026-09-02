package gdpr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cherry/api/internal/store"
)

func ReadFile(root, rel string) ([]byte, error) {
	root = filepath.Clean(root)
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("%w: MCP kökü boş", store.ErrPath)
	}
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return nil, fmt.Errorf("%w: yol boş", store.ErrPath)
	}
	if filepath.IsAbs(rel) {
		return nil, fmt.Errorf("%w: mutlak yol yok", store.ErrPath)
	}
	joined := filepath.Join(root, filepath.FromSlash(rel))
	clean := filepath.Clean(joined)
	sep := string(os.PathSeparator)
	if clean != root && !strings.HasPrefix(clean, root+sep) {
		return nil, fmt.Errorf("%w: kök dışı", store.ErrPath)
	}
	data, err := os.ReadFile(clean)
	if err != nil {
		return nil, err
	}
	return data, nil
}
