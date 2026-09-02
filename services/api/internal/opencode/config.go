package opencode

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cherry/api/internal/sidecar"
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

func WriteAgents(dir, name, stack, brief, sourceRule string) error {
	body := "# " + name + "\n\n" +
		"Bu klasör Cherry müşteri uygulamasıdır. OpenCode yalnızca bu kökte yazar.\n\n" +
		"- Yığın: " + stack + "\n" +
		"- Brif: " + brief + "\n" +
		"- " + sourceRule + "\n" +
		"- Asıl çıktı: `frontend/` (seçilen dil), `backend/`, `maestro/`.\n" +
		"- Mimari: Clean Architecture. domain / data / presentation / app (composition). Katmanları tek dosyaya yığma.\n" +
		"- Dil güncel kalsın: Expo SDK 57 + TS strict; Flutter 3.47 / Dart 3.13; SwiftUI Swift 6.\n" +
		"- `preview/*.html` stüdyo maketidir. Uygulamayı HTML ile değiştirme; zip’e HTML site koyma.\n" +
		"- Barındırma yok. Teslim klasör / zip / git — seçilen dilin kaynağı.\n" +
		"- Maestro YAML yaz; cihaz yoksa test çalıştırma, SKIPPED bırak.\n" +
		"- Cherry GraphQL’ine dokunma; müşteri backend’i ayrıdır.\n"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(body), 0o644)
}
