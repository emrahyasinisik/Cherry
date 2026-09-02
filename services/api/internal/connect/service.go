package connect

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/icerde/api/internal/store"
)

var ErrConnect = errors.New("bağlantı")

func wrap(msg string) error {
	return fmt.Errorf("%w: %s", ErrConnect, msg)
}

type GitPusher interface {
	Push(dir, repo, token string) error
}

type Service struct {
	Store     store.Store
	Git       GitPusher
	WebOrigin string
	APIOrigin string
	Clients   map[store.ConnectionKind]OAuthClient
	HTTP      HTTPDoer
	pending   *sync.Map
}

func (s *Service) Catalog(ctx context.Context, userID string) ([]store.Connection, error) {
	saved, err := s.Store.ListConnections(ctx, userID)
	if err != nil {
		return nil, err
	}
	byKind := map[store.ConnectionKind]store.Connection{}
	for _, conn := range saved {
		byKind[conn.Kind] = conn
	}
	out := make([]store.Connection, 0, len(kinds()))
	for _, kind := range kinds() {
		if conn, ok := byKind[kind]; ok {
			conn.Token = ""
			if len(conn.Scopes) == 0 {
				conn.Scopes = catalogScopes(kind)
			}
			out = append(out, conn)
			continue
		}
		out = append(out, store.Connection{
			Kind:       kind,
			Status:     store.ConnDisconnected,
			Note:       catalogNote(kind),
			Account:    "",
			AuthMethod: store.AuthNone,
			Scopes:     catalogScopes(kind),
		})
	}
	return out, nil
}

func (s *Service) Connect(ctx context.Context, userID, kindRaw, account, token string) (store.Connection, error) {
	kind, err := parseKind(kindRaw)
	if err != nil {
		return store.Connection{}, err
	}
	account = strings.TrimSpace(account)
	token = strings.TrimSpace(token)
	if len(account) < 2 {
		return store.Connection{}, wrap("Hesap adı gerekli (ör. github kullanıcı veya proje ref).")
	}
	if len(token) < 8 {
		return store.Connection{}, wrap("Token gerekli. Boş kayıt bağlı sayılmaz.")
	}
	now := time.Now().UTC()
	conn := store.Connection{
		UserID:     userID,
		Kind:       kind,
		Status:     store.ConnConnected,
		Account:    account,
		Token:      token,
		TokenHint:  hint(token),
		Note:       "Anahtar bu makinede duruyor. İçerde barındırmaz. Token GraphQL’e dönmez.",
		AuthMethod: store.AuthToken,
		Scopes:     catalogScopes(kind),
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

func (s *Service) Disconnect(ctx context.Context, userID, kindRaw string) (store.Connection, error) {
	kind, err := parseKind(kindRaw)
	if err != nil {
		return store.Connection{}, err
	}
	if err := s.Store.DeleteConnection(ctx, userID, kind); err != nil {
		return store.Connection{}, err
	}
	return store.Connection{
		Kind:       kind,
		Status:     store.ConnDisconnected,
		Note:       catalogNote(kind),
		AuthMethod: store.AuthNone,
		Scopes:     catalogScopes(kind),
	}, nil
}

func (s *Service) Require(ctx context.Context, userID string, kind store.ConnectionKind) (store.Connection, error) {
	conn, err := s.Store.GetConnection(ctx, userID, kind)
	if err != nil {
		return store.Connection{}, wrap("Bu sağlayıcı bağlı değil.")
	}
	if conn.Status != store.ConnConnected || strings.TrimSpace(conn.Token) == "" {
		return store.Connection{}, wrap("Bu sağlayıcı bağlı değil.")
	}
	return *conn, nil
}

type PushResult struct {
	OK   bool
	Note string
}

func (s *Service) PushGitHub(ctx context.Context, userID, projectRoot, repo string) (PushResult, error) {
	repo = strings.TrimSpace(repo)
	if !validRepo(repo) {
		return PushResult{}, wrap("Repo owner/ad biçiminde olmalı (ör. emrah/kahve).")
	}
	conn, err := s.Require(ctx, userID, store.KindGithub)
	if err != nil {
		return PushResult{}, err
	}
	if strings.HasPrefix(conn.Token, grantPrefix) {
		return PushResult{
			OK:   false,
			Note: "Yerel OAuth izni GitHub’a push etmez. GitHub OAuth uygulaması (ICERDE_GITHUB_CLIENT_ID) veya PAT gerekir.",
		}, nil
	}
	if s.Git == nil {
		return PushResult{OK: false, Note: "Git bağlı değil."}, nil
	}
	if err := s.Git.Push(projectRoot, repo, conn.Token); err != nil {
		return PushResult{OK: false, Note: err.Error()}, nil
	}
	return PushResult{OK: true, Note: "GitHub’a gönderildi: " + repo}, nil
}

func kinds() []store.ConnectionKind {
	return []store.ConnectionKind{
		store.KindSupabase,
		store.KindCloudflare,
		store.KindGithub,
		store.KindVercel,
		store.KindRender,
	}
}

func parseKind(raw string) (store.ConnectionKind, error) {
	kind := store.ConnectionKind(strings.TrimSpace(raw))
	switch kind {
	case store.KindSupabase, store.KindCloudflare, store.KindGithub, store.KindVercel, store.KindRender:
		return kind, nil
	default:
		return "", wrap("Bilinmeyen sağlayıcı.")
	}
}

func catalogNote(kind store.ConnectionKind) string {
	switch kind {
	case store.KindSupabase:
		return "OAuth 2.0. Kişinin Supabase projesi. Müşteri backend hedefi olabilir. İçerde host değil."
	case store.KindCloudflare:
		return "OAuth 2.0. Workers / D1 / R2. Kişinin hesabı."
	case store.KindGithub:
		return "OAuth 2.0. Geliştirilen projeyi kişinin reposuna push."
	case store.KindVercel:
		return "OAuth 2.0. Kişinin Vercel hesabına frontend deploy. İçerde host değil."
	case store.KindRender:
		return "OAuth 2.0. Kişinin Render hesabına servis. İçerde host değil."
	default:
		return ""
	}
}

func hint(token string) string {
	if len(token) <= 4 {
		return "••••"
	}
	return "…" + token[len(token)-4:]
}

func validRepo(repo string) bool {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return false
	}
	for _, part := range parts {
		if part == "" || strings.ContainsAny(part, " :@") {
			return false
		}
	}
	return true
}
