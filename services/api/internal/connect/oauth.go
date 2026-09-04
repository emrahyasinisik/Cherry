package connect

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cherry/api/internal/store"
)

const (
	oauthModeConsent  = "CONSENT"
	oauthModeProvider = "PROVIDER"
	grantPrefix       = "cherry_grant_"
	pendingTTL        = 10 * time.Minute
)

type OAuthStart struct {
	AuthorizeURL string
	State        string
	Mode         string
}

type OAuthClient struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserURL      string
	Scopes       []string
}

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type pendingAuth struct {
	State     string
	UserID    string
	Account   string
	Kind      store.ConnectionKind
	Code      string
	Approved  bool
	ExpiresAt time.Time
}

func (s *Service) ensureOAuth() {
	if s.pending == nil {
		s.pending = &sync.Map{}
	}
	if s.HTTP == nil {
		s.HTTP = http.DefaultClient
	}
}

// StartOAuth begins an OAuth 2.0 authorization-code request.
// With provider client credentials this redirects to GitHub/Vercel/etc.
// Otherwise it uses the local consent screen (izin ekranı).
func (s *Service) StartOAuth(_ context.Context, userID, account string, kindRaw string) (OAuthStart, error) {
	s.ensureOAuth()
	kind, err := parseKind(kindRaw)
	if err != nil {
		return OAuthStart{}, err
	}
	account = strings.TrimSpace(account)
	if account == "" {
		return OAuthStart{}, wrap("Oturum hesabı yok.")
	}
	state, err := randomHex(24)
	if err != nil {
		return OAuthStart{}, wrap("OAuth state üretilemedi.")
	}
	s.pending.Store(state, &pendingAuth{
		State:     state,
		UserID:    userID,
		Account:   account,
		Kind:      kind,
		ExpiresAt: time.Now().UTC().Add(pendingTTL),
	})
	if client, ok := s.clientFor(kind); ok {
		redirect := strings.TrimRight(s.APIOrigin, "/") + "/oauth/provider/callback"
		values := url.Values{
			"client_id":     {client.ClientID},
			"redirect_uri":  {redirect},
			"state":         {state},
			"response_type": {"code"},
			"scope":         {strings.Join(client.Scopes, " ")},
		}
		return OAuthStart{
			AuthorizeURL: client.AuthURL + "?" + values.Encode(),
			State:        state,
			Mode:         oauthModeProvider,
		}, nil
	}
	values := url.Values{
		"client_id":     {"cherry"},
		"response_type": {"code"},
		"state":         {state},
		"kind":          {string(kind)},
		"scope":         {strings.Join(catalogScopes(kind), " ")},
	}
	return OAuthStart{
		AuthorizeURL: strings.TrimRight(s.WebOrigin, "/") + "/oauth/authorize?" + values.Encode(),
		State:        state,
		Mode:         oauthModeConsent,
	}, nil
}

func (s *Service) CompleteOAuth(ctx context.Context, userID, code, state string) (store.Connection, error) {
	s.ensureOAuth()
	code = strings.TrimSpace(code)
	state = strings.TrimSpace(state)
	if code == "" || state == "" {
		return store.Connection{}, wrap("OAuth code/state eksik.")
	}
	pending, err := s.takePending(state)
	if err != nil {
		return store.Connection{}, err
	}
	if pending.UserID != userID {
		return store.Connection{}, wrap("OAuth oturumu bu hesaba ait değil.")
	}
	if !pending.Approved || pending.Code == "" || pending.Code != code {
		return store.Connection{}, wrap("OAuth izni tamamlanmadı veya kod geçersiz.")
	}
	token, err := randomHex(20)
	if err != nil {
		return store.Connection{}, wrap("OAuth grant üretilemedi.")
	}
	return s.saveOAuth(ctx, userID, pending.Kind, pending.Account, grantPrefix+token, catalogScopes(pending.Kind))
}

