package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cherry/api/internal/sidecar"
)

// CompatibleProviderID is the OpenCode provider id for OpenAI-compatible
// endpoints (Colab FastAPI, local proxies). Built-in "openai" uses
// @ai-sdk/openai which posts to /v1/responses — Colab only serves
// /v1/chat/completions.
const CompatibleProviderID = "cherry-colab"

// DefaultColabModel matches the notebook BASE_MODEL / GET /v1/models id.
const DefaultColabModel = "Qwen/Qwen2.5-1.5B-Instruct"

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
		remoteModel := strings.TrimSpace(ep.Model)
		if remoteModel == "" {
			remoteModel = DefaultColabModel
		}
		// Strip accidental provider prefixes from callers.
		remoteModel = stripProviderPrefix(remoteModel)
		localID := localModelID(remoteModel)
		root["model"] = CompatibleProviderID + "/" + localID
		root["provider"] = map[string]any{
			CompatibleProviderID: map[string]any{
				"npm":  "@ai-sdk/openai-compatible",
				"name": "Cherry Colab",
				"options": map[string]any{
					"baseURL": base,
					"apiKey":  "{env:OPENAI_API_KEY}",
				},
				"models": map[string]any{
					localID: map[string]any{
						"name": displayModelName(remoteModel),
						// Wire OpenCode local id → remote /v1/models id (may contain '/').
						"id": remoteModel,
					},
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

// CLIModel returns the provider/model string for `opencode run --model`.
func CLIModel(ep Endpoint) string {
	base := strings.TrimRight(strings.TrimSpace(ep.BaseURL), "/")
	model := strings.TrimSpace(ep.Model)
	if base != "" {
		if model == "" {
			model = DefaultColabModel
		}
		model = stripProviderPrefix(model)
		return CompatibleProviderID + "/" + localModelID(model)
	}
	if model == "" {
		return ""
	}
	if !strings.Contains(model, "/") {
		return "openai/" + model
	}
	return model
}

func stripProviderPrefix(model string) string {
	for _, prefix := range []string{CompatibleProviderID + "/", "openai/"} {
		if strings.HasPrefix(model, prefix) {
			return strings.TrimPrefix(model, prefix)
		}
	}
	return model
}

// localModelID makes a slash-free OpenCode model key. Remote id stays in models[].id.
func localModelID(remote string) string {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return "model"
	}
	return strings.ReplaceAll(remote, "/", "-")
}

func displayModelName(remote string) string {
	if i := strings.LastIndex(remote, "/"); i >= 0 && i+1 < len(remote) {
		return remote[i+1:]
	}
	return remote
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
