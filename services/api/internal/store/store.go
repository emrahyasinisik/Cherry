package store

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrValidation         = errors.New("validation")
	ErrNotFound           = errors.New("not found")
	ErrExists             = errors.New("exists")
	ErrLocked             = errors.New("locked")
	ErrExpired            = errors.New("expired")
	ErrMailFailed         = errors.New("mail failed")
	ErrLLMFailed          = errors.New("llm failed")
	ErrPath               = errors.New("path")
	ErrBusy               = errors.New("busy")
)

type WorkspaceKind string

const (
	WorkspacePersonal     WorkspaceKind = "PERSONAL"
	WorkspaceOrganization WorkspaceKind = "ORGANIZATION"
)

type Purpose string

const (
	PurposeNewDevice       Purpose = "NEW_DEVICE"
	PurposeLoginChallenge  Purpose = "LOGIN_CHALLENGE"
	PurposeEmailVerify     Purpose = "EMAIL_VERIFY"
	PurposeSuspiciousLogin Purpose = "SUSPICIOUS_LOGIN"
)

type User struct {
	ID            string
	Email         string
	PasswordHash  string
	WorkspaceKind WorkspaceKind
	TotpSecret    string
	TotpEnabled   bool
}

type ProjectStack string

const (
	StackExpo    ProjectStack = "EXPO"
	StackFlutter ProjectStack = "FLUTTER"
	StackNative  ProjectStack = "NATIVE"
)

type ProjectStatus string

const (
	StatusQueued  ProjectStatus = "QUEUED"
	StatusWriting ProjectStatus = "WRITING"
	StatusTesting ProjectStatus = "TESTING"
	StatusReady   ProjectStatus = "READY"
	StatusFailed  ProjectStatus = "FAILED"
)

type MaestroResult string

const (
	MaestroSkipped MaestroResult = "SKIPPED"
	MaestroPassed  MaestroResult = "PASSED"
	MaestroFailed  MaestroResult = "FAILED"
)

type Project struct {
	ID        string
	UserID    string
	Name      string
	Brief     string
	Stack     ProjectStack
	Status    ProjectStatus
	RootPath  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChatRole string

const (
	RoleUser   ChatRole = "USER"
	RoleAgent  ChatRole = "AGENT"
	RoleSystem ChatRole = "SYSTEM"
)

type JobLog struct {
	ID        string
	ProjectID string
	At        time.Time
	Message   string
	Role      ChatRole
}

type Challenge struct {
	ID           string
	UserID       string
	Purpose      Purpose
	CodeHash     string
	LinkHash     string
	Attempts     int
	MaxAttempts  int
	ExpiresAt    time.Time
	DeviceFPHash string
	DeviceLabel  string
	CodeVerified bool
	Consumed     bool
	TrustDevice  bool
}

type Session struct {
	ID          string
	UserID      string
	TokenHash   string
	DeviceID    string
	DeviceLabel string
	CreatedAt   time.Time
	Revoked     bool
}

type Device struct {
	ID       string
	UserID   string
	FPHash   string
	Label    string
	Trusted  bool
	LastSeen time.Time
}

type Mail struct {
	ID          string
	UserID      string
	ChallengeID string
	Subject     string
	Body        string
	Purpose     Purpose
	CreatedAt   time.Time
}

type Store interface {
	Name() string
	Ping(ctx context.Context) error

	CreateUser(ctx context.Context, user User) (User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	UpdateUser(ctx context.Context, user User) error

	PutChallenge(ctx context.Context, challenge Challenge) error
	GetChallenge(ctx context.Context, id string) (*Challenge, error)
	GetChallengeByLinkHash(ctx context.Context, linkHash string) (*Challenge, error)
	InvalidateChallenges(ctx context.Context, userID string, purpose Purpose) error

	CreateSession(ctx context.Context, session Session) error
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	ListSessions(ctx context.Context, userID string) ([]Session, error)
	RevokeSession(ctx context.Context, id string) error
	RevokeOtherSessions(ctx context.Context, userID, keepID string) error

	UpsertDevice(ctx context.Context, device Device) (Device, error)
	GetDeviceByFP(ctx context.Context, userID, fpHash string) (*Device, error)
	ListDevices(ctx context.Context, userID string) ([]Device, error)
	RevokeDevice(ctx context.Context, id string) error

	AddMail(ctx context.Context, mail Mail) error
	ListMail(ctx context.Context, userID string) ([]Mail, error)
	MailByChallenge(ctx context.Context, challengeID string) (*Mail, error)

	Projects(ctx context.Context, userID string) ([]Project, error)
	CreateProject(ctx context.Context, project Project) (Project, error)
	GetProject(ctx context.Context, id string) (*Project, error)
	UpdateProject(ctx context.Context, project Project) error
	AppendLog(ctx context.Context, log JobLog) error
	ListLogs(ctx context.Context, projectID string) ([]JobLog, error)

	PutLlmVersion(ctx context.Context, version LlmVersion) error
	ListLlmVersions(ctx context.Context, slot LlmSlot) ([]LlmVersion, error)
	GetLlmVersion(ctx context.Context, id string) (*LlmVersion, error)
	GetLlmState(ctx context.Context) (LlmState, error)
	SetLlmState(ctx context.Context, state LlmState) error
	AddAudit(ctx context.Context, event AuditEvent) error
	ListAudit(ctx context.Context, userID string) ([]AuditEvent, error)
	DeleteUserData(ctx context.Context, userID string, wipeProjects bool) error
}

type LlmSlot string

const (
	SlotA LlmSlot = "A"
	SlotB LlmSlot = "B"
)

type LlmVersion struct {
	ID        string
	Slot      LlmSlot
	Name      string
	Note      string
	CreatedAt time.Time
}

type LlmState struct {
	ActiveAID string
	McpRoot   string
}

type AuditEvent struct {
	ID               string
	UserID           string
	ProjectID        string
	Purpose          string
	LegalBasis       string
	Slot             LlmSlot
	VersionID        string
	VersionName      string
	Channel          string
	InputRedactions  int
	OutputRedactions int
	PromptPreview    string
	OutputPreview    string
	CreatedAt        time.Time
}
