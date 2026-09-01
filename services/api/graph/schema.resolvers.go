package graph

import (
	"context"
	"errors"
	"time"

	"github.com/icerde/api/internal/store"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const apiVersion = "0.2.0-auth"

func gqlErr(err error) error {
	switch {
	case errors.Is(err, store.ErrValidation):
		return &gqlerror.Error{Message: "E-posta geçerli olmalı, şifre en az 8 karakter."}
	case errors.Is(err, store.ErrInvalidCredentials):
		return &gqlerror.Error{Message: "Bilgiler geçersiz."}
	case errors.Is(err, store.ErrExists):
		return &gqlerror.Error{Message: "Bu e-posta kayıtlı. Giriş yap."}
	case errors.Is(err, store.ErrLocked):
		return &gqlerror.Error{Message: "Çok fazla deneme. Biraz bekle."}
	case errors.Is(err, store.ErrExpired):
		return &gqlerror.Error{Message: "Kod veya link süresi doldu."}
	case errors.Is(err, store.ErrUnauthorized):
		return &gqlerror.Error{Message: "Oturum gerekli."}
	case errors.Is(err, store.ErrNotFound):
		return &gqlerror.Error{Message: "Kayıt bulunamadı."}
	default:
		return &gqlerror.Error{Message: "İşlem yapılamadı."}
	}
}

func (r *mutationResolver) Register(ctx context.Context, email string, password string, deviceFingerprint string, deviceLabel string) (*LoginResult, error) {
	result, err := r.Auth.Register(ctx, email, password, deviceFingerprint, deviceLabel, IPFrom(ctx))
	if err != nil {
		return nil, gqlErr(err)
	}
	return mapLoginResult(result)
}

func (r *mutationResolver) Login(ctx context.Context, email string, password string, deviceFingerprint string, deviceLabel string) (*LoginResult, error) {
	result, err := r.Auth.Login(ctx, email, password, deviceFingerprint, deviceLabel, IPFrom(ctx))
	if err != nil {
		return nil, gqlErr(err)
	}
	return mapLoginResult(result)
}

func (r *mutationResolver) VerifyCode(ctx context.Context, challengeID string, code string, trustDevice bool) (*LoginResult, error) {
	result, err := r.Auth.VerifyCode(ctx, challengeID, code, trustDevice, IPFrom(ctx))
	if err != nil {
		return nil, gqlErr(err)
	}
	return mapLoginResult(result)
}

func (r *mutationResolver) VerifyLink(ctx context.Context, token string, deviceFingerprint string, deviceLabel string) (*LoginResult, error) {
	result, err := r.Auth.VerifyLink(ctx, token, deviceFingerprint, deviceLabel, IPFrom(ctx))
	if err != nil {
		return nil, gqlErr(err)
	}
	return mapLoginResult(result)
}

func (r *mutationResolver) VerifyTotp(ctx context.Context, challengeID string, code string) (*LoginResult, error) {
	result, err := r.Auth.VerifyTotp(ctx, challengeID, code, IPFrom(ctx))
	if err != nil {
		return nil, gqlErr(err)
	}
	return mapLoginResult(result)
}

func (r *mutationResolver) EnableTotp(ctx context.Context) (*TotpSetup, error) {
	user, _, err := r.Auth.SessionUser(ctx, TokenFrom(ctx))
	if err != nil {
		return nil, gqlErr(err)
	}
	secret, url, err := r.Auth.EnableTotp(ctx, user.ID)
	if err != nil {
		return nil, gqlErr(err)
	}
	return &TotpSetup{Secret: secret, OtpauthURL: url}, nil
}

func (r *mutationResolver) ConfirmTotp(ctx context.Context, code string) (bool, error) {
	user, _, err := r.Auth.SessionUser(ctx, TokenFrom(ctx))
	if err != nil {
		return false, gqlErr(err)
	}
	if err := r.Auth.ConfirmTotp(ctx, user.ID, code); err != nil {
		return false, gqlErr(err)
	}
	return true, nil
}

func (r *mutationResolver) DisableTotp(ctx context.Context, code string) (bool, error) {
	user, _, err := r.Auth.SessionUser(ctx, TokenFrom(ctx))
	if err != nil {
		return false, gqlErr(err)
	}
	if err := r.Auth.DisableTotp(ctx, user.ID, code); err != nil {
		return false, gqlErr(err)
	}
	return true, nil
}

func (r *mutationResolver) RevokeSession(ctx context.Context, id string) (bool, error) {
	user, _, err := r.Auth.SessionUser(ctx, TokenFrom(ctx))
	if err != nil {
		return false, gqlErr(err)
	}
	sessions, err := r.Auth.Store.ListSessions(ctx, user.ID)
	if err != nil {
		return false, gqlErr(err)
	}
	found := false
	for _, session := range sessions {
		if session.ID == id {
			found = true
			break
		}
	}
	if !found {
		return false, gqlErr(store.ErrNotFound)
	}
	if err := r.Auth.Store.RevokeSession(ctx, id); err != nil {
		return false, gqlErr(err)
	}
	return true, nil
}

func (r *mutationResolver) RevokeDevice(ctx context.Context, id string) (bool, error) {
	user, _, err := r.Auth.SessionUser(ctx, TokenFrom(ctx))
	if err != nil {
		return false, gqlErr(err)
	}
	devices, err := r.Auth.Store.ListDevices(ctx, user.ID)
	if err != nil {
		return false, gqlErr(err)
	}
	found := false
	for _, device := range devices {
		if device.ID == id {
			found = true
			break
		}
	}
	if !found {
		return false, gqlErr(store.ErrNotFound)
	}
	if err := r.Auth.Store.RevokeDevice(ctx, id); err != nil {
		return false, gqlErr(err)
	}
	return true, nil
}

func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
	if err := r.Auth.Logout(ctx, TokenFrom(ctx)); err != nil {
		return false, gqlErr(err)
	}
	return true, nil
}

