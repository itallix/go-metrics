package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

type HashService interface {
	Sha256sum(msg []byte) string
	Matches(msg []byte, hash string) bool
}

type HashServiceImpl struct {
	secret []byte
}

func NewHashService(secret string) *HashServiceImpl {
	return &HashServiceImpl{
		secret: []byte(secret),
	}
}

func (s *HashServiceImpl) Sha256sum(msg []byte) string {
	h := hmac.New(sha256.New, s.secret)
	h.Write(msg)
	sha256Sum := h.Sum(nil)
	return hex.EncodeToString(sha256Sum)
}

func (s *HashServiceImpl) Matches(msg []byte, hash string) bool {
	h := hmac.New(sha256.New, s.secret)
	h.Write(msg)
	sha256Sum := h.Sum(nil)
	encoded := hex.EncodeToString(sha256Sum)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(encoded)) == 1
}
