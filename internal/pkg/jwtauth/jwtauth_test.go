package jwtauth_test

import (
	"testing"
	"time"

	"github.com/Leopold1975/yadro_app/internal/auth/models"
	"github.com/Leopold1975/yadro_app/internal/pkg/jwtauth"
	"github.com/stretchr/testify/require"
)

var (
	userExample = models.User{
		ID:           1,
		Username:     "user",
		PasswordHash: "1234",
		Role:         "admin",
	}
	secret     = "secret"
	defaultTTL = time.Minute * 5
)

func TestBasic(t *testing.T) {
	token, err := jwtauth.GetToken(userExample, defaultTTL, secret)
	require.NoError(t, err)

	role, err := jwtauth.ValidateTokenRole(token, secret)
	require.NoError(t, err)
	require.Equal(t, userExample.Role, models.Role(role))
}

func TestValidateToken(t *testing.T) {
	expiredTokenString, err := jwtauth.GetToken(userExample, -1*time.Minute, secret)
	require.NoError(t, err)

	_, err = jwtauth.ValidateTokenRole(expiredTokenString, secret)
	require.ErrorIs(t, err, jwtauth.ErrTokenExpired)

	tokenString, _ := jwtauth.GetToken(userExample, defaultTTL, secret)

	_, err = jwtauth.ValidateTokenRole(tokenString, "wrongsecret")
	require.NotNil(t, err)

	_, err = jwtauth.ValidateTokenRole(tokenString+"1", secret)
	require.NotNil(t, err)

	wrongUser := models.User{
		ID:           1,
		Username:     "user",
		PasswordHash: "1234",
	}
	_, err = jwtauth.GetToken(wrongUser, defaultTTL, secret)
	require.ErrorIs(t, err, jwtauth.ErrNoClaim)

	_, err = jwtauth.ValidateTokenRole("asasas", "wrongsecret")
	require.NotNil(t, err)
}
