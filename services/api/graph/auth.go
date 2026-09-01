package graph

import "context"

type ctxKey string

const tokenCtxKey ctxKey = "authToken"

func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenCtxKey, token)
}

func TokenFrom(ctx context.Context) string {
	value, _ := ctx.Value(tokenCtxKey).(string)
	return value
}
