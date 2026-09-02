package store

import (
	"context"
	"strings"
	"sync"
)

type Memory struct {
	mu         sync.RWMutex
	users      map[string]User
	usersByID  map[string]User
	challenges map[string]Challenge
	sessions   map[string]Session
	devices    map[string]Device
	mail       map[string]Mail
	projects   map[string]Project
	logs       map[string][]JobLog
	versions   map[string]LlmVersion
	audits     []AuditEvent
	llmState   LlmState
}

func NewMemory() *Memory {
	return &Memory{
		users:      make(map[string]User),
		usersByID:  make(map[string]User),
		challenges: make(map[string]Challenge),
		sessions:   make(map[string]Session),
		devices:    make(map[string]Device),
		mail:       make(map[string]Mail),
		projects:   make(map[string]Project),
		logs:       make(map[string][]JobLog),
		versions:   make(map[string]LlmVersion),
		audits:     make([]AuditEvent, 0),
	}
}

func (m *Memory) Name() string { return "memory" }

func (m *Memory) Ping(context.Context) error { return nil }

func (m *Memory) CreateUser(_ context.Context, user User) (User, error) {
	email := strings.ToLower(strings.TrimSpace(user.Email))
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[email]; ok {
		return User{}, ErrExists
	}
	if user.ID == "" {
		user.ID = NewID()
	}
	user.Email = email
	m.users[email] = user
	m.usersByID[user.ID] = user
	return user, nil
}

func (m *Memory) GetUserByEmail(_ context.Context, email string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	user, ok := m.users[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return nil, ErrNotFound
	}
	copy := user
	return &copy, nil
}

func (m *Memory) GetUserByID(_ context.Context, id string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	user, ok := m.usersByID[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := user
	return &copy, nil
}

func (m *Memory) UpdateUser(_ context.Context, user User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.usersByID[user.ID]
	if !ok {
		return ErrNotFound
	}
	delete(m.users, existing.Email)
	m.users[user.Email] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *Memory) PutChallenge(_ context.Context, challenge Challenge) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.challenges[challenge.ID] = challenge
	return nil
}

func (m *Memory) GetChallenge(_ context.Context, id string) (*Challenge, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	challenge, ok := m.challenges[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := challenge
	return &copy, nil
}

func (m *Memory) GetChallengeByLinkHash(_ context.Context, linkHash string) (*Challenge, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, challenge := range m.challenges {
		if challenge.LinkHash == linkHash && !challenge.Consumed {
			copy := challenge
			return &copy, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) InvalidateChallenges(_ context.Context, userID string, purpose Purpose) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, challenge := range m.challenges {
		if challenge.UserID == userID && challenge.Purpose == purpose && !challenge.Consumed {
			challenge.Consumed = true
			m.challenges[id] = challenge
		}
	}
	return nil
}

func (m *Memory) CreateSession(_ context.Context, session Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
	return nil
}

func (m *Memory) GetSessionByTokenHash(_ context.Context, tokenHash string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, session := range m.sessions {
		if session.TokenHash == tokenHash && !session.Revoked {
			copy := session
			return &copy, nil
		}
	}
	return nil, ErrUnauthorized
}

func (m *Memory) ListSessions(_ context.Context, userID string) ([]Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Session, 0)
	for _, session := range m.sessions {
		if session.UserID == userID && !session.Revoked {
			out = append(out, session)
		}
	}
	return out, nil
}

func (m *Memory) RevokeSession(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[id]
	if !ok {
		return ErrNotFound
	}
	session.Revoked = true
	m.sessions[id] = session
	return nil
}

func (m *Memory) RevokeOtherSessions(_ context.Context, userID, keepID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, session := range m.sessions {
		if session.UserID == userID && session.ID != keepID {
			session.Revoked = true
			m.sessions[id] = session
		}
	}
	return nil
}

func (m *Memory) UpsertDevice(_ context.Context, device Device) (Device, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, existing := range m.devices {
		if existing.UserID == device.UserID && existing.FPHash == device.FPHash {
			existing.Label = device.Label
			existing.Trusted = device.Trusted
			existing.LastSeen = device.LastSeen
			m.devices[id] = existing
			return existing, nil
		}
	}
	if device.ID == "" {
		device.ID = NewID()
	}
	m.devices[device.ID] = device
	return device, nil
}

func (m *Memory) GetDeviceByFP(_ context.Context, userID, fpHash string) (*Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, device := range m.devices {
		if device.UserID == userID && device.FPHash == fpHash {
			copy := device
			return &copy, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) ListDevices(_ context.Context, userID string) ([]Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Device, 0)
	for _, device := range m.devices {
		if device.UserID == userID {
			out = append(out, device)
		}
	}
	return out, nil
}

func (m *Memory) RevokeDevice(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	device, ok := m.devices[id]
	if !ok {
		return ErrNotFound
	}
	device.Trusted = false
	m.devices[id] = device
	return nil
}

func (m *Memory) AddMail(_ context.Context, mail Mail) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if mail.ID == "" {
		mail.ID = NewID()
	}
	m.mail[mail.ID] = mail
	return nil
}

func (m *Memory) ListMail(_ context.Context, userID string) ([]Mail, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Mail, 0)
	for _, mail := range m.mail {
		if mail.UserID == userID {
			out = append(out, mail)
		}
	}
	return out, nil
}

