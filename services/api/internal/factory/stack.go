package factory

import (
	"fmt"

	"github.com/icerde/api/internal/store"
)

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
		return "Expo / React Native", nil
	case store.StackFlutter:
		return "Flutter", nil
	case store.StackNative:
		return "Native iOS + Android", nil
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
		return "native-stub", nil
	default:
		return "", fmt.Errorf("unhandled stack: %s", stack)
	}
}

func stackSourceRule(stack store.ProjectStack) (string, error) {
	switch stack {
	case store.StackExpo:
		return "Uygulama kodu TypeScript / React Native (Expo): `frontend/` içinde `.tsx` ve `.ts`. HTML/CSS sitesi yazma. `preview/` yalnızca stüdyo maketi, teslim değil.", nil
	case store.StackFlutter:
		return "Uygulama kodu Dart (Flutter): `frontend/pubspec.yaml` ve `frontend/lib/*.dart`. HTML sitesi yazma. `preview/` yalnızca stüdyo maketi, teslim değil.", nil
	case store.StackNative:
		return "Uygulama kodu Swift (`frontend/ios`) ve Kotlin (`frontend/android`). HTML sitesi yazma. `preview/` yalnızca stüdyo maketi, teslim değil.", nil
	default:
		return "", fmt.Errorf("unhandled stack: %s", stack)
	}
}
