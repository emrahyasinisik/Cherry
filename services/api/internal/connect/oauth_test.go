package connect

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/icerde/api/internal/store"
)

func TestOAuthConsentThenComplete(t *testing.T) {
	mem := store.NewMemory()
	svc := &Service{
		Store:     mem,
		WebOrigin: "http://127.0.0.1:43147",
		APIOrigin: "http://127.0.0.1:43148",
	}
	ctx := context.Background()
	start, err := svc.StartOAuth(ctx, "u1", "emrah@icerde.dev", "GITHUB")
	if err != nil {
		t.Fatal(err)
	}
	if start.Mode != oauthModeConsent {
		t.Fatalf("mode %s", start.Mode)
	}
	if !strings.Contains(start.AuthorizeURL, "/oauth/authorize") || !strings.Contains(start.AuthorizeURL, "state=") {
		t.Fatalf("url %s", start.AuthorizeURL)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/oauth/decision?state="+url.QueryEscape(start.State)+"&decision=allow&account=emrah", nil)
	svc.HandleDecision(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("status %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	parsed, err := url.Parse(loc)
	if err != nil {
		t.Fatal(err)
	}
	code := parsed.Query().Get("code")
	if code == "" || parsed.Query().Get("state") != start.State {
		t.Fatalf("callback %s", loc)
	}

	got, err := svc.CompleteOAuth(ctx, "u1", code, start.State)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != store.ConnConnected || got.AuthMethod != store.AuthOAuth || got.Token != "" {
		t.Fatalf("%+v", got)
	}
	if got.Account != "emrah" {
		t.Fatalf("account %s", got.Account)
	}

	push, err := svc.PushGitHub(ctx, "u1", "/tmp", "emrah/kahve")
	if err != nil {
		t.Fatal(err)
	}
	if push.OK {
		t.Fatal("local grant must not fake github push")
	}
	if !strings.Contains(push.Note, "OAuth") && !strings.Contains(push.Note, "PAT") {
		t.Fatalf("note %s", push.Note)
	}
}

func TestOAuthDenyDoesNotConnect(t *testing.T) {
	svc := &Service{
		Store:     store.NewMemory(),
		WebOrigin: "http://127.0.0.1:43147",
	}
	start, err := svc.StartOAuth(context.Background(), "u1", "emrah@icerde.dev", "VERCEL")
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/oauth/decision?state="+url.QueryEscape(start.State)+"&decision=deny", nil)
	svc.HandleDecision(rec, req)
	if rec.Code != http.StatusFound || !strings.Contains(rec.Header().Get("Location"), "denied") {
		t.Fatalf("%d %s", rec.Code, rec.Header().Get("Location"))
	}
	_, err = svc.CompleteOAuth(context.Background(), "u1", "nope", start.State)
	if err == nil {
		t.Fatal("expected")
	}
}

func TestOAuthWrongUser(t *testing.T) {
	svc := &Service{
		Store:     store.NewMemory(),
		WebOrigin: "http://127.0.0.1:43147",
	}
	ctx := context.Background()
	start, err := svc.StartOAuth(ctx, "u1", "emrah@icerde.dev", "SUPABASE")
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	svc.HandleDecision(rec, httptest.NewRequest(http.MethodGet, "/oauth/decision?state="+start.State+"&decision=allow&account=org-1", nil))
	loc, _ := url.Parse(rec.Header().Get("Location"))
	_, err = svc.CompleteOAuth(ctx, "u2", loc.Query().Get("code"), start.State)
	if err == nil {
		t.Fatal("expected")
	}
}

func TestProviderOAuthStartUsesGithub(t *testing.T) {
	svc := &Service{
		Store:     store.NewMemory(),
		WebOrigin: "http://127.0.0.1:43147",
		APIOrigin: "http://127.0.0.1:43148",
		Clients: map[store.ConnectionKind]OAuthClient{
			store.KindGithub: {
				ClientID: "cid",
				AuthURL:  "https://github.com/login/oauth/authorize",
				Scopes:   catalogScopes(store.KindGithub),
			},
		},
	}
	start, err := svc.StartOAuth(context.Background(), "u1", "emrah@icerde.dev", "GITHUB")
	if err != nil {
		t.Fatal(err)
	}
	if start.Mode != oauthModeProvider {
		t.Fatalf("mode %s", start.Mode)
	}
	if !strings.Contains(start.AuthorizeURL, "github.com/login/oauth/authorize") {
		t.Fatalf("url %s", start.AuthorizeURL)
	}
}

func TestProviderCallbackStoresToken(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "gho_live_token_ok"})
			return
		}
		if r.URL.Path == "/user" {
			_ = json.NewEncoder(w).Encode(map[string]string{"login": "emrah"})
			return
		}
		w.WriteHeader(404)
	}))
	defer tokenSrv.Close()

	mem := store.NewMemory()
	svc := &Service{
		Store:     mem,
		WebOrigin: "http://web.test",
		APIOrigin: "http://api.test",
		HTTP:      tokenSrv.Client(),
		Clients: map[store.ConnectionKind]OAuthClient{
			store.KindGithub: {
				ClientID:     "cid",
				ClientSecret: "sec",
				AuthURL:      tokenSrv.URL + "/auth",
				TokenURL:     tokenSrv.URL + "/token",
				UserURL:      tokenSrv.URL + "/user",
				Scopes:       catalogScopes(store.KindGithub),
			},
		},
	}
	start, err := svc.StartOAuth(context.Background(), "u1", "emrah@icerde.dev", "GITHUB")
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/oauth/provider/callback?code=abc&state="+start.State, nil)
	svc.HandleProviderCallback(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Location"), "connected") {
		t.Fatalf("loc %s", rec.Header().Get("Location"))
	}
	conn, err := mem.GetConnection(context.Background(), "u1", store.KindGithub)
	if err != nil {
		t.Fatal(err)
	}
	if conn.Token != "gho_live_token_ok" || conn.Account != "emrah" || conn.AuthMethod != store.AuthOAuth {
		t.Fatalf("%+v", conn)
	}
}

func TestCatalogScopesExhaustive(t *testing.T) {
	for _, kind := range kinds() {
		if len(catalogScopes(kind)) == 0 {
			t.Fatalf("scopes %s", kind)
		}
	}
}

func TestExchangeRejectsEmptyToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"access_token":""}`)
	}))
	defer srv.Close()
	svc := &Service{HTTP: srv.Client(), APIOrigin: "http://api.test"}
	_, err := svc.exchangeCode(OAuthClient{TokenURL: srv.URL, ClientID: "x"}, "code")
	if err == nil {
		t.Fatal("expected")
	}
}
