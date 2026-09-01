package graph

import "context"

type ctxKey string

const (
	tokenCtxKey ctxKey = "authToken"
	ipCtxKey    ctxKey = "clientIP"
)

func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenCtxKey, token)
}

func TokenFrom(ctx context.Context) string {
	value, _ := ctx.Value(tokenCtxKey).(string)
	return value
}

func WithIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ipCtxKey, ip)
}

func IPFrom(ctx context.Context) string {
	value, _ := ctx.Value(ipCtxKey).(string)
	if value == "" {
		return "unknown"
	}
	return value
}
