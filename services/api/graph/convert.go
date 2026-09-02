package graph

import (
	"fmt"
	"time"

	"github.com/icerde/api/internal/activate"
	"github.com/icerde/api/internal/auth"
	"github.com/icerde/api/internal/factory"
	"github.com/icerde/api/internal/mailer"
	"github.com/icerde/api/internal/store"
)

func mapUser(user store.User) (*User, error) {
	kind, err := mapWorkspaceKind(user.WorkspaceKind)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:            user.ID,
		Email:         user.Email,
		WorkspaceKind: kind,
		TotpEnabled:   user.TotpEnabled,
	}, nil
}

func mapWorkspaceKind(kind store.WorkspaceKind) (WorkspaceKind, error) {
	switch kind {
	case store.WorkspacePersonal:
		return WorkspaceKindPersonal, nil
	case store.WorkspaceOrganization:
		return WorkspaceKindOrganization, nil
	default:
		return "", fmt.Errorf("unhandled workspace kind: %s", kind)
	}
}

func mapLoginNext(next string) (LoginNext, error) {
	switch next {
	case auth.NextSession:
		return LoginNextSession, nil
	case auth.NextDeviceCode:
		return LoginNextDeviceCode, nil
	case auth.NextTOTP:
		return LoginNextTotp, nil
	default:
		return "", fmt.Errorf("unhandled login next: %s", next)
	}
}

func mapPurpose(purpose store.Purpose) (VerifyPurpose, error) {
	switch purpose {
	case store.PurposeNewDevice:
		return VerifyPurposeNewDevice, nil
	case store.PurposeLoginChallenge:
		return VerifyPurposeLoginChallenge, nil
	case store.PurposeEmailVerify:
		return VerifyPurposeEmailVerify, nil
	case store.PurposeSuspiciousLogin:
		return VerifyPurposeSuspiciousLogin, nil
	default:
		return "", fmt.Errorf("unhandled purpose: %s", purpose)
	}
}

func mapLoginResult(result auth.Result) (*LoginResult, error) {
	next, err := mapLoginNext(result.Next)
	if err != nil {
		return nil, err
	}
	channel, err := mapEmailChannel(result.EmailChannel)
	if err != nil {
		return nil, err
	}
	out := &LoginResult{
		Next:         next,
		EmailSent:    result.EmailSent,
		EmailChannel: channel,
	}
	if result.Token != "" {
		token := result.Token
		out.Token = &token
	}
	if result.ChallengeID != "" {
		id := result.ChallengeID
		out.ChallengeID = &id
	}
	if result.User.ID != "" {
		user, err := mapUser(result.User)
		if err != nil {
			return nil, err
		}
		out.User = user
	}
	return out, nil
}

func mapEmailChannel(channel string) (string, error) {
	switch channel {
	case "", "none":
		return "", nil
	case mailer.ChannelInbox, mailer.ChannelSMTP, mailer.ChannelResend:
		return channel, nil
	default:
		return "", fmt.Errorf("unhandled mail channel: %s", channel)
	}
}

func mapProjects(rows []store.Project) ([]*Project, error) {
	out := make([]*Project, 0, len(rows))
	for _, row := range rows {
		item, err := mapProjectMeta(row)
		if err != nil {
			return nil, err
		}
		item.Logs = []*JobLog{}
		item.Files = []*ProjectFile{}
		item.Maestro = emptyMaestro()
		item.Activate = mapActivate(activate.Snapshot{Status: activate.StatusIdle, Note: "Yerel API kapalı."})
		out = append(out, item)
	}
	return out, nil
}

