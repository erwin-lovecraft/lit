package iam

import (
	"context"
	"reflect"
	"testing"
)

func TestGetM2MProfileFromContext_NoValue(t *testing.T) {
	got := GetM2MProfileFromContext(context.Background())
	if !reflect.DeepEqual(got, M2MProfile{}) {
		t.Errorf("GetM2MProfileFromContext without value = %v; want zero value", got)
	}
}

func TestGetM2MProfileFromContext_InvalidType(t *testing.T) {
	ctx := context.WithValue(context.Background(), m2mProfileContextKey{}, "invalid")
	got := GetM2MProfileFromContext(ctx)
	if !reflect.DeepEqual(got, M2MProfile{}) {
		t.Errorf("GetM2MProfileFromContext with invalid type = %v; want zero value", got)
	}
}

func TestSetAndGetM2MProfileInContext(t *testing.T) {
	profile := M2MProfile{}
	ctx := SetM2MProfileInContext(context.Background(), profile)
	got := GetM2MProfileFromContext(ctx)
	if !reflect.DeepEqual(got, profile) {
		t.Errorf("After Set/Get, got = %v; want %v", got, profile)
	}
}

func TestGetUserProfileFromContext_NoValue(t *testing.T) {
	got := GetUserProfileFromContext(context.Background())
	if !reflect.DeepEqual(got, UserProfile{}) {
		t.Errorf("GetUserProfileFromContext without value = %v; want zero value", got)
	}
}

func TestGetUserProfileFromContext_InvalidType(t *testing.T) {
	ctx := context.WithValue(context.Background(), userProfileContextKey{}, 123)
	got := GetUserProfileFromContext(ctx)
	if !reflect.DeepEqual(got, UserProfile{}) {
		t.Errorf("GetUserProfileFromContext with invalid type = %v; want zero value", got)
	}
}

func TestSetAndGetUserProfileInContext(t *testing.T) {
	profile := UserProfile{}
	ctx := SetUserProfileInContext(context.Background(), profile)
	got := GetUserProfileFromContext(ctx)
	if !reflect.DeepEqual(got, profile) {
		t.Errorf("After Set/Get, got = %v; want %v", got, profile)
	}
}
