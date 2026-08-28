package mug

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

// Authable interface for structs that handle authentication
// Authable interface for structs that handle authentication
// Authable interface for structs that handle authentication
type Authable interface {
	Authenticate(w http.ResponseWriter, r *http.Request) bool
}

// Auth is a convenience alias for BearerAuth with default RegisteredClaims
type Auth = BearerAuth[jwt.RegisteredClaims]

// BearerAuth mixin for Bearer token authentication with custom claims
type BearerAuth[T jwt.Claims] struct {
	Claims T
}

var validate = validator.New()

func (b *BearerAuth[T]) Authenticate(w http.ResponseWriter, r *http.Request) bool {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		http.Error(w, missingTokenPayload, http.StatusUnauthorized)
		return false
	}

	// Initialize Claims if it's a pointer or zero value
	// For simple structs, we can just use &b.Claims
	// But we need to be careful about initialization if T is a pointer type

	// Parse token
	_, err := jwt.ParseWithClaims(token, any(&b.Claims).(jwt.Claims), func(t *jwt.Token) (interface{}, error) {
		return []byte(JWT_TOKEN_SECRET), nil
	})

	if errors.Is(err, jwt.ErrTokenExpired) {
		http.Error(w, tokenExpiredPayload, http.StatusForbidden)
		return false
	} else if err != nil {
		http.Error(w, fmt.Sprintf(invalidTokenPayload, err.Error()), http.StatusUnauthorized)
		return false
	}

	// Validate claims
	if err := validate.Struct(b); err != nil {
		http.Error(w, fmt.Sprintf(invalidTokenPayload, err.Error()), http.StatusUnauthorized)
		return false
	}

	return true
}

// Bodyable interface for structs that handle JSON body
type Bodyable interface {
	GetBodyPtr() any
}

// JsonBody mixin for JSON request body with type parameter
type JsonBody[T any] struct {
	Body T `json:"body"`
}

func (j *JsonBody[T]) GetBodyPtr() any {
	return &j.Body
}
