package auth

import (
	"context"
	"strings"
	"testing"

	"github.com/icerde/api/internal/mailer"
	"github.com/icerde/api/internal/store"
)

func TestRegisterVerifyCodeSession(t *testing.T) {
	mem := store.NewMemory()
	mail := &mailer.Service{Store: mem, WebURL: "http://127.0.0.1:43147"}
	svc := New(mem, mail, "pepper", "http://127.0.0.1:43147")
	ctx := context.Background()

	result, err := svc.Register(ctx, "ada@icerde.dev", "secret12", "fp-1", "Test", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Next != NextDeviceCode || result.ChallengeID == "" {
		t.Fatalf("expected device code, got %#v", result)
	}

	msg, err := mem.MailByChallenge(ctx, result.ChallengeID)
	if err != nil {
		t.Fatal(err)
	}
	code := extractCode(msg.Body)
	if len(code) != 6 {
		t.Fatalf("code in mailbox: %q body=%q", code, msg.Body)
	}

	verified, err := svc.VerifyCode(ctx, result.ChallengeID, code, true, "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if verified.Next != NextSession || verified.Token == "" {
		t.Fatalf("expected session, got %#v", verified)
	}

	user, sess, err := svc.SessionUser(ctx, verified.Token)
	if err != nil || user.Email != "ada@icerde.dev" || sess == nil {
		t.Fatalf("session: %v %#v", err, user)
	}

	again, err := svc.Login(ctx, "ada@icerde.dev", "secret12", "fp-1", "Test", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if again.Next != NextSession {
		t.Fatalf("trusted device should skip code, got %#v", again)
	}
}

func extractCode(body string) string {
	for _, field := range strings.Fields(body) {
		if len(field) == 6 {
			ok := true
			for _, r := range field {
				if r < '0' || r > '9' {
					ok = false
					break
				}
			}
			if ok {
				return field
			}
		}
	}
	return ""
}
