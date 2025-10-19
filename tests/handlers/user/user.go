package user

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sh-lucas/mug/pkg/mug"
)

// testing struct desserialization
type CreateUserInput struct {
	Username string `json:"username" validate:"required,min=6"`
}

type returnType struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message"`
	Role    string `json:"role,omitempty"`
}

// mug:handler POST /user/register
// > CoolMiddleware > FactLoggingMiddleware
func CreateUser(input CreateUserInput) (code int, body returnType) {
	if input.Username == "batman" {
		return 200, returnType{
			Message: "User created Sucessfully!",
			Role:    "admin",
		}
	}

	return 200, returnType{
		Error:   "Authorization error",
		Message: "User could not be created.",
	}
}

// input type ALIAS, notice the = sign
type PublishInput = mug.ShortBrew[PublishBody, PublishAuth]
type PublishBody struct {
	Text string `json:"text"`
}
type PublishAuth struct {
	Name string `json:"name"`
	jwt.RegisteredClaims
}

// mug:handler POST /rabbit
// > CoolMiddleware
func PublishToRabbit(ctx PublishInput) (code int, body any) {

	// need this to be valid on compile time =)
	// var A mug.Muggable = mug.Muggable(&ctx)

	return http.StatusAccepted, mug.M{
		"message": "ok",
		"greeting": fmt.Sprintf(
			"Hello, %s! Your message '%s' has been sent to the rabbit queue.",
			ctx.Auth.Name, ctx.Body.Text,
		),
	}
}
