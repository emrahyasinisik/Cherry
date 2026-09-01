package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"
)

type session struct {
	userID    string
	expiresAt time.Time
}

type Memory struct {
	mu       sync.RWMutex
	users    map[string]User
	sessions map[string]session
}

func NewMemory() *Memory {
	return &Memory{
		users:    make(map[string]User),
		sessions: make(map[string]session),
	}
}

func (m *Memory) Name() string { return "memory" }

func (m *Memory) Ping(context.Context) error { return nil }

func (m *Memory) Login(_ context.Context, email, password string) (string, User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") {
		return "", User{}, ErrValidation
	}
	if strings.TrimSpace(password) == "" {
		return "", User{}, ErrValidation
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	user, ok := m.users[email]
	if !ok {
		user = User{
			ID:            newID(),
			Email:         email,
			WorkspaceKind: WorkspacePersonal,
		}
		m.users[email] = user
	}

	token := newID()
	m.sessions[token] = session{
		userID:    user.ID,
		expiresAt: time.Now().Add(12 * time.Hour),
	}
	return token, user, nil
}

func (m *Memory) Logout(_ context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, token)
	return nil
}

func (m *Memory) Me(_ context.Context, token string) (*User, error) {
	user, err := m.userForToken(token)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *Memory) Projects(_ context.Context, token string) ([]Project, error) {
	if _, err := m.userForToken(token); err != nil {
		return nil, err
	}
	return []Project{}, nil
}

func (m *Memory) userForToken(token string) (User, error) {
	if token == "" {
		return User{}, ErrUnauthorized
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	sess, ok := m.sessions[token]
	if !ok || time.Now().After(sess.expiresAt) {
		return User{}, ErrUnauthorized
	}
	for _, user := range m.users {
		if user.ID == sess.userID {
			return user, nil
		}
	}
	return User{}, ErrUnauthorized
}

func newID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
