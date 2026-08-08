package service

import (
	"context"
)

type identityContextKey struct{}

func WithIdentity(requestContext context.Context, identity Identity) context.Context {
	return context.WithValue(requestContext, identityContextKey{}, identity)
}

func IdentityFromContext(requestContext context.Context) (Identity, bool) {
	identity, ok := requestContext.Value(identityContextKey{}).(Identity)
	return identity, ok
}
