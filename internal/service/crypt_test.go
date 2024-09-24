package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptData(t *testing.T) {
	data := []byte("secret message")
	tests := []struct {
		name     string
		givenKey string
		wantErr  string
	}{
		{
			name:    "NoKey",
			wantErr: "error reading public key",
		},
		{
			name:     "InvalidKey",
			givenKey: "build_info.tpl",
			wantErr:  "failed to parse PEM block",
		},
		{
			name:     "OK",
			givenKey: "../../test_data/client.pem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := EncryptData(data, tt.givenKey)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Empty(t, encoded)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, encoded)
			}
		})
	}
}

func TestDecryptData(t *testing.T) {
	publicKeyPath := "../../test_data/client.pem"
	privateKeyPath := "../../test_data/server.pem"
	msg := "secret message"

	encrypted, err := EncryptData([]byte(msg), publicKeyPath)
	require.NoError(t, err)

	decrypted, err := DecryptData(encrypted, privateKeyPath)
	require.NoError(t, err)
	assert.Equal(t, []byte("secret message"), decrypted)
}
