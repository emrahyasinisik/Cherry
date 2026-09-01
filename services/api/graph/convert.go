package graph

import (
	"fmt"
	"time"

	"github.com/icerde/api/internal/auth"
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
	out := &LoginResult{Next: next}
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

func mapProjects(rows []store.Project) []*Project {
	out := make([]*Project, 0, len(rows))
	for _, row := range rows {
		item := row
		out = append(out, &Project{
			ID:     item.ID,
			Name:   item.Name,
			Stack:  item.Stack,
			Status: item.Status,
		})
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
