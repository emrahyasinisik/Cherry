package graph

import (
	"testing"

	"github.com/icerde/api/internal/mailer"
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

func TestMapEmailChannelExhaustive(t *testing.T) {
	inbox, err := mapEmailChannel(mailer.ChannelInbox)
	if err != nil || inbox != mailer.ChannelInbox {
		t.Fatalf("inbox: %v %s", err, inbox)
	}
	smtp, err := mapEmailChannel(mailer.ChannelSMTP)
	if err != nil || smtp != mailer.ChannelSMTP {
		t.Fatalf("smtp: %v %s", err, smtp)
	}
	resend, err := mapEmailChannel(mailer.ChannelResend)
	if err != nil || resend != mailer.ChannelResend {
		t.Fatalf("resend: %v %s", err, resend)
	}
	none, err := mapEmailChannel("")
	if err != nil || none != "" {
		t.Fatalf("empty: %v %q", err, none)
	}
	if _, err := mapEmailChannel("pigeon"); err == nil {
		t.Fatal("expected error for unknown channel")
	}
}
