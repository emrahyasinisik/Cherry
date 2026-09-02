package graph

import (
	"errors"
	"strings"

	"github.com/icerde/api/internal/connect"
	"github.com/icerde/api/internal/store"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const apiVersion = "0.8.0-local"

func gqlErr(err error) error {
	switch {
	case errors.Is(err, store.ErrValidation):
		_, after, ok := strings.Cut(err.Error(), ": ")
		if ok && after != "" && !strings.HasPrefix(after, "yığın") {
			return &gqlerror.Error{Message: after}
		}
		return &gqlerror.Error{Message: "E-posta geçerli olmalı, şifre en az 8 karakter."}
	case errors.Is(err, store.ErrInvalidCredentials):
		return &gqlerror.Error{Message: "Bilgiler geçersiz."}
	case errors.Is(err, store.ErrExists):
		return &gqlerror.Error{Message: "Bu e-posta kayıtlı. Giriş yap."}
	case errors.Is(err, store.ErrLocked):
		return &gqlerror.Error{Message: "Çok fazla deneme. Biraz bekle."}
	case errors.Is(err, store.ErrExpired):
		return &gqlerror.Error{Message: "Kod veya link süresi doldu."}
	case errors.Is(err, store.ErrMailFailed):
		return &gqlerror.Error{Message: "E-posta gönderilemedi. SMTP veya Resend ayarını kontrol et."}
	case errors.Is(err, store.ErrLLMFailed):
		return &gqlerror.Error{Message: "LLM çağrısı başarısız. Anahtar veya sarmalayıcıyı kontrol et."}
	case errors.Is(err, store.ErrBusy):
		return &gqlerror.Error{Message: "Ajan yazıyor. Bitince sohbete yaz."}
	case errors.Is(err, store.ErrPath):
		return &gqlerror.Error{Message: "MCP kökü dışına çıkılamaz."}
	case errors.Is(err, store.ErrUnauthorized):
		return &gqlerror.Error{Message: "Oturum gerekli."}
	case errors.Is(err, store.ErrNotFound):
		return &gqlerror.Error{Message: "Kayıt bulunamadı."}
	case errors.Is(err, connect.ErrConnect):
		return &gqlerror.Error{Message: strings.TrimPrefix(err.Error(), "bağlantı: ")}
	default:
		return &gqlerror.Error{Message: "İşlem yapılamadı."}
	}
}
