package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func refreshTokenSHA(raw string) string {
	raw = strings.TrimSpace(raw)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
