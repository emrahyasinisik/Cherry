package graph

//go:generate go run github.com/99designs/gqlgen generate

import (
	"github.com/icerde/api/internal/auth"
	"github.com/icerde/api/internal/factory"
)

type Resolver struct {
	Auth    *auth.Service
	Factory *factory.Service
}
