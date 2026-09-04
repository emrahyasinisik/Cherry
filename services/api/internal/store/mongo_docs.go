package store

import "time"

type userDoc struct {
	ID            string `bson:"_id"`
	Email         string `bson:"email"`
	PasswordHash  string `bson:"passwordHash"`
	WorkspaceKind string `bson:"workspaceKind"`
	TotpSecret    string `bson:"totpSecret"`
	TotpEnabled   bool   `bson:"totpEnabled"`
}

func (d userDoc) toUser() User {
	return User{
		ID:            d.ID,
		Email:         d.Email,
		PasswordHash:  d.PasswordHash,
		WorkspaceKind: WorkspaceKind(d.WorkspaceKind),
		TotpSecret:    d.TotpSecret,
		TotpEnabled:   d.TotpEnabled,
	}
}

type challengeDoc struct {
	ID           string    `bson:"_id"`
	UserID       string    `bson:"userId"`
	Purpose      string    `bson:"purpose"`
	CodeHash     string    `bson:"codeHash"`
	LinkHash     string    `bson:"linkHash"`
	Attempts     int       `bson:"attempts"`
	MaxAttempts  int       `bson:"maxAttempts"`
	ExpiresAt    time.Time `bson:"expiresAt"`
	DeviceFPHash string    `bson:"deviceFpHash"`
	DeviceLabel  string    `bson:"deviceLabel"`
	CodeVerified bool      `bson:"codeVerified"`
	Consumed     bool      `bson:"consumed"`
	TrustDevice  bool      `bson:"trustDevice"`
}

func (d challengeDoc) toChallenge() Challenge {
	return Challenge{
		ID:           d.ID,
		UserID:       d.UserID,
		Purpose:      Purpose(d.Purpose),
		CodeHash:     d.CodeHash,
		LinkHash:     d.LinkHash,
		Attempts:     d.Attempts,
		MaxAttempts:  d.MaxAttempts,
		ExpiresAt:    d.ExpiresAt,
		DeviceFPHash: d.DeviceFPHash,
		DeviceLabel:  d.DeviceLabel,
		CodeVerified: d.CodeVerified,
		Consumed:     d.Consumed,
		TrustDevice:  d.TrustDevice,
	}
}

type sessionDoc struct {
	ID          string    `bson:"_id"`
	UserID      string    `bson:"userId"`
	TokenHash   string    `bson:"tokenHash"`
	DeviceID    string    `bson:"deviceId"`
	DeviceLabel string    `bson:"deviceLabel"`
	CreatedAt   time.Time `bson:"createdAt"`
	Revoked     bool      `bson:"revoked"`
}

func (d sessionDoc) toSession() Session {
	return Session{
		ID:          d.ID,
		UserID:      d.UserID,
		TokenHash:   d.TokenHash,
		DeviceID:    d.DeviceID,
		DeviceLabel: d.DeviceLabel,
		CreatedAt:   d.CreatedAt,
		Revoked:     d.Revoked,
	}
}

type deviceDoc struct {
	ID       string    `bson:"_id"`
	UserID   string    `bson:"userId"`
	FPHash   string    `bson:"fpHash"`
	Label    string    `bson:"label"`
	Trusted  bool      `bson:"trusted"`
	LastSeen time.Time `bson:"lastSeen"`
}

func (d deviceDoc) toDevice() Device {
	return Device{
		ID:       d.ID,
		UserID:   d.UserID,
		FPHash:   d.FPHash,
		Label:    d.Label,
		Trusted:  d.Trusted,
		LastSeen: d.LastSeen,
	}
}

type mailDoc struct {
	ID          string    `bson:"_id"`
	UserID      string    `bson:"userId"`
	ChallengeID string    `bson:"challengeId"`
	Subject     string    `bson:"subject"`
	Body        string    `bson:"body"`
	Purpose     string    `bson:"purpose"`
	CreatedAt   time.Time `bson:"createdAt"`
}

func (d mailDoc) toMail() Mail {
	return Mail{
		ID:          d.ID,
		UserID:      d.UserID,
		ChallengeID: d.ChallengeID,
		Subject:     d.Subject,
		Body:        d.Body,
		Purpose:     Purpose(d.Purpose),
		CreatedAt:   d.CreatedAt,
	}
}

