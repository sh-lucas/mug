package mug

import (
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type M map[string]any

// Pour into a mug so your requests are always served hot! ☕️
type Muggable interface {
	Pour(w http.ResponseWriter, r *http.Request) (ok bool)
}

// extends registeredClaims so you can't forget to
// add them on personalized claims
type Public struct {
	jwt.RegisteredClaims
}

// global vars

var JWT_TOKEN_SECRET = os.Getenv("JWT_TOKEN_SECRET")

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
