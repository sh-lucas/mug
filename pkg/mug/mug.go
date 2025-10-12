package mug

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	jsoniter "github.com/json-iterator/go"
)

type M map[string]any

// Pour into a mug so your requests are always served hot! ☕️
type Muggable interface {
	Pour(w http.ResponseWriter, r *http.Request) (ok bool)
}

// simple and straightforward, just like a good cup of coffee.
// pass along the basic stuff from request using
type ShortBrew[AuthT any] struct {
	Writer  http.ResponseWriter `json:"-" bson:"-"`
	Request *http.Request       `json:"-" bson:"-"`
	Auth    AuthT               `json:"-" bson:"-"`
}

var missingTokenPayload = `{
	"error": "missing token",
	"message": "An Authorization header with a Bearer token is required to access this resource."
}`
var tokenExpiredPayload = `{
	"error": "token expired",
	"message": "The provided token has expired. Please authenticate again to obtain a new token."
}`
var invalidTokenPayload = `{
	"error": "invalid token",
	"message": "%s"
}`

var JWT_TOKEN_SECRET = (os.Getenv("JWT_TOKEN_SECRET"))

func (payload *ShortBrew[AuthT]) Pour(w http.ResponseWriter, r *http.Request) bool {
	payload.Writer = w
	payload.Request = r

	// body unmarshalling into struct for convenience =)
	_ = jsoniter.NewDecoder(r.Body).Decode(payload)

	// verifies if the Brew expects jwt claims
	claims, ok := any(&payload.Auth).(jwt.Claims)
	if !ok {
		// ignores because the handler is not expecting jwt claims
		return true
	}

	// AuthT parsing
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		http.Error(w, missingTokenPayload, http.StatusUnauthorized)
		return false
	}

	// Inicializar Auth com zero value
	var auth AuthT
	payload.Auth = auth

	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(JWT_TOKEN_SECRET), nil
	})

	if errors.Is(err, jwt.ErrTokenExpired) {
		http.Error(w, tokenExpiredPayload, http.StatusUnauthorized)
		return false
	} else if err != nil {
		http.Error(w, fmt.Sprintf(invalidTokenPayload, err.Error()), http.StatusUnauthorized)
		return false
	}

	return true
}
