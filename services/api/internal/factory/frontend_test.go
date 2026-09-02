package factory

import (
	"strings"
	"testing"

	"github.com/cherry/api/internal/store"
)

func TestFrontendCleanArchTrees(t *testing.T) {
	cases := []struct {
		stack store.ProjectStack
		must  []string
	}{
		{
			stack: store.StackExpo,
			must: []string{
				"frontend/ARCHITECTURE.md",
				"frontend/src/domain/entities/item.ts",
				"frontend/src/data/repositories/item-repository-impl.ts",
				"frontend/src/presentation/screens/home-screen.tsx",
				"frontend/src/app/composition.ts",
				"frontend/app/index.tsx",
			},
		},
		{
			stack: store.StackFlutter,
			must: []string{
				"frontend/ARCHITECTURE.md",
				"frontend/lib/features/home/domain/entities/home_item.dart",
				"frontend/lib/features/home/data/repositories/home_repository_impl.dart",
				"frontend/lib/features/home/presentation/pages/home_page.dart",
				"frontend/lib/app/di.dart",
				"frontend/lib/main.dart",
			},
		},
		{
			stack: store.StackNative,
			must: []string{
				"frontend/ARCHITECTURE.md",
				"frontend/Domain/Entities/HomeItem.swift",
				"frontend/Data/Repositories/HomeRepositoryImpl.swift",
				"frontend/Presentation/Home/HomeView.swift",
				"frontend/App/Composition.swift",
			},
		},
	}
	for _, tc := range cases {
		t.Run(string(tc.stack), func(t *testing.T) {
			files, err := frontendFiles(tc.stack, "Kahve", "kahve", "sipariş kuyruğu uygulaması")
			if err != nil {
				t.Fatal(err)
			}
			seen := map[string]string{}
			for _, file := range files {
				seen[file.rel] = file.body
				if strings.Contains(strings.ToLower(file.body), "<!doctype html") {
					t.Fatalf("scaffold HTML in %s", file.rel)
				}
			}
			for _, rel := range tc.must {
				body, ok := seen[rel]
				if !ok || strings.TrimSpace(body) == "" {
					t.Fatalf("missing %s", rel)
				}
			}
			arch := seen["frontend/ARCHITECTURE.md"]
			if !strings.Contains(arch, "Clean Architecture") {
				t.Fatalf("ARCHITECTURE.md: %s", arch)
			}
		})
	}
}
