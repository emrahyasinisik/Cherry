package store

import (
	"context"
	"os"
	"testing"
	"time"
)

func openTestMongo(t *testing.T) *Mongo {
	t.Helper()
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://127.0.0.1:27017/cherry_test"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	m, err := OpenMongo(ctx, uri)
	if err != nil {
		t.Skipf("mongo unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = m.db.Drop(context.Background())
		_ = m.Close(context.Background())
	})
	return m
}

func TestMongoCreateUserAndMail(t *testing.T) {
	m := openTestMongo(t)
	ctx := context.Background()
	user, err := m.CreateUser(ctx, User{Email: "ada@cherry.dev", PasswordHash: "x", WorkspaceKind: WorkspacePersonal})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.CreateUser(ctx, User{Email: "Ada@cherry.dev", PasswordHash: "x"}); err != ErrExists {
		t.Fatalf("expected exists, got %v", err)
	}
	got, err := m.GetUserByEmail(ctx, "ada@cherry.dev")
	if err != nil || got.ID != user.ID {
		t.Fatalf("get by email: %v %#v", err, got)
	}
	if err := m.AddMail(ctx, Mail{UserID: user.ID, ChallengeID: "c1", Subject: "s", Body: "b", CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	mail, err := m.MailByChallenge(ctx, "c1")
	if err != nil || mail.Body != "b" {
		t.Fatalf("mail: %v %#v", err, mail)
	}
}

func TestMongoSessionAndDelete(t *testing.T) {
	m := openTestMongo(t)
	ctx := context.Background()
	user, err := m.CreateUser(ctx, User{Email: "bob@cherry.dev", PasswordHash: "x", WorkspaceKind: WorkspacePersonal})
	if err != nil {
		t.Fatal(err)
	}
	sess := Session{ID: NewID(), UserID: user.ID, TokenHash: "th-" + NewID(), CreatedAt: time.Now().UTC()}
	if err := m.CreateSession(ctx, sess); err != nil {
		t.Fatal(err)
	}
	got, err := m.GetSessionByTokenHash(ctx, sess.TokenHash)
	if err != nil || got.ID != sess.ID {
		t.Fatalf("session: %v %#v", err, got)
	}
	if err := m.DeleteUserData(ctx, user.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := m.GetUserByID(ctx, user.ID); err != ErrNotFound {
		t.Fatalf("expected not found after delete, got %v", err)
	}
	if _, err := m.GetSessionByTokenHash(ctx, sess.TokenHash); err != ErrUnauthorized {
		t.Fatalf("expected unauthorized after delete, got %v", err)
	}
}

func TestDBNameFromURI(t *testing.T) {
	cases := map[string]string{
		"mongodb://127.0.0.1:27017/cherry":          "cherry",
		"mongodb://127.0.0.1:27017/cherry_test?w=1": "cherry_test",
		"mongodb://user:pass@host:27017/":           "cherry",
		"mongodb://host:27017":                      "cherry",
	}
	for uri, want := range cases {
		if got := dbNameFromURI(uri); got != want {
			t.Fatalf("%s: got %q want %q", uri, got, want)
		}
	}
}
