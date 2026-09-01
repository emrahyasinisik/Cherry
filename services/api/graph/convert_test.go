package graph

import (
	"testing"

	"github.com/icerde/api/internal/store"
)

func TestMapWorkspaceKindExhaustive(t *testing.T) {
	personal, err := mapWorkspaceKind(store.WorkspacePersonal)
	if err != nil || personal != WorkspaceKindPersonal {
		t.Fatalf("personal: %v %s", err, personal)
	}
	org, err := mapWorkspaceKind(store.WorkspaceOrganization)
	if err != nil || org != WorkspaceKindOrganization {
		t.Fatalf("org: %v %s", err, org)
	}
	if _, err := mapWorkspaceKind(store.WorkspaceKind("NOPE")); err == nil {
		t.Fatal("expected error for unknown kind")
	}
}