func (r *queryResolver) Health(ctx context.Context) (*Health, error) {
	ok := r.Auth.Store.Ping(ctx) == nil
	return &Health{Ok: ok, Store: r.Auth.Store.Name(), Version: apiVersion}, nil
}

func (r *queryResolver) Me(ctx context.Context) (*User, error) {
	user, _, err := r.Auth.SessionUser(ctx, TokenFrom(ctx))
	if err != nil {
		if errors.Is(err, store.ErrUnauthorized) {
			return nil, nil
		}
		return nil, gqlErr(err)
	}
	return mapUser(*user)
}

func (r *queryResolver) Projects(ctx context.Context) ([]*Project, error) {
	user, _, err := r.Auth.SessionUser(ctx, TokenFrom(ctx))
	if err != nil {
		return nil, gqlErr(err)
	}
	rows, err := r.Auth.Store.Projects(ctx, user.ID)
	if err != nil {
		return nil, gqlErr(err)
	}
	return mapProjects(rows), nil
}

func (r *queryResolver) Devices(ctx context.Context) ([]*Device, error) {
	user, sess, err := r.Auth.SessionUser(ctx, TokenFrom(ctx))
	if err != nil {
		return nil, gqlErr(err)
	}
	rows, err := r.Auth.Store.ListDevices(ctx, user.ID)
	if err != nil {
		return nil, gqlErr(err)
	}
	out := make([]*Device, 0, len(rows))
	for _, row := range rows {
		item := row
		out = append(out, &Device{
			ID:         item.ID,
			Label:      item.Label,
			Trusted:    item.Trusted,
			Current:    item.ID == sess.DeviceID,
			LastSeenAt: item.LastSeen.UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}

func (r *queryResolver) Sessions(ctx context.Context) ([]*Session, error) {
	user, current, err := r.Auth.SessionUser(ctx, TokenFrom(ctx))
	if err != nil {
		return nil, gqlErr(err)
	}
	rows, err := r.Auth.Store.ListSessions(ctx, user.ID)
	if err != nil {
		return nil, gqlErr(err)
	}
	out := make([]*Session, 0, len(rows))
	for _, row := range rows {
		item := row
		out = append(out, &Session{
			ID:          item.ID,
			Current:     item.ID == current.ID,
			CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
			DeviceLabel: item.DeviceLabel,
		})
	}
	return out, nil
}

func (r *queryResolver) Mailbox(ctx context.Context) ([]*MailMessage, error) {
	user, _, err := r.Auth.SessionUser(ctx, TokenFrom(ctx))
	if err != nil {
		return nil, gqlErr(err)
	}
	rows, err := r.Auth.Store.ListMail(ctx, user.ID)
	if err != nil {
		return nil, gqlErr(err)
	}
	out := make([]*MailMessage, 0, len(rows))
	for _, row := range rows {
		msg, err := mapMail(row)
		if err != nil {
			return nil, gqlErr(err)
		}
		out = append(out, msg)
	}
	return out, nil
}

func (r *queryResolver) ChallengeMailbox(ctx context.Context, challengeID string) (*MailMessage, error) {
	mail, err := r.Auth.Store.MailByChallenge(ctx, challengeID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, nil
		}
		return nil, gqlErr(err)
	}
	return mapMail(*mail)
}

func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