func (s *Service) HandleDecision(w http.ResponseWriter, r *http.Request) {
	s.ensureOAuth()
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	state := strings.TrimSpace(firstForm(r, "state"))
	decision := strings.ToLower(strings.TrimSpace(firstForm(r, "decision")))
	account := strings.TrimSpace(firstForm(r, "account"))
	web := strings.TrimRight(s.WebOrigin, "/")
	if web == "" {
		web = "http://127.0.0.1:43147"
	}
	pending, err := s.peekPending(state)
	if err != nil {
		http.Redirect(w, r, web+"/connections?oauth=expired", http.StatusFound)
		return
	}
	if decision == "deny" || decision == "cancel" {
		s.pending.Delete(state)
		http.Redirect(w, r, web+"/connections?oauth=denied&kind="+url.QueryEscape(string(pending.Kind)), http.StatusFound)
		return
	}
	if decision != "allow" && decision != "approve" {
		http.Redirect(w, r, web+"/connections?oauth=denied&kind="+url.QueryEscape(string(pending.Kind)), http.StatusFound)
		return
	}
	if account == "" {
		account = pending.Account
	}
	if len(account) < 2 {
		http.Redirect(w, r, web+"/oauth/authorize?"+url.Values{
			"kind":          {string(pending.Kind)},
			"state":         {state},
			"client_id":     {"cherry"},
			"response_type": {"code"},
			"error":         {"account"},
		}.Encode(), http.StatusFound)
		return
	}
	code, err := randomHex(20)
	if err != nil {
		http.Redirect(w, r, web+"/connections?oauth=error", http.StatusFound)
		return
	}
	pending.Account = account
	pending.Code = code
	pending.Approved = true
	s.pending.Store(state, pending)
	http.Redirect(w, r, web+"/oauth/callback?"+url.Values{
		"code":  {code},
		"state": {state},
	}.Encode(), http.StatusFound)
}

func (s *Service) HandleProviderCallback(w http.ResponseWriter, r *http.Request) {
	s.ensureOAuth()
	web := strings.TrimRight(s.WebOrigin, "/")
	if web == "" {
		web = "http://127.0.0.1:43147"
	}
	q := r.URL.Query()
	if errParam := strings.TrimSpace(q.Get("error")); errParam != "" {
		http.Redirect(w, r, web+"/connections?oauth=denied", http.StatusFound)
		return
	}
	state := strings.TrimSpace(q.Get("state"))
	code := strings.TrimSpace(q.Get("code"))
	pending, err := s.takePending(state)
	if err != nil || code == "" {
		http.Redirect(w, r, web+"/connections?oauth=expired", http.StatusFound)
		return
	}
	client, ok := s.clientFor(pending.Kind)
	if !ok {
		http.Redirect(w, r, web+"/connections?oauth=error", http.StatusFound)
		return
	}
	token, err := s.exchangeCode(client, code)
	if err != nil {
		http.Redirect(w, r, web+"/connections?oauth=error", http.StatusFound)
		return
	}
	account, err := s.fetchAccount(client, token)
	if err != nil || account == "" {
		account = pending.Account
	}
	if _, err := s.saveOAuth(r.Context(), pending.UserID, pending.Kind, account, token, client.Scopes); err != nil {
		http.Redirect(w, r, web+"/connections?oauth=error", http.StatusFound)
		return
	}
	http.Redirect(w, r, web+"/connections?oauth=connected&kind="+url.QueryEscape(string(pending.Kind)), http.StatusFound)
}

func (s *Service) saveOAuth(ctx context.Context, userID string, kind store.ConnectionKind, account, token string, scopes []string) (store.Connection, error) {
	account = strings.TrimSpace(account)
	token = strings.TrimSpace(token)
	if len(account) < 2 || len(token) < 8 {
		return store.Connection{}, wrap("OAuth hesabı veya jetonu eksik.")
	}
	now := time.Now().UTC()
	conn := store.Connection{
		UserID:     userID,
		Kind:       kind,
		Status:     store.ConnConnected,
		Account:    account,
		Token:      token,
		TokenHint:  hint(token),
		Note:       oauthNote(kind, strings.HasPrefix(token, grantPrefix)),
		AuthMethod: store.AuthOAuth,
		Scopes:     append([]string{}, scopes...),
		UpdatedAt:  now,
	}
	existing, err := s.Store.GetConnection(ctx, userID, kind)
	if err == nil {
		conn.ID = existing.ID
	}
	saved, err := s.Store.UpsertConnection(ctx, conn)
	if err != nil {
		return store.Connection{}, err
	}
	saved.Token = ""
	return saved, nil
}

