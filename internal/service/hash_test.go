package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashService(t *testing.T) {
	service := NewHashService("secret")

	sum := service.Sha256sum([]byte("Some text"))

	assert.True(t, service.Matches([]byte("Some text"), sum))
	assert.False(t, service.Matches([]byte("Some another text"), sum))
}
