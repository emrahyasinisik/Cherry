package store

import (
	"context"
	"testing"
)

func TestMemoryLoginValidation(t *testing.T) {
	mem := NewMemory()
	ctx := context.Background()

	_, _, err := mem.Login(ctx, "not-an-email", "x")
	if err != ErrValidation {
		t.Fatalf("expected validation, got %v", err)
	}

	token, user, err := mem.Login(ctx, "ada@icerde.dev", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || user.Email != "ada@icerde.dev" {
		t.Fatalf("unexpected login: %#v %#v", token, user)
	}

	me, err := mem.Me(ctx, token)
	if err != nil || me.Email != user.Email {
		t.Fatalf("me: %v %#v", err, me)
	}

	projects, err := mem.Projects(ctx, token)
	if err != nil || len(projects) != 0 {
		t.Fatalf("projects should be empty: %v %#v", err, projects)
	}
}
