package domain

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID returns a 32-char hex ID. It panics on entropy failure, which
// is acceptable because the only thing to do in that case is crash.
func NewID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("rand.Read: " + err.Error())
	}
	return hex.EncodeToString(b)
}