func (s *Service) clientFor(kind store.ConnectionKind) (OAuthClient, bool) {
	if s.Clients == nil {
		return OAuthClient{}, false
	}
	client, ok := s.Clients[kind]
	if !ok || strings.TrimSpace(client.ClientID) == "" || strings.TrimSpace(client.AuthURL) == "" {
		return OAuthClient{}, false
	}
	return client, true
}

func (s *Service) peekPending(state string) (*pendingAuth, error) {
	if strings.TrimSpace(state) == "" {
		return nil, wrap("OAuth state eksik.")
	}
	raw, ok := s.pending.Load(state)
	if !ok {
		return nil, wrap("OAuth isteği bulunamadı veya süresi doldu.")
	}
	pending, ok := raw.(*pendingAuth)
	if !ok || pending == nil || time.Now().UTC().After(pending.ExpiresAt) {
		s.pending.Delete(state)
		return nil, wrap("OAuth isteği bulunamadı veya süresi doldu.")
	}
	return pending, nil
}

func (s *Service) takePending(state string) (*pendingAuth, error) {
	pending, err := s.peekPending(state)
	if err != nil {
		return nil, err
	}
	s.pending.Delete(state)
	return pending, nil
}

func (s *Service) exchangeCode(client OAuthClient, code string) (string, error) {
	redirect := strings.TrimRight(s.APIOrigin, "/") + "/oauth/provider/callback"
	form := url.Values{
		"client_id":     {client.ClientID},
		"client_secret": {client.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirect},
	}
	req, err := http.NewRequest(http.MethodPost, client.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	res, err := s.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("token status %d", res.StatusCode)
	}
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return "", wrap("Sağlayıcı access_token vermedi.")
	}
	return payload.AccessToken, nil
}