func (m *Memory) MailByChallenge(_ context.Context, challengeID string) (*Mail, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var latest *Mail
	for _, mail := range m.mail {
		if mail.ChallengeID == challengeID {
			copy := mail
			if latest == nil || copy.CreatedAt.After(latest.CreatedAt) {
				latest = &copy
			}
		}
	}
	if latest == nil {
		return nil, ErrNotFound
	}
	return latest, nil
}

func (m *Memory) Projects(ctx context.Context, userID string) ([]Project, error) {
	return m.ListProjects(ctx, userID)
}

func (m *Memory) ListProjects(_ context.Context, userID string) ([]Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Project, 0)
	for _, project := range m.projects {
		if project.UserID == userID {
			out = append(out, project)
		}
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].CreatedAt.After(out[i].CreatedAt) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out, nil
}

func (m *Memory) CreateProject(_ context.Context, project Project) (Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if project.ID == "" {
		project.ID = NewID()
	}
	m.projects[project.ID] = project
	return project, nil
}

func (m *Memory) GetProject(_ context.Context, id string) (*Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	project, ok := m.projects[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := project
	return &copy, nil
}

func (m *Memory) UpdateProject(_ context.Context, project Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.projects[project.ID]; !ok {
		return ErrNotFound
	}
	m.projects[project.ID] = project
	return nil
}

func (m *Memory) AppendLog(_ context.Context, log JobLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if log.ID == "" {
		log.ID = NewID()
	}
	m.logs[log.ProjectID] = append(m.logs[log.ProjectID], log)
	return nil
}

func (m *Memory) ListLogs(_ context.Context, projectID string) ([]JobLog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	src := m.logs[projectID]
	out := make([]JobLog, len(src))
	copy(out, src)
	return out, nil
}

func (m *Memory) PutLlmVersion(_ context.Context, version LlmVersion) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if version.ID == "" {
		version.ID = NewID()
	}
	m.versions[version.ID] = version
	return nil
}

func (m *Memory) ListLlmVersions(_ context.Context, slot LlmSlot) ([]LlmVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]LlmVersion, 0)
	for _, version := range m.versions {
		if version.Slot == slot {
			out = append(out, version)
		}
	}
	return out, nil
}

func (m *Memory) GetLlmVersion(_ context.Context, id string) (*LlmVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	version, ok := m.versions[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := version
	return &copy, nil
}

func (m *Memory) GetLlmState(_ context.Context) (LlmState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.llmState, nil
}

func (m *Memory) SetLlmState(_ context.Context, state LlmState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.llmState = state
	return nil
}

func (m *Memory) AddAudit(_ context.Context, event AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if event.ID == "" {
		event.ID = NewID()
	}
	m.audits = append(m.audits, event)
	return nil
}

func (m *Memory) ListAudit(_ context.Context, userID string) ([]AuditEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]AuditEvent, 0)
	for i := len(m.audits) - 1; i >= 0; i-- {
		if m.audits[i].UserID == userID {
			out = append(out, m.audits[i])
		}
	}
	return out, nil
}

func (m *Memory) DeleteUserData(_ context.Context, userID string, wipeProjects bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	user, ok := m.usersByID[userID]
	if !ok {
		return ErrNotFound
	}
	delete(m.users, user.Email)
	delete(m.usersByID, userID)
	for id, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, id)
		}
	}
	for id, device := range m.devices {
		if device.UserID == userID {
			delete(m.devices, id)
		}
	}
	for id, mail := range m.mail {
		if mail.UserID == userID {
			delete(m.mail, id)
		}
	}
	for id, challenge := range m.challenges {
		if challenge.UserID == userID {
			delete(m.challenges, id)
		}
	}
	kept := make([]AuditEvent, 0)
	for _, event := range m.audits {
		if event.UserID != userID {
			kept = append(kept, event)
		}
	}
	m.audits = kept
	if wipeProjects {
		for id, project := range m.projects {
			if project.UserID == userID {
				delete(m.projects, id)
				delete(m.logs, id)
			}
		}
	}
	return nil
}
