package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cherry/api/internal/sidecar"
)

type Endpoint struct {
	BaseURL string
	APIKey  string
	Model   string
}

func EndpointFromEnv() Endpoint {
	return Endpoint{
		BaseURL: strings.TrimSpace(os.Getenv("CHERRY_LLM_BASE_URL")),
		APIKey:  strings.TrimSpace(os.Getenv("CHERRY_LLM_API_KEY")),
		Model:   strings.TrimSpace(firstNonEmpty(os.Getenv("CHERRY_OPENCODE_MODEL"), os.Getenv("CHERRY_LLM_MODEL"))),
	}
}

func WriteConfig(dir string, ep ...Endpoint) error {
	endpoint := EndpointFromEnv()
	if len(ep) > 0 {
		if strings.TrimSpace(ep[0].BaseURL) != "" {
			endpoint.BaseURL = strings.TrimSpace(ep[0].BaseURL)
		}
		if strings.TrimSpace(ep[0].APIKey) != "" {
			endpoint.APIKey = strings.TrimSpace(ep[0].APIKey)
		}
		if strings.TrimSpace(ep[0].Model) != "" {
			endpoint.Model = strings.TrimSpace(ep[0].Model)
		}
	}
	body, err := configJSON(endpoint)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "opencode.json"), body, 0o644)
}

func configJSON(ep Endpoint) ([]byte, error) {
	root := map[string]any{
		"$schema": "https://opencode.ai/config.json",
	}
	base := strings.TrimRight(strings.TrimSpace(ep.BaseURL), "/")
	if base != "" {
		model := strings.TrimSpace(ep.Model)
		if model == "" {
			model = "gpt-4o-mini"
		}
		if !strings.HasPrefix(model, "openai/") {
			model = "openai/" + model
		}
		root["model"] = model
		root["provider"] = map[string]any{
			"openai": map[string]any{
				"options": map[string]any{
					"baseURL": base,
					"apiKey":  "{env:OPENAI_API_KEY}",
				},
			},
		}
	}
	if hit, err := sidecar.Look("maestro"); err == nil {
		root["mcp"] = map[string]any{
			"maestro": map[string]any{
				"type":    "local",
				"command": []string{hit.Path, "mcp"},
			},
		}
	}
	raw, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("opencode config: %w", err)
	}
	return append(raw, '\n'), nil
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
