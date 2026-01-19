// handlers/cursor.go
package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type taskCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
}

func encodeCursor(createdAt time.Time, id primitive.ObjectID) (string, error) {
	c := taskCursor{
		CreatedAt: createdAt.UTC(),
		ID:        id.Hex(),
	}
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func decodeCursor(cur string) (time.Time, primitive.ObjectID, error) {
	cur = strings.TrimSpace(cur)
	if cur == "" {
		return time.Time{}, primitive.NilObjectID, errors.New("empty cursor")
	}

	raw, err := base64.RawURLEncoding.DecodeString(cur)
	if err != nil {
		return time.Time{}, primitive.NilObjectID, errors.New("invalid cursor encoding")
	}

	var c taskCursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return time.Time{}, primitive.NilObjectID, errors.New("invalid cursor json")
	}

	oid, err := primitive.ObjectIDFromHex(c.ID)
	if err != nil {
		return time.Time{}, primitive.NilObjectID, errors.New("invalid cursor id")
	}

	return c.CreatedAt.UTC(), oid, nil
}
