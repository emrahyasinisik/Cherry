package graph

import (
	"context"
	"errors"

	"github.com/icerde/api/internal/store"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const apiVersion = "0.1.0-scaffold"

func (r *mutationResolver) Login(ctx context.Context, email string, password string) (*AuthPayload, error) {
	token, user, err := r.Store.Login(ctx, email, password)
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			return nil, &gqlerror.Error{Message: "E-posta ve şifre gerekli."}
		}
		if errors.Is(err, store.ErrInvalidCredentials) {
			return nil, &gqlerror.Error{Message: "Giriş bilgileri geçersiz."}
		}
		return nil, &gqlerror.Error{Message: "Giriş yapılamadı."}
	}
	mapped, err := mapUser(user)
	if err != nil {
		return nil, &gqlerror.Error{Message: "Giriş yapılamadı."}
	}
	return &AuthPayload{Token: token, User: mapped}, nil
}

func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
	if err := r.Store.Logout(ctx, TokenFrom(ctx)); err != nil {
		return false, &gqlerror.Error{Message: "Çıkış yapılamadı."}
	}
	return true, nil
}

func (r *queryResolver) Health(ctx context.Context) (*Health, error) {
	storeName := r.Store.Name()
	if err := r.Store.Ping(ctx); err != nil {
		return &Health{Ok: false, Store: storeName, Version: apiVersion}, nil
	}
	return &Health{Ok: true, Store: storeName, Version: apiVersion}, nil
}

func (r *queryResolver) Me(ctx context.Context) (*User, error) {
	user, err := r.Store.Me(ctx, TokenFrom(ctx))
	if err != nil {
		if errors.Is(err, store.ErrUnauthorized) {
			return nil, nil
		}
		return nil, &gqlerror.Error{Message: "Oturum okunamadı."}
	}
	return mapUser(*user)
}

func (r *queryResolver) Projects(ctx context.Context) ([]*Project, error) {
	rows, err := r.Store.Projects(ctx, TokenFrom(ctx))
	if err != nil {
		if errors.Is(err, store.ErrUnauthorized) {
			return nil, &gqlerror.Error{Message: "Oturum gerekli."}
		}
		return nil, &gqlerror.Error{Message: "Projeler yüklenemedi."}
	}
	return mapProjects(rows), nil
}

func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
