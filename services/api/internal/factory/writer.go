package factory

import (
	"archive/zip"
	"html"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/cherry/api/internal/store"
)

type fileSpec struct {
	rel  string
	body string
	kind string
}

func writeTree(project store.Project) error {
	label, err := stackLabel(project.Stack)
	if err != nil {
		return err
	}
	kind, err := frontendKind(project.Stack)
	if err != nil {
		return err
	}
	files, err := treeFiles(project, label, kind)
	if err != nil {
		return err
	}
	for _, file := range files {
		path := filepath.Join(project.RootPath, filepath.FromSlash(file.rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(file.body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func treeFiles(project store.Project, label, kind string) ([]fileSpec, error) {
	name := strings.TrimSpace(project.Name)
	brief := strings.TrimSpace(project.Brief)
	slug := slugify(name)
	be := project.Backend
	if be == "" {
		be = store.TargetLocal
	}
	beLabel, err := backendLabel(be)
	if err != nil {
		return nil, err
	}
	out := []fileSpec{
		{
			rel:  "README.md",
			kind: "readme",
			body: "# " + name + "\n\n" + brief + "\n\nYığın: " + label + " (`" + kind + "`).\nMimari: Clean Architecture (domain / data / presentation).\nBackend hedefi: " + beLabel + ".\n\nCherry bu klasörü yazar; barındırma yoktur. Teslim: bu dizin / zip / git — seçilen dil, HTML değil.\n",
		},
		{
			rel:  "backend/README.md",
			kind: "backend",
			body: "# Müşteri API\n\nBu, Cherry GraphQL’i değil üretilen uygulamanın backend’idir.\nHedef: " + beLabel + ".\nYerel aktif localhost’ta ayağa kaldırır (47000–47999).\n",
		},
		{
			rel:  "backend/TARGET.md",
			kind: "backend",
			body: "# Backend hedefi\n\n" + beLabel + " (`" + string(be) + "`).\n\nToken zip’e yazılmaz. Kişinin Bağlantılar hesabı. Cherry host değil.\n",
		},
		{
			rel:  "backend/main.go",
			kind: "backend",
			body: "package main\n\nimport (\n\t\"encoding/json\"\n\t\"log\"\n\t\"net/http\"\n\t\"os\"\n)\n\nfunc main() {\n\taddr := os.Getenv(\"CHERRY_CUSTOMER_ADDR\")\n\tif addr == \"\" {\n\t\taddr = \"127.0.0.1:18080\"\n\t}\n\thttp.HandleFunc(\"/health\", func(w http.ResponseWriter, _ *http.Request) {\n\t\t_ = json.NewEncoder(w).Encode(map[string]any{\"ok\": true, \"app\": \"" + slug + "\"})\n\t})\n\tlog.Println(\"generated customer api on\", addr)\n\tlog.Fatal(http.ListenAndServe(addr, nil))\n}\n",
		},
		{
			rel:  "maestro/login.yaml",
			kind: "maestro",
			body: "appId: dev.cherry." + slug + "\n---\n- launchApp\n- assertVisible: \"Giriş\"\n- tapOn: \"Devam\"\n",
		},
		{
			rel:  "maestro/home.yaml",
			kind: "maestro",
			body: "appId: dev.cherry." + slug + "\n---\n- launchApp\n- assertVisible: \"" + name + "\"\n",
		},
		{
			rel:  "preview/login.html",
			kind: "preview",
			body: screenHTML(name, "Giriş", "E-posta ve şifre ile içeri. SMS yok.", brief, "login"),
		},
		{
			rel:  "preview/home.html",
			kind: "preview",
			body: screenHTML(name, "Ana ekran", brief, "Ajanın ürettiği kabuk. OpenCode bu dosyaları doldurur.", "home"),
		},
	}
	fe, err := frontendFiles(project.Stack, name, slug, brief)
	if err != nil {
		return nil, err
	}
	return append(out, fe...), nil
}

func screenHTML(app, title, lead, detail, variant string) string {
	accent := "#C4A574"
	if variant == "home" {
		accent = "#6F9E7A"
	}
	cta := "Devam"
	emptyHint := ""
	switch variant {
	case "login":
		emptyHint = `<p class="hint">E-posta · şifre — SMS yok</p>`
	case "home":
		emptyHint = `<p class="hint">Bugünkü siparişler burada durur. Henüz kayıt yok.</p>`
		cta = "Menüye bak"
	default:
		emptyHint = `<p class="hint">İçerik bekleniyor.</p>`
	}
	return `<!doctype html><html lang="tr"><head><meta charset="utf-8"><style>
body{margin:0;background:#0E1114;color:#E8E4DC;font-family:ui-sans-serif,system-ui,sans-serif}
.phone{padding:28px 20px;min-height:100vh;box-sizing:border-box;background:radial-gradient(120% 80% at 10% 0%,#161B20 0%,#0E1114 55%)}
.mark{font-size:11px;letter-spacing:.08em;color:#8B939C;margin:0 0 20px}
h1{font-size:22px;font-weight:500;margin:0 0 8px;color:#E8E4DC}
p{font-size:13px;line-height:1.45;color:#8B939C;margin:0 0 16px}
.hint{font-size:12px;color:#8B939C;margin:8px 0 16px}
.cta{display:block;background:` + accent + `;color:#0E1114;border:0;border-radius:8px;padding:10px 14px;font-size:13px;width:100%}
.field{height:36px;border:1px solid #2A323A;border-radius:8px;margin:0 0 8px;background:#161B20;display:flex;align-items:center;padding:0 10px;font-size:12px;color:#8B939C}
.card{border:1px dashed #2A323A;border-radius:10px;padding:14px;margin:0 0 16px;background:#12161A}
</style></head><body><div class="phone">
<p class="mark">` + html.EscapeString(app) + `</p>
<h1>` + html.EscapeString(title) + `</h1>
<p>` + html.EscapeString(lead) + `</p>
` + fieldsFor(variant) + `
` + emptyHint + `
<p>` + html.EscapeString(clip(detail, 180)) + `</p>
<button class="cta">` + html.EscapeString(cta) + `</button>
</div></body></html>`
}

func fieldsFor(variant string) string {
	switch variant {
	case "login":
		return `<div class="field">e-posta</div><div class="field">şifre</div>`
	case "home":
		return `<div class="card"><p class="hint" style="margin:0">Boş tezgâh — ilk sipariş gelince liste dolar.</p></div>`
	default:
		return `<div class="field"></div>`
	}
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func slugify(name string) string {
	var b strings.Builder
	lastDash := true
	for _, r := range name {
		r = asciiFold(r)
		if r >= 'A' && r <= 'Z' {
			r = r - 'A' + 'a'
		} else {
			r = unicode.ToLower(r)
			r = asciiFold(r)
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "uygulama"
	}
	return out
}

// asciiFold maps Turkish and common Latin letters to ASCII for Expo slug / appId.
func asciiFold(r rune) rune {
	switch r {
	case 'ı', 'İ', 'I', 'ì', 'í', 'î', 'ï':
		return 'i'
	case 'ğ', 'Ğ':
		return 'g'
	case 'ü', 'Ü', 'ù', 'ú', 'û':
		return 'u'
	case 'ş', 'Ş', 'š', 'Š':
		return 's'
	case 'ö', 'Ö', 'ò', 'ó', 'ô':
		return 'o'
	case 'ç', 'Ç':
		return 'c'
	case 'ä', 'Ä', 'à', 'á', 'â', 'ã':
		return 'a'
	case 'é', 'è', 'ê', 'ë', 'É':
		return 'e'
	case 'ñ', 'Ñ':
		return 'n'
	default:
		return r
	}
}

func zipProject(root, dest string) error {
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()
	zw := zip.NewWriter(file)
	defer zw.Close()
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		slash := filepath.ToSlash(rel)
		if info.IsDir() {
			if path != root && skipDeliveryDir(filepath.Base(path)) {
				return filepath.SkipDir
			}
			return nil
		}
		if skipDeliveryRel(slash) {
			return nil
		}
		w, err := zw.Create(slash)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	})
}

func listFiles(root string) ([]string, error) {
	var out []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	return out, err
}

func skipDeliveryDir(base string) bool {
	switch base {
	case ".git", "node_modules", ".opencode", "preview", "llm":
		return true
	default:
		return false
	}
}

func skipDeliveryRel(rel string) bool {
	rel = filepath.ToSlash(rel)
	switch rel {
	case "cherry.zip", "opencode.json", "AGENTS.md":
		return true
	}
	first, _, _ := strings.Cut(rel, "/")
	return skipDeliveryDir(first)
}

func rankDelivery(rel string) int {
	switch {
	case rel == "README.md":
		return 0
	case strings.HasPrefix(rel, "frontend/"):
		return 1
	case strings.HasPrefix(rel, "backend/"):
		return 2
	case strings.HasPrefix(rel, "maestro/"):
		return 3
	default:
		return 4
	}
}
