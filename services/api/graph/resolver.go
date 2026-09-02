package graph

//go:generate go run github.com/99designs/gqlgen generate

import "github.com/icerde/api/internal/auth"

type Resolver struct {
	Auth *auth.Service
}
