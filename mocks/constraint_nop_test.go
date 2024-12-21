package mocks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNopConstraint_GetName(t *testing.T) {
	c := &nopConstraint{}
	got := c.GetName()
	require.Equal(t, "nop", got, "GetName() should return 'nop'")
}

func TestNopConstraint_Verify(t *testing.T) {
	c := &nopConstraint{}
	got := c.Verify(nil)
	require.Nil(t, got, "Verify() should return nil")
}
