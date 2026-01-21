package integration

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"github.com/ajiraj2411/wunderlist-backend/internal/auth"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPasswordResetFlow(t *testing.T) {
	email := "reset-" + time.Now().UTC().Format("20060102150405") + "@test.com"
	password := "password123"

	signupUserWithPassword(t, email, password)

	// forgot password
	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password",
		bytes.NewBufferString(`{"email":"`+email+`"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("forgot password failed: %d %s", w.Code, w.Body.String())
	}

	// get user id
	userIDHex := getUserIDByEmail(t, email)
	uid, _ := primitive.ObjectIDFromHex(userIDHex)

	// create reset token (test helper)
	rawToken, err := auth.CreatePasswordResetTokenForTests(context.Background(), uid)
	if err != nil {
		t.Fatalf("failed to create reset token: %v", err)
	}

	// reset password
	payload := `{"token":"` + rawToken + `","new_password":"newpassword123"}`
	req = httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("reset password failed: %d %s", w.Code, w.Body.String())
	}

	// verify OLD password fails
	loginPayload := `{"email":"` + email + `","password":"password123"}`
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatal("old password should not work after reset")
	}

	// verify NEW password works
	loginPayload = `{"email":"` + email + `","password":"newpassword123"}`
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("new password should work, got %d %s", w.Code, w.Body.String())
	}
}
