package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

var TestRouter *gin.Engine

func signupTestUser(t *testing.T) {
	payload := map[string]string{
		"email":    "int@test.com",
		"password": "password123",
	}

	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(
		http.MethodPost,
		"/signup",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	// 201 = created
	// 409 = already exists (acceptable for repeat runs)
	if w.Code != http.StatusCreated && w.Code != http.StatusConflict {
		t.Fatalf("signup failed: %s", w.Body.String())
	}
}
