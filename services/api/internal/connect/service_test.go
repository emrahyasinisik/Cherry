package connect

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/icerde/api/internal/store"
)

func TestCatalogShowsAllDisconnected(t *testing.T) {
	svc := &Service{Store: store.NewMemory()}
	list, err := svc.Catalog(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 5 {
		t.Fatalf("len %d", len(list))
	}
	for _, conn := range list {
		if conn.Status != store.ConnDisconnected {
			t.Fatalf("%s %s", conn.Kind, conn.Status)
		}
		if conn.Token != "" {
			t.Fatal("token leaked")
		}
	}
}

func TestConnectRejectsEmptyToken(t *testing.T) {
	svc := &Service{Store: store.NewMemory()}
	_, err := svc.Connect(context.Background(), "u1", "GITHUB", "emrah", "short")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrConnect) {
		t.Fatalf("%v", err)
	}
}

func TestConnectAndDisconnect(t *testing.T) {
	svc := &Service{Store: store.NewMemory()}
	ctx := context.Background()
	got, err := svc.Connect(ctx, "u1", "GITHUB", "emrah", "ghp_12345678secret")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != store.ConnConnected || got.Token != "" || !strings.Contains(got.TokenHint, "cret") {
		t.Fatalf("%+v", got)
	}
	list, err := svc.Catalog(ctx, "u1")
	if err != nil {
		t.Fatal(err)
	}
	var github store.Connection
	for _, conn := range list {
		if conn.Kind == store.KindGithub {
			github = conn
		}
	}
	if github.Status != store.ConnConnected || github.Token != "" {
		t.Fatalf("%+v", github)
	}
	off, err := svc.Disconnect(ctx, "u1", "GITHUB")
	if err != nil {
		t.Fatal(err)
	}
	if off.Status != store.ConnDisconnected {
		t.Fatalf("%s", off.Status)
	}
}

func TestPushRequiresGithub(t *testing.T) {
	svc := &Service{Store: store.NewMemory(), Git: &fakeGit{}}
	_, err := svc.PushGitHub(context.Background(), "u1", "/tmp", "emrah/kahve")
	if err == nil {
		t.Fatal("expected")
	}
}

func TestPushCallsGit(t *testing.T) {
	mem := store.NewMemory()
	fake := &fakeGit{}
	svc := &Service{Store: mem, Git: fake}
	ctx := context.Background()
	if _, err := svc.Connect(ctx, "u1", "GITHUB", "emrah", "ghp_12345678secret"); err != nil {
		t.Fatal(err)
	}
	res, err := svc.PushGitHub(ctx, "u1", "/projects/x", "emrah/kahve")
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || fake.repo != "emrah/kahve" || fake.token != "ghp_12345678secret" {
		t.Fatalf("%+v %+v", res, fake)
	}
}

func TestParseKindExhaustive(t *testing.T) {
	for _, kind := range kinds() {
		if _, err := parseKind(string(kind)); err != nil {
			t.Fatal(err)
		}
		if catalogNote(kind) == "" {
			t.Fatalf("note %s", kind)
		}
	}
	if _, err := parseKind("TWILIO"); err == nil {
		t.Fatal("expected")
	}
}

type fakeGit struct {
	dir   string
	repo  string
	token string
}

func (f *fakeGit) Push(dir, repo, token string) error {
	f.dir = dir
	f.repo = repo
	f.token = token
	return nil
}
