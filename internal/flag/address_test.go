package flag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAddress_Valid(t *testing.T) {
	addr := NewRunAddress()
	err := addr.Set("localhost:5432")

	require.NoError(t, err)
	assert.Equal(t, 5432, addr.Port)
	assert.Equal(t, "localhost", addr.Host)
}

func TestSetAddress_Invalid(t *testing.T) {
	addr := NewRunAddress()
	err := addr.Set("localhost:")

	require.Error(t, err)
	assert.Equal(t, DefaultPort, addr.Port)
	assert.Equal(t, DefaultHost, addr.Host)
}

func TestAddress_ToString(t *testing.T) {
	addr := NewRunAddress()
	assert.Equal(t, "localhost:8080", addr.String())
}
