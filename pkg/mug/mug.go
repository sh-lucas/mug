package mug

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Pour into a mug so your requests are always served hot! ☕️
type Muggable interface {
	Pour(w http.ResponseWriter, r *http.Request) error
}

// simple and straightforward, just like a good cup of coffee.
// pass along the basic stuff from request using
type ShortBrew[AuthT any] struct {
	Writer        http.ResponseWriter
	Request       *http.Request
	Authorization jwt.Claims
}

func (lb *ShortBrew[T]) Pour(w http.ResponseWriter, r *http.Request) error {
	lb.Writer = w
	lb.Request = r

	// AuthT parsing
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		return errors.New("missing token")
	}

	_, err := jwt.ParseWithClaims(token, lb.Authorization, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_TOKEN_SECRET")), nil
	})
	if errors.Is(err, jwt.ErrTokenExpired) {
		return errors.New("token expired")
	} else if err != nil {
		return errors.New("invalid token")
	}

	// log.Println("ShortBrew Auth:", lb.Authorization)

	return nil
}
