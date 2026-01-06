package auth

import "testing"

func TestPasswordHashAndCompare(t *testing.T) {
	raw := "password123"

	hash, err := HashPassword(raw)
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}

	if ComparePassword(hash, raw) != nil {
		t.Fatal("password should match")
	}

	if ComparePassword(hash, "wrong") == nil {
		t.Fatal("wrong password should not match")
	}
}
