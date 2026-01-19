package auth

import (
	"testing"
	"time"
)

func TestGenerateJWT(t *testing.T) {
	InitJWT("test-secret", accessTTL, refreshTTL)

	token, err := GenerateJWT("user123", "user", 5*time.Minute, "")
	if err != nil {
		t.Fatalf("jwt generation failed: %v", err)
	}

	if token == "" {
		t.Fatal("token should not be empty")
	}
}
