package tests

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sh-lucas/mug/pkg/mug"
	"github.com/sh-lucas/mug/pkg/spout"
)

type TestPayload struct {
	mug.Espresso
	mug.BearerAuth
	mug.JsonBody
	Data string `json:"data"`
}

func TestComposition(t *testing.T) {
	// Setup
	mug.JWT_TOKEN_SECRET = "secret"
	handler := func(input TestPayload) (int, any) {
		if input.Data != "test" {
			return 500, "body not decoded"
		}
		if input.Claims.Subject != "user123" {
			return 401, "auth not parsed"
		}
		return 200, "ok"
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject: "user123",
	})
	tokenString, _ := token.SignedString([]byte("secret"))

	// Create request
	body, _ := json.Marshal(map[string]string{"data": "test"})
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	h := spout.ConvertHandler(handler)
	h.ServeHTTP(w, req)

	// Verify
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
