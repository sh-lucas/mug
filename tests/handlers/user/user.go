package user

import (
	"net/http"

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

type Message struct {
	Text string `json:"text"`
	mug.ShortBrew[any]
}

// mug:handler POST /rabbit
// > CoolMiddleware
func PublishToRabbit(message Message) (code int, body any) {

	// rabbit.Send("test", message)

	return http.StatusAccepted, map[string]string{
		"message": "ok",
	}
}