type projectDoc struct {
	ID        string    `bson:"_id"`
	UserID    string    `bson:"userId"`
	Name      string    `bson:"name"`
	Brief     string    `bson:"brief"`
	Stack     string    `bson:"stack"`
	Status    string    `bson:"status"`
	RootPath  string    `bson:"rootPath"`
	Backend   string    `bson:"backend"`
	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

func (d projectDoc) toProject() Project {
	return Project{
		ID:        d.ID,
		UserID:    d.UserID,
		Name:      d.Name,
		Brief:     d.Brief,
		Stack:     ProjectStack(d.Stack),
		Status:    ProjectStatus(d.Status),
		RootPath:  d.RootPath,
		Backend:   BackendTarget(d.Backend),
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

type jobLogDoc struct {
	ID        string    `bson:"_id"`
	ProjectID string    `bson:"projectId"`
	At        time.Time `bson:"at"`
	Message   string    `bson:"message"`
	Role      string    `bson:"role"`
}

func (d jobLogDoc) toJobLog() JobLog {
	return JobLog{
		ID:        d.ID,
		ProjectID: d.ProjectID,
		At:        d.At,
		Message:   d.Message,
		Role:      ChatRole(d.Role),
	}
}

type connectionDoc struct {
	ID         string    `bson:"_id"`
	UserID     string    `bson:"userId"`
	Kind       string    `bson:"kind"`
	Status     string    `bson:"status"`
	Account    string    `bson:"account"`
	Token      string    `bson:"token"`
	TokenHint  string    `bson:"tokenHint"`
	Note       string    `bson:"note"`
	AuthMethod string    `bson:"authMethod"`
	Scopes     []string  `bson:"scopes"`
	UpdatedAt  time.Time `bson:"updatedAt"`
}

func (d connectionDoc) toConnection() Connection {
	scopes := d.Scopes
	if scopes == nil {
		scopes = []string{}
	}
	return Connection{
		ID:         d.ID,
		UserID:     d.UserID,
		Kind:       ConnectionKind(d.Kind),
		Status:     ConnectionStatus(d.Status),
		Account:    d.Account,
		Token:      d.Token,
		TokenHint:  d.TokenHint,
		Note:       d.Note,
		AuthMethod: ConnectionAuth(d.AuthMethod),
		Scopes:     scopes,
		UpdatedAt:  d.UpdatedAt,
	}
}

type llmVersionDoc struct {
	ID            string    `bson:"_id"`
	Slot          string    `bson:"slot"`
	Name          string    `bson:"name"`
	Note          string    `bson:"note"`
	CheckpointRef string    `bson:"checkpointRef"`
	CreatedAt     time.Time `bson:"createdAt"`
}

func (d llmVersionDoc) toLlmVersion() LlmVersion {
	return LlmVersion{
		ID:            d.ID,
		Slot:          LlmSlot(d.Slot),
		Name:          d.Name,
		Note:          d.Note,
		CheckpointRef: d.CheckpointRef,
		CreatedAt:     d.CreatedAt,
	}
}

type llmStateDoc struct {
	ID        string `bson:"_id"`
	ActiveAID string `bson:"activeAId"`
	ActiveBID string `bson:"activeBId"`
	McpRoot   string `bson:"mcpRoot"`
}

type auditDoc struct {
	ID               string    `bson:"_id"`
	UserID           string    `bson:"userId"`
	ProjectID        string    `bson:"projectId"`
	Purpose          string    `bson:"purpose"`
	LegalBasis       string    `bson:"legalBasis"`
	Slot             string    `bson:"slot"`
	VersionID        string    `bson:"versionId"`
	VersionName      string    `bson:"versionName"`
	Channel          string    `bson:"channel"`
	InputRedactions  int       `bson:"inputRedactions"`
	OutputRedactions int       `bson:"outputRedactions"`
	PromptPreview    string    `bson:"promptPreview"`
	OutputPreview    string    `bson:"outputPreview"`
	CreatedAt        time.Time `bson:"createdAt"`
}

func (d auditDoc) toAudit() AuditEvent {
	return AuditEvent{
		ID:               d.ID,
		UserID:           d.UserID,
		ProjectID:        d.ProjectID,
		Purpose:          d.Purpose,
		LegalBasis:       d.LegalBasis,
		Slot:             LlmSlot(d.Slot),
		VersionID:        d.VersionID,
		VersionName:      d.VersionName,
		Channel:          d.Channel,
		InputRedactions:  d.InputRedactions,
		OutputRedactions: d.OutputRedactions,
		PromptPreview:    d.PromptPreview,
		OutputPreview:    d.OutputPreview,
		CreatedAt:        d.CreatedAt,
	}
}
