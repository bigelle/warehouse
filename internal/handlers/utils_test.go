package handlers_test

import (
	"testing"

	"github.com/bigelle/warehouse/internal/handlers"
	"github.com/bigelle/warehouse/internal/schemas"
	"github.com/stretchr/testify/require"
)

func TestIsAppropriateRole(t *testing.T) {
	// expect true:
	role := schemas.RoleAdmin
	ok := handlers.IsAppropriateRole(role, schemas.RoleAdmin)
	require.True(t, ok)

	// expect false
	role = schemas.RoleUser
	ok = handlers.IsAppropriateRole(role, schemas.RoleAdmin)
	require.False(t, ok)
}
