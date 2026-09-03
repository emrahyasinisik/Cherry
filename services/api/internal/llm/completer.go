package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cherry/api/internal/store"
)

type MockCompleter struct{}

func (MockCompleter) Channel() string { return "mock" }

func (MockCompleter) Complete(_ context.Context, version store.LlmVersion, prompt string) (string, error) {
	who := "LLM A"
	if version.Slot == store.SlotB {
		who = "LLM B"
	}
	switch version.ID {
	case "ver-a-2", "ver-b-2":
		return who + " " + version.Name + " plan:\n- Ekranları yığın dosyalarına böl.\n- Maestro akışını giriş + ana ekran tut.\nKaynak (redakte):\n" + clip(prompt, 400), nil
	case "ver-a-1", "ver-b-1":
		return who + " " + version.Name + " iskelet:\n- README ve frontend/backend duruyor.\n- OpenCode bu planla dosyaları yazar.\nBrif (redakte):\n" + clip(prompt, 400), nil
	default:
		return who + " " + version.Name + ":\n" + clip(prompt, 400), nil
	}
}

type HTTPCompleter struct {
	Key     string
	BaseURL string
	Model   string
	Client  *http.Client
}

func (c HTTPCompleter) Channel() string { return "http" }

func (c HTTPCompleter) Complete(ctx context.Context, version store.LlmVersion, prompt string) (string, error) {
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	model := c.Model
	if model == "" {
		model = "gpt-4o-mini"
	}
	body, err := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "Cherry kod ajanısın. Kısa Türkçe plan yaz. PII uydurma. Versiyon: " + version.Name},
			{"role": "user", "content": prompt},
		},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("status %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty completion")
	}
	return parsed.Choices[0].Message.Content, nil
}

// ColabTunnelCompleter sends requests to a Colab inference tunnel
// (named fixed hostname or trycloudflare quick tunnel).
// No API key — the Colab server is unauthenticated behind the tunnel.
// Cherry stores only the public HTTPS URL; the Cloudflare tunnel token stays in Colab.
type ColabTunnelCompleter struct {
	BaseURL string
	Model   string
	Client  *http.Client
}

func (c ColabTunnelCompleter) Channel() string { return "colab-tunnel" }

func (c ColabTunnelCompleter) Complete(ctx context.Context, version store.LlmVersion, prompt string) (string, error) {
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		return "", fmt.Errorf("colab tunnel URL boş")
	}
	model := c.Model
	if model == "" {
		model = "Qwen/Qwen2.5-1.5B-Instruct"
	}
	body, err := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "Cherry kod ajanısın. Kısa Türkçe plan yaz. PII uydurma. Versiyon: " + version.Name},
			{"role": "user", "content": prompt},
		},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 90 * time.Second}
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("colab tunnel status %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty colab tunnel completion")
	}
	return parsed.Choices[0].Message.Content, nil
}

func NewCompleter(apiKey, baseURL, model string) Completer {
	if strings.TrimSpace(apiKey) == "" {
		return MockCompleter{}
	}
	return HTTPCompleter{Key: apiKey, BaseURL: baseURL, Model: model}
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
