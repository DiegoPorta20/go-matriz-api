package token

import "context"

type accessTokenKey struct{}

func WithAccessToken(ctx context.Context, raw string) context.Context {
	return context.WithValue(ctx, accessTokenKey{}, raw)
}

func AccessTokenFrom(ctx context.Context) (string, bool) {
	raw, ok := ctx.Value(accessTokenKey{}).(string)
	return raw, ok && raw != ""
}