func mapProjectMeta(row store.Project) (*Project, error) {
	stack, err := mapProjectStack(row.Stack)
	if err != nil {
		return nil, err
	}
	status, err := mapProjectStatus(row.Status)
	if err != nil {
		return nil, err
	}
	backend, err := mapBackendTarget(row.Backend)
	if err != nil {
		return nil, err
	}
	return &Project{
		ID:            row.ID,
		Name:          row.Name,
		Brief:         row.Brief,
		Stack:         stack,
		Status:        status,
		RootPath:      row.RootPath,
		BackendTarget: backend,
		CreatedAt:     row.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func mapBackendTarget(target store.BackendTarget) (BackendTarget, error) {
	if target == "" {
		target = store.TargetLocal
	}
	switch target {
	case store.TargetLocal:
		return BackendTargetLocal, nil
	case store.TargetSupabase:
		return BackendTargetSupabase, nil
	case store.TargetCloudflare:
		return BackendTargetCloudflare, nil
	case store.TargetRender:
		return BackendTargetRender, nil
	default:
		return "", fmt.Errorf("unhandled backend target: %s", target)
	}
}

func mapConnectionKind(kind store.ConnectionKind) (ConnectionKind, error) {
	switch kind {
	case store.KindSupabase:
		return ConnectionKindSupabase, nil
	case store.KindCloudflare:
		return ConnectionKindCloudflare, nil
	case store.KindGithub:
		return ConnectionKindGithub, nil
	case store.KindVercel:
		return ConnectionKindVercel, nil
	case store.KindRender:
		return ConnectionKindRender, nil
	default:
		return "", fmt.Errorf("unhandled connection kind: %s", kind)
	}
}

func mapConnectionStatus(status store.ConnectionStatus) (ConnectionStatus, error) {
	switch status {
	case store.ConnDisconnected:
		return ConnectionStatusDisconnected, nil
	case store.ConnConnected:
		return ConnectionStatusConnected, nil
	case store.ConnFailed:
		return ConnectionStatusFailed, nil
	default:
		return "", fmt.Errorf("unhandled connection status: %s", status)
	}
}

func mapConnection(row store.Connection) (*Connection, error) {
	kind, err := mapConnectionKind(row.Kind)
	if err != nil {
		return nil, err
	}
	status, err := mapConnectionStatus(row.Status)
	if err != nil {
		return nil, err
	}
	auth, err := mapConnectionAuth(row.AuthMethod, row.Status)
	if err != nil {
		return nil, err
	}
	scopes := row.Scopes
	if scopes == nil {
		scopes = []string{}
	}
	return &Connection{
		Kind:       kind,
		Status:     status,
		Account:    row.Account,
		TokenHint:  row.TokenHint,
		Note:       row.Note,
		AuthMethod: auth,
		Scopes:     scopes,
	}, nil
}

func mapConnectionAuth(method store.ConnectionAuth, status store.ConnectionStatus) (ConnectionAuth, error) {
	switch method {
	case store.AuthOAuth:
		return ConnectionAuthOauth, nil
	case store.AuthToken:
		return ConnectionAuthToken, nil
	case store.AuthNone, "":
		if status == store.ConnConnected {
			return ConnectionAuthToken, nil
		}
		return ConnectionAuthNone, nil
	default:
		return "", fmt.Errorf("unhandled connection auth: %s", method)
	}
}

func mapOAuthMode(mode string) (OAuthMode, error) {
	switch mode {
	case "CONSENT":
		return OAuthModeConsent, nil
	case "PROVIDER":
		return OAuthModeProvider, nil
	default:
		return "", fmt.Errorf("unhandled oauth mode: %s", mode)
	}
}

func mapProjectStack(stack store.ProjectStack) (ProjectStack, error) {
	switch stack {
	case store.StackExpo:
		return ProjectStackExpo, nil
	case store.StackFlutter:
		return ProjectStackFlutter, nil
	case store.StackNative:
		return ProjectStackNative, nil
	default:
		return "", fmt.Errorf("unhandled project stack: %s", stack)
	}
}

func mapProjectStatus(status store.ProjectStatus) (ProjectStatus, error) {
	switch status {
	case store.StatusQueued:
		return ProjectStatusQueued, nil
	case store.StatusWriting:
		return ProjectStatusWriting, nil
	case store.StatusTesting:
		return ProjectStatusTesting, nil
	case store.StatusReady:
		return ProjectStatusReady, nil
	case store.StatusFailed:
		return ProjectStatusFailed, nil
	default:
		return "", fmt.Errorf("unhandled project status: %s", status)
	}
}

func mapMaestroResult(result store.MaestroResult) (MaestroResult, error) {
	switch result {
	case store.MaestroSkipped:
		return MaestroResultSkipped, nil
	case store.MaestroPassed:
		return MaestroResultPassed, nil
	case store.MaestroFailed:
		return MaestroResultFailed, nil
	default:
		return "", fmt.Errorf("unhandled maestro result: %s", result)
	}
}

func mapMaestro(studio factory.MaestroStudio) (*MaestroStudio, error) {
	out := &MaestroStudio{
		Ready:        studio.Ready,
		DeviceStatus: studio.DeviceStatus,
		Screens:      make([]*DesignScreen, 0, len(studio.Screens)),
		Flows:        make([]*MaestroFlow, 0, len(studio.Flows)),
	}
	for _, screen := range studio.Screens {
		item := screen
		out.Screens = append(out.Screens, &DesignScreen{ID: item.ID, Name: item.Name, HTML: item.HTML})
	}
	for _, flow := range studio.Flows {
		result, err := mapMaestroResult(flow.Result)
		if err != nil {
			return nil, err
		}
		item := flow
		out.Flows = append(out.Flows, &MaestroFlow{
			ID:     item.ID,
			Name:   item.Name,
			Yaml:   item.YAML,
			Result: result,
			Note:   item.Note,
		})
	}
	return out, nil
}

func mapActivate(snap activate.Snapshot) *LocalActivate {
	status, err := mapActivateStatus(snap.Status)
	if err != nil {
		status = ActivateStatusIdle
	}
	out := &LocalActivate{Status: status, Note: snap.Note}
	if snap.URL != "" {
		url := snap.URL
		out.URL = &url
	}
	if snap.Port > 0 {
		port := snap.Port
		out.Port = &port
	}
	if snap.PID > 0 {
		pid := snap.PID
		out.Pid = &pid
	}
	return out
}

func mapActivateStatus(status activate.Status) (ActivateStatus, error) {
	switch status {
	case activate.StatusIdle:
		return ActivateStatusIdle, nil
	case activate.StatusStarting:
		return ActivateStatusStarting, nil
	case activate.StatusRunning:
		return ActivateStatusRunning, nil
	case activate.StatusStopping:
		return ActivateStatusStopping, nil
	case activate.StatusFailed:
		return ActivateStatusFailed, nil
	default:
		return "", fmt.Errorf("unhandled activate status: %s", status)
	}
}

func emptyMaestro() *MaestroStudio {
	return &MaestroStudio{
		Ready:        false,
		DeviceStatus: "none",
		Screens:      []*DesignScreen{},
		Flows:        []*MaestroFlow{},
	}
}

func mapLogs(rows []store.JobLog) ([]*JobLog, error) {
	out := make([]*JobLog, 0, len(rows))
	for _, row := range rows {
		item := row
		role, err := mapChatRole(item.Role)
		if err != nil {
			return nil, err
		}
		out = append(out, &JobLog{At: item.At.UTC().Format(time.RFC3339), Message: item.Message, Role: role})
	}
	return out, nil
}

func mapChatRole(role store.ChatRole) (ChatRole, error) {
	switch role {
	case "", store.RoleSystem:
		return ChatRoleSystem, nil
	case store.RoleUser:
		return ChatRoleUser, nil
	case store.RoleAgent:
		return ChatRoleAgent, nil
	default:
		return "", fmt.Errorf("unhandled chat role: %s", role)
	}
}

func mapFiles(rows []factory.DiskFile) []*ProjectFile {
	out := make([]*ProjectFile, 0, len(rows))
	for _, row := range rows {
		item := row
		out = append(out, &ProjectFile{Path: item.Path, Kind: item.Kind})
	}
	return out
}

func mapMail(mail store.Mail) (*MailMessage, error) {
	purpose, err := mapPurpose(mail.Purpose)
	if err != nil {
		return nil, err
	}
	return &MailMessage{
		ID:        mail.ID,
		Subject:   mail.Subject,
		Body:      mail.Body,
		Purpose:   purpose,
		CreatedAt: mail.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}
