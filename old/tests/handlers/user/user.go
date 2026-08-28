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
	Message string `json:"message"`
	Role    string `json:"role,omitempty"`
	Error   string `json:"error,omitempty"`
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

// input type ALIAS, notice the = sign, now using Composition!
type PublishInput struct {
	mug.JsonBody[PublishBody]
	mug.Auth // Alias for BearerAuth[jwt.RegisteredClaims]
}

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
	// Acessando Body e Claims diretamente através da composição da struct
	return http.StatusAccepted, mug.M{
		"message": "ok",
		"greeting": fmt.Sprintf(
			"Hello, %s! Your message '%s' has been sent to the rabbit queue.",
			ctx.Claims.Subject, ctx.Body.Text, // Note: using Subject as Name is in claims, but for this example I'll stick to what standard claims offer or need custom auth
		),
	}
}

type PourInput struct {
	mug.JsonBody[struct {
		Owner      string `json:"owner"`
		CoffeeType string `json:"coffee_type"`
	}]
	// mug.BearerAuth[struct {
	// 	Name string `json:"name"`
	// 	jwt.RegisteredClaims
	// }]
}

// mug:handler GET /coffee
func PourSomeCoffee(input PourInput) (code int, body any) {
	return http.StatusOK, mug.M{
		"message": "coffee ready!",
		"owner":   input.Body.Owner,
	}
}
