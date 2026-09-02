package graph

import (
	"testing"

	"github.com/icerde/api/internal/activate"
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

func TestMapProjectEnumsExhaustive(t *testing.T) {
	if _, err := mapProjectStack(store.StackExpo); err != nil {
		t.Fatal(err)
	}
	if _, err := mapProjectStack(store.StackFlutter); err != nil {
		t.Fatal(err)
	}
	if _, err := mapProjectStack(store.StackNative); err != nil {
		t.Fatal(err)
	}
	if _, err := mapProjectStack(store.ProjectStack("X")); err == nil {
		t.Fatal("expected stack error")
	}
	for _, status := range []store.ProjectStatus{
		store.StatusQueued, store.StatusWriting, store.StatusTesting, store.StatusReady, store.StatusFailed,
	} {
		if _, err := mapProjectStatus(status); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := mapProjectStatus(store.ProjectStatus("X")); err == nil {
		t.Fatal("expected status error")
	}
	for _, result := range []store.MaestroResult{store.MaestroSkipped, store.MaestroPassed, store.MaestroFailed} {
		if _, err := mapMaestroResult(result); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := mapMaestroResult(store.MaestroResult("X")); err == nil {
		t.Fatal("expected maestro error")
	}
	for _, role := range []store.ChatRole{store.RoleUser, store.RoleAgent, store.RoleSystem, ""} {
		if _, err := mapChatRole(role); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := mapChatRole(store.ChatRole("X")); err == nil {
		t.Fatal("expected chat role error")
	}
	for _, status := range []activate.Status{
		activate.StatusIdle, activate.StatusStarting, activate.StatusRunning, activate.StatusStopping, activate.StatusFailed,
	} {
		if _, err := mapActivateStatus(status); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := mapActivateStatus(activate.Status("X")); err == nil {
		t.Fatal("expected activate error")
	}
	for _, target := range []store.BackendTarget{
		store.TargetLocal, store.TargetSupabase, store.TargetCloudflare, store.TargetRender, "",
	} {
		if _, err := mapBackendTarget(target); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := mapBackendTarget(store.BackendTarget("X")); err == nil {
		t.Fatal("expected backend error")
	}
	for _, kind := range []store.ConnectionKind{
		store.KindSupabase, store.KindCloudflare, store.KindGithub, store.KindVercel, store.KindRender,
	} {
		if _, err := mapConnectionKind(kind); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := mapConnectionKind(store.ConnectionKind("X")); err == nil {
		t.Fatal("expected kind error")
	}
	for _, status := range []store.ConnectionStatus{
		store.ConnDisconnected, store.ConnConnected, store.ConnFailed,
	} {
		if _, err := mapConnectionStatus(status); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := mapConnectionStatus(store.ConnectionStatus("X")); err == nil {
		t.Fatal("expected conn status error")
	}
	for _, method := range []store.ConnectionAuth{store.AuthNone, store.AuthOAuth, store.AuthToken, ""} {
		if _, err := mapConnectionAuth(method, store.ConnDisconnected); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := mapConnectionAuth(store.ConnectionAuth("X"), store.ConnDisconnected); err == nil {
		t.Fatal("expected auth error")
	}
	connectedToken, err := mapConnectionAuth("", store.ConnConnected)
	if err != nil || connectedToken != ConnectionAuthToken {
		t.Fatalf("%v %s", err, connectedToken)
	}
	for _, mode := range []string{"CONSENT", "PROVIDER"} {
		if _, err := mapOAuthMode(mode); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := mapOAuthMode("X"); err == nil {
		t.Fatal("expected mode error")
	}
}
