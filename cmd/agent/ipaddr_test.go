package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalIP(t *testing.T) {
	localIP := GetLocalIP()
	fmt.Printf("Local IP: %s", localIP)
	assert.NotEqual(t, "127.0.0.1", localIP)
}
