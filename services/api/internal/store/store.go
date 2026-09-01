package store

import (
	"context"
	"errors"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrValidation         = errors.New("validation")
)

type WorkspaceKind string

const (
	WorkspacePersonal     WorkspaceKind = "PERSONAL"
	WorkspaceOrganization WorkspaceKind = "ORGANIZATION"
)

type User struct {
	ID            string
	Email         string
	WorkspaceKind WorkspaceKind
}

type Project struct {
	ID     string
	Name   string
	Stack  string
	Status string
}

type Store interface {
	Name() string
	Ping(ctx context.Context) error
	Login(ctx context.Context, email, password string) (token string, user User, err error)
	Logout(ctx context.Context, token string) error
	Me(ctx context.Context, token string) (*User, error)
	Projects(ctx context.Context, token string) ([]Project, error)
}
