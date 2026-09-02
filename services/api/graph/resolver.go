package graph

//go:generate go run github.com/99designs/gqlgen generate

import (
	"github.com/icerde/api/internal/auth"
	"github.com/icerde/api/internal/connect"
	"github.com/icerde/api/internal/factory"
	"github.com/icerde/api/internal/llm"
)

type Resolver struct {
	Auth    *auth.Service
	Factory *factory.Service
	LLM     *llm.Service
	Connect *connect.Service
}
