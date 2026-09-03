package graph

//go:generate go run github.com/99designs/gqlgen generate

import (
	"github.com/cherry/api/internal/auth"
	"github.com/cherry/api/internal/colabbridge"
	"github.com/cherry/api/internal/connect"
	"github.com/cherry/api/internal/factory"
	"github.com/cherry/api/internal/llm"
)

type Resolver struct {
	Auth    *auth.Service
	Factory *factory.Service
	LLM     *llm.Service
	Connect *connect.Service
	Bridge  *colabbridge.Service
}
