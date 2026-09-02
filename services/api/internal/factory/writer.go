package factory

import (
	"archive/zip"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/icerde/api/internal/store"
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
	out := []fileSpec{
		{
			rel:  "README.md",
			kind: "readme",
			body: "# " + name + "\n\n" + brief + "\n\nYığın: " + label + " (`" + kind + "`).\n\nİçerde bu klasörü yazar; barındırma yoktur. Teslim: bu dizin / zip / git.\n",
		},
		{
			rel:  "backend/README.md",
			kind: "backend",
			body: "# Müşteri API\n\nBu, İçerde GraphQL’i değil üretilen uygulamanın backend’idir.\nYerel aktif localhost’ta ayağa kaldırır (47000–47999).\n",
		},
		{
			rel:  "backend/main.go",
			kind: "backend",
			body: "package main\n\nimport (\n\t\"encoding/json\"\n\t\"log\"\n\t\"net/http\"\n\t\"os\"\n)\n\nfunc main() {\n\taddr := os.Getenv(\"ICERDE_CUSTOMER_ADDR\")\n\tif addr == \"\" {\n\t\taddr = \"127.0.0.1:18080\"\n\t}\n\thttp.HandleFunc(\"/health\", func(w http.ResponseWriter, _ *http.Request) {\n\t\t_ = json.NewEncoder(w).Encode(map[string]any{\"ok\": true, \"app\": \"" + slug + "\"})\n\t})\n\tlog.Println(\"generated customer api on\", addr)\n\tlog.Fatal(http.ListenAndServe(addr, nil))\n}\n",
		},
		{
			rel:  "maestro/login.yaml",
			kind: "maestro",
			body: "appId: dev.icerde." + slug + "\n---\n- launchApp\n- assertVisible: \"Giriş\"\n- tapOn: \"Devam\"\n",
		},
		{
			rel:  "maestro/home.yaml",
			kind: "maestro",
			body: "appId: dev.icerde." + slug + "\n---\n- launchApp\n- assertVisible: \"" + name + "\"\n",
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

func frontendFiles(stack store.ProjectStack, name, slug, brief string) ([]fileSpec, error) {
	switch stack {
	case store.StackExpo:
		return []fileSpec{
			{rel: "frontend/package.json", kind: "frontend", body: "{\n  \"name\": \"" + slug + "\",\n  \"private\": true,\n  \"main\": \"expo-router/entry\"\n}\n"},
			{rel: "frontend/app/index.tsx", kind: "frontend", body: "import { Text, View } from \"react-native\";\n\nexport default function Home() {\n  return (\n    <View>\n      <Text>" + name + "</Text>\n      <Text>" + strings.ReplaceAll(brief, "\"", "'") + "</Text>\n    </View>\n  );\n}\n"},
		}, nil
	case store.StackFlutter:
		return []fileSpec{
			{rel: "frontend/pubspec.yaml", kind: "frontend", body: "name: " + slug + "\ndescription: " + strings.ReplaceAll(brief, "\n", " ") + "\n"},
			{rel: "frontend/lib/main.dart", kind: "frontend", body: "import 'package:flutter/material.dart';\n\nvoid main() => runApp(const MaterialApp(home: Scaffold(body: Center(child: Text('" + name + "')))));\n"},
		}, nil
	case store.StackNative:
		return []fileSpec{
			{rel: "frontend/README.md", kind: "frontend", body: "# Native stub\n\n" + name + " — iOS + Android iskeleti. Tam native adaptör henüz bağlı değil; sahte başarı yok, klasör duruyor.\n"},
			{rel: "frontend/ios/README.md", kind: "frontend", body: "iOS hedefi (stub).\n"},
			{rel: "frontend/android/README.md", kind: "frontend", body: "Android hedefi (stub).\n"},
		}, nil
	default:
		return nil, fmt.Errorf("unhandled stack: %s", stack)
	}
}

func screenHTML(app, title, lead, detail, variant string) string {
	accent := "#C4A574"
	if variant == "home" {
		accent = "#6F9E7A"
	}
	return `<!doctype html><html lang="tr"><head><meta charset="utf-8"><style>
body{margin:0;background:#0E1114;color:#E8E4DC;font-family:ui-sans-serif,system-ui,sans-serif}
.phone{padding:28px 20px;min-height:100vh;box-sizing:border-box}
.mark{font-size:11px;letter-spacing:.08em;color:#8B939C;margin:0 0 20px}
h1{font-size:22px;font-weight:500;margin:0 0 8px}
p{font-size:13px;line-height:1.45;color:#8B939C;margin:0 0 16px}
.cta{display:block;background:` + accent + `;color:#0E1114;border:0;border-radius:8px;padding:10px 14px;font-size:13px;width:100%}
.field{height:36px;border:1px solid #2A323A;border-radius:8px;margin:0 0 8px;background:#161B20}
</style></head><body><div class="phone">
<p class="mark">` + html.EscapeString(app) + `</p>
<h1>` + html.EscapeString(title) + `</h1>
<p>` + html.EscapeString(lead) + `</p>
` + fieldsFor(variant) + `
<p>` + html.EscapeString(clip(detail, 180)) + `</p>
<button class="cta">Devam</button>
</div></body></html>`
}

func fieldsFor(variant string) string {
	if variant != "login" {
		return `<div class="field"></div><div class="field"></div>`
	}
	return `<div class="field"></div><div class="field"></div>`
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
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
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
	case "icerde.zip", "opencode.json", "AGENTS.md":
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
