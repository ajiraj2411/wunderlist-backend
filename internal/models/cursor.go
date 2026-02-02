package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrInvalidCursor = errors.New("invalid cursor")

const cursorVersion = 1

var cursorSigningSecret = []byte(getEnv(
	"CURSOR_SIGNING_SECRET",
	"dev-insecure-cursor-secret",
))

// =======================
// SHARED CURSOR STRUCTS
// =======================

type CursorPayload struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
}

type SignedCursor struct {
	Version int           `json:"v"`
	Payload CursorPayload `json:"payload"`
	Sig     string        `json:"sig"`
}

// =======================
// ENCODE
// =======================

func EncodeCursor(
	createdAt time.Time,
	id primitive.ObjectID,
	userID primitive.ObjectID,
) (string, error) {

	payload := CursorPayload{
		CreatedAt: createdAt.UTC(),
		ID:        id.Hex(),
		UserID:    userID.Hex(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	sig := signCursor(payloadBytes)

	wrapped := SignedCursor{
		Version: cursorVersion,
		Payload: payload,
		Sig:     sig,
	}

	finalBytes, err := json.Marshal(wrapped)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(finalBytes), nil
}

// =======================
// DECODE + VERIFY
// =======================

func DecodeCursor(
	raw string,
	expectedUserID primitive.ObjectID,
) (time.Time, primitive.ObjectID, error) {

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, primitive.NilObjectID, ErrInvalidCursor
	}

	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return time.Time{}, primitive.NilObjectID, ErrInvalidCursor
	}

	var wrapped SignedCursor
	if err := json.Unmarshal(decoded, &wrapped); err != nil {
		return time.Time{}, primitive.NilObjectID, ErrInvalidCursor
	}

	if wrapped.Version != cursorVersion {
		return time.Time{}, primitive.NilObjectID, ErrInvalidCursor
	}

	payloadBytes, err := json.Marshal(wrapped.Payload)
	if err != nil {
		return time.Time{}, primitive.NilObjectID, ErrInvalidCursor
	}

	if !verifyCursor(payloadBytes, wrapped.Sig) {
		return time.Time{}, primitive.NilObjectID, ErrInvalidCursor
	}

	if wrapped.Payload.UserID != expectedUserID.Hex() {
		return time.Time{}, primitive.NilObjectID, ErrInvalidCursor
	}

	oid, err := primitive.ObjectIDFromHex(wrapped.Payload.ID)
	if err != nil {
		return time.Time{}, primitive.NilObjectID, ErrInvalidCursor
	}

	if wrapped.Payload.CreatedAt.IsZero() {
		return time.Time{}, primitive.NilObjectID, ErrInvalidCursor
	}

	return wrapped.Payload.CreatedAt.UTC(), oid, nil
}

// =======================
// HMAC HELPERS
// =======================

func signCursor(data []byte) string {
	mac := hmac.New(sha256.New, cursorSigningSecret)
	mac.Write(data)
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func verifyCursor(data []byte, sig string) bool {
	expected := signCursor(data)
	return hmac.Equal([]byte(expected), []byte(sig))
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
