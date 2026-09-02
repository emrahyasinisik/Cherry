package factory

import (
	"fmt"
	"strings"

	"github.com/icerde/api/internal/store"
)

func parseBackendTarget(raw string) (store.BackendTarget, error) {
	if strings.TrimSpace(raw) == "" {
		return store.TargetLocal, nil
	}
	target := store.BackendTarget(raw)
	switch target {
	case store.TargetLocal, store.TargetSupabase, store.TargetCloudflare, store.TargetRender:
		return target, nil
	default:
		return "", fmt.Errorf("%w: backend %s", store.ErrValidation, raw)
	}
}

func backendLabel(target store.BackendTarget) (string, error) {
	switch target {
	case store.TargetLocal:
		return "Yerel API", nil
	case store.TargetSupabase:
		return "Supabase", nil
	case store.TargetCloudflare:
		return "Cloudflare", nil
	case store.TargetRender:
		return "Render", nil
	default:
		return "", fmt.Errorf("unhandled backend: %s", target)
	}
}

func parseStack(raw string) (store.ProjectStack, error) {
	switch store.ProjectStack(raw) {
	case store.StackExpo, store.StackFlutter, store.StackNative:
		return store.ProjectStack(raw), nil
	default:
		return "", fmt.Errorf("%w: yığın %s", store.ErrValidation, raw)
	}
}

func stackLabel(stack store.ProjectStack) (string, error) {
	switch stack {
	case store.StackExpo:
		return "Expo", nil
	case store.StackFlutter:
		return "Flutter", nil
	case store.StackNative:
		return "SwiftUI", nil
	default:
		return "", fmt.Errorf("unhandled stack: %s", stack)
	}
}

func frontendKind(stack store.ProjectStack) (string, error) {
	switch stack {
	case store.StackExpo:
		return "expo", nil
	case store.StackFlutter:
		return "flutter", nil
	case store.StackNative:
		return "swiftui", nil
	default:
		return "", fmt.Errorf("unhandled stack: %s", stack)
	}
}

func stackSourceRule(stack store.ProjectStack) (string, error) {
	switch stack {
	case store.StackExpo:
		return "Expo SDK 57, TypeScript strict, React 19, RN 0.86. Clean Architecture: src/domain, src/data, src/presentation, src/app; Expo Router yalnızca app/. Function component + hook. Class component yok. HTML site yazma. Katmanları tek dosyaya yığma. preview/ teslim değil.", nil
	case store.StackFlutter:
		return "Flutter 3.47 / Dart 3.13. Clean Architecture: lib/features/<özellik>/{domain,data,presentation}, lib/core, lib/app. sealed/final class, abstract interface, const, Material 3. Tek main.dart’a yığma. HTML yazma. preview/ teslim değil.", nil
	case store.StackNative:
		return "SwiftUI, Swift 6, iOS 18+. Clean Architecture: Domain, Data, Presentation, App. @Observable ViewModel, async/await, #Preview. UIKit ViewController yok. HTML yazma. preview/ teslim değil.", nil
	default:
		return "", fmt.Errorf("unhandled stack: %s", stack)
	}
}