func (s *Service) fetchAccount(client OAuthClient, token string) (string, error) {
	if strings.TrimSpace(client.UserURL) == "" {
		return "", wrap("user url yok")
	}
	req, err := http.NewRequest(http.MethodGet, client.UserURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	res, err := s.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("user status %d", res.StatusCode)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	for _, key := range []string{"login", "username", "name", "email", "id"} {
		if value, ok := payload[key]; ok {
			text := strings.TrimSpace(fmt.Sprint(value))
			if text != "" && text != "<nil>" {
				return text, nil
			}
		}
	}
	if user, ok := payload["user"].(map[string]any); ok {
		for _, key := range []string{"username", "email", "name"} {
			if value, ok := user[key]; ok {
				text := strings.TrimSpace(fmt.Sprint(value))
				if text != "" {
					return text, nil
				}
			}
		}
	}
	return "", wrap("hesap adı yok")
}

func LoadClientsFromEnv() map[store.ConnectionKind]OAuthClient {
	out := map[store.ConnectionKind]OAuthClient{}
	if id := strings.TrimSpace(os.Getenv("CHERRY_GITHUB_CLIENT_ID")); id != "" {
		out[store.KindGithub] = OAuthClient{
			ClientID:     id,
			ClientSecret: strings.TrimSpace(os.Getenv("CHERRY_GITHUB_CLIENT_SECRET")),
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
			UserURL:      "https://api.github.com/user",
			Scopes:       catalogScopes(store.KindGithub),
		}
	}
	if id := strings.TrimSpace(os.Getenv("CHERRY_VERCEL_CLIENT_ID")); id != "" {
		out[store.KindVercel] = OAuthClient{
			ClientID:     id,
			ClientSecret: strings.TrimSpace(os.Getenv("CHERRY_VERCEL_CLIENT_SECRET")),
			AuthURL:      getenv("CHERRY_VERCEL_AUTH_URL", "https://vercel.com/oauth/authorize"),
			TokenURL:     getenv("CHERRY_VERCEL_TOKEN_URL", "https://api.vercel.com/v2/oauth/access_token"),
			UserURL:      "https://api.vercel.com/v2/user",
			Scopes:       catalogScopes(store.KindVercel),
		}
	}
	if id := strings.TrimSpace(os.Getenv("CHERRY_SUPABASE_CLIENT_ID")); id != "" {
		out[store.KindSupabase] = OAuthClient{
			ClientID:     id,
			ClientSecret: strings.TrimSpace(os.Getenv("CHERRY_SUPABASE_CLIENT_SECRET")),
			AuthURL:      getenv("CHERRY_SUPABASE_AUTH_URL", "https://api.supabase.com/v1/oauth/authorize"),
			TokenURL:     getenv("CHERRY_SUPABASE_TOKEN_URL", "https://api.supabase.com/v1/oauth/token"),
			UserURL:      getenv("CHERRY_SUPABASE_USER_URL", "https://api.supabase.com/v1/organizations"),
			Scopes:       catalogScopes(store.KindSupabase),
		}
	}
	// Cloudflare / Render: set CLIENT_ID (+ secret) and Auth/Token URLs for your OAuth app.
	// Without AUTH_URL Cherry keeps the local consent screen; token paste still works.
	if id := strings.TrimSpace(os.Getenv("CHERRY_CLOUDFLARE_CLIENT_ID")); id != "" {
		authURL := strings.TrimSpace(os.Getenv("CHERRY_CLOUDFLARE_AUTH_URL"))
		tokenURL := strings.TrimSpace(os.Getenv("CHERRY_CLOUDFLARE_TOKEN_URL"))
		if authURL != "" && tokenURL != "" {
			out[store.KindCloudflare] = OAuthClient{
				ClientID:     id,
				ClientSecret: strings.TrimSpace(os.Getenv("CHERRY_CLOUDFLARE_CLIENT_SECRET")),
				AuthURL:      authURL,
				TokenURL:     tokenURL,
				UserURL:      getenv("CHERRY_CLOUDFLARE_USER_URL", "https://api.cloudflare.com/client/v4/user"),
				Scopes:       catalogScopes(store.KindCloudflare),
			}
		}
	}
	if id := strings.TrimSpace(os.Getenv("CHERRY_RENDER_CLIENT_ID")); id != "" {
		out[store.KindRender] = OAuthClient{
			ClientID:     id,
			ClientSecret: strings.TrimSpace(os.Getenv("CHERRY_RENDER_CLIENT_SECRET")),
			AuthURL:      getenv("CHERRY_RENDER_AUTH_URL", "https://api.render.com/oauth/authorize"),
			TokenURL:     getenv("CHERRY_RENDER_TOKEN_URL", "https://api.render.com/oauth/token"),
			UserURL:      getenv("CHERRY_RENDER_USER_URL", "https://api.render.com/v1/owners"),
			Scopes:       catalogScopes(store.KindRender),
		}
	}
	return out
}

// ModeFor reports whether Bağlan will open the provider site or the local consent page.
func (s *Service) ModeFor(kind store.ConnectionKind) string {
	s.ensureOAuth()
	if _, ok := s.clientFor(kind); ok {
		return oauthModeProvider
	}
	return oauthModeConsent
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func firstForm(r *http.Request, key string) string {
	if v := r.FormValue(key); v != "" {
		return v
	}
	return r.URL.Query().Get(key)
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func oauthNote(kind store.ConnectionKind, localGrant bool) string {
	if localGrant {
		return "OAuth 2.0 izin ekranı onaylandı. Cherry host değil. Yerel grant; sağlayıcı client id yoksa gerçek API çağrısı için token gerekir."
	}
	_ = kind
	return "OAuth 2.0 ile bağlandı. Cherry host değil. Token GraphQL’e dönmez."
}

func catalogScopes(kind store.ConnectionKind) []string {
	switch kind {
	case store.KindGithub:
		return []string{"repo", "read:user", "user:email"}
	case store.KindVercel:
		return []string{"read:user", "deploy"}
	case store.KindSupabase:
		return []string{"organizations:read", "projects:read"}
	case store.KindCloudflare:
		return []string{"workers:edit", "d1:edit", "r2:edit"}
	case store.KindRender:
		return []string{"services:write"}
	default:
		return nil
	}
}
