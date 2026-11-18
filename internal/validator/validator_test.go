package validator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReservedUsernames(t *testing.T) {
	v := NewValidator()

	for _, name := range []string{
		"all",
		"admin-panel",
		"admin",
	} {
		t.Run("allowed "+name, func(t *testing.T) {
			require.NoError(t, v.Var(name, "notreserved"))
		})
	}

	for _, name := range []string{
		"assets",
		"explore",
		"init",
		"healthcheck",
		"preview",
		"metrics",
		"mfa",
		"webauthn",
		"oauth",
		"login",
		"register",
		"logout",
		"settings",
		"search",
		"topics",
		"users",
	} {
		t.Run("reserved "+name, func(t *testing.T) {
			require.Error(t, v.Var(name, "notreserved"))
		})
	}
}

func TestUsernameRequiresAlphanumericCharacter(t *testing.T) {
	v := NewValidator()

	for _, name := range []string{"-", "---"} {
		require.Error(t, v.Var(name, "alphanumdash"))
	}
	for _, name := range []string{"-", "---", "_", "-_-"} {
		require.Error(t, v.Var(name, "alphanumdashunder"))
	}

	require.NoError(t, v.Var("user-name", "alphanumdash"))
	require.NoError(t, v.Var("user_name", "alphanumdashunder"))
}
