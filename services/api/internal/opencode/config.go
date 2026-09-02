package opencode

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/icerde/api/internal/sidecar"
)

const configBare = `{
  "$schema": "https://opencode.ai/config.json"
}
`

func WriteConfig(dir string) error {
	body := configBare
	if hit, err := sidecar.Look("maestro"); err == nil {
		body = fmt.Sprintf(`{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "maestro": {
      "type": "local",
      "command": [%q, "mcp"]
    }
  }
}
`, hit.Path)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(body), 0o644)
}

func WriteAgents(dir, name, stack, brief string) error {
	body := "# " + name + "\n\n" +
		"Bu klasör İçerde müşteri uygulamasıdır. OpenCode yalnızca bu kökte yazar.\n\n" +
		"- Yığın: " + stack + "\n" +
		"- Brif: " + brief + "\n" +
		"- Çıktı: `frontend/`, `backend/`, `maestro/`, `preview/`\n" +
		"- Barındırma yok. Teslim klasör / zip / git.\n" +
		"- Maestro YAML yaz; cihaz yoksa test çalıştırma, SKIPPED bırak.\n" +
		"- İçerde GraphQL’ine dokunma; müşteri backend’i ayrıdır.\n"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(body), 0o644)
}
