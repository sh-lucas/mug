package mug

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	jsoniter "github.com/json-iterator/go"
)

// simple and straightforward, just like a good cup of coffee.
// pass along the basic stuff from request using
type ShortBrew[BodyT any, AuthT jwt.Claims] struct {
	Writer  http.ResponseWriter `json:"-" bson:"-"`
	Request *http.Request       `json:"-" bson:"-"`
	Auth    AuthT               `json:"-" bson:"-"`
	Body    BodyT               `json:"-" bson:"-"`
}

func (payload *ShortBrew[BodyT, AuthT]) Pour(w http.ResponseWriter, r *http.Request) bool {
	payload.Writer = w
	payload.Request = r

	// body unmarshalling into struct for convenience =)
	err := jsoniter.NewDecoder(r.Body).Decode(&((*payload).Body))
	if err != nil {
		fmt.Println("Error decoding body:", err)
	}

	// Public implements jwt.Claims, but here we ignore auth for it
	_, isPublic := any(&payload.Auth).(Public)
	if isPublic {
		return true
	}

	// by type definition, AuthT must implement jwt.Claims =)
	claims, _ := any(&payload.Auth).(jwt.Claims)

	// AuthT parsing
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		http.Error(w, missingTokenPayload, http.StatusUnauthorized)
		return false
	}

	// Inicializar Auth com zero value
	var auth AuthT
	payload.Auth = auth

	_, err = jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
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
