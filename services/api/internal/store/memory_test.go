package store

import (
	"context"
	"testing"
)

func TestMemoryCreateUserAndMail(t *testing.T) {
	mem := NewMemory()
	ctx := context.Background()
	user, err := mem.CreateUser(ctx, User{Email: "ada@cherry.dev", PasswordHash: "x", WorkspaceKind: WorkspacePersonal})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mem.CreateUser(ctx, User{Email: "ada@cherry.dev", PasswordHash: "x"}); err != ErrExists {
		t.Fatalf("expected exists, got %v", err)
	}
	if err := mem.AddMail(ctx, Mail{UserID: user.ID, ChallengeID: "c1", Subject: "s", Body: "b"}); err != nil {
		t.Fatal(err)
	}
	got, err := mem.MailByChallenge(ctx, "c1")
	if err != nil || got.Body != "b" {
		t.Fatalf("mail: %v %#v", err, got)
	}
}
