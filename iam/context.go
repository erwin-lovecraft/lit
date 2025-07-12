package iam

import (
	"context"
)

type m2mProfileContextKey struct{}

type userProfileContextKey struct{}

func SetM2MProfileInContext(ctx context.Context, profile M2MProfile) context.Context {
	return context.WithValue(ctx, m2mProfileContextKey{}, profile)
}

func GetM2MProfileFromContext(ctx context.Context) M2MProfile {
	if p, ok := ctx.Value(m2mProfileContextKey{}).(M2MProfile); ok {
		return p
	}

	return M2MProfile{}
}

func SetUserProfileInContext(ctx context.Context, profile UserProfile) context.Context {
	return context.WithValue(ctx, userProfileContextKey{}, profile)
}

func GetUserProfileFromContext(ctx context.Context) UserProfile {
	if p, ok := ctx.Value(userProfileContextKey{}).(UserProfile); ok {
		return p
	}

	return UserProfile{}
}
