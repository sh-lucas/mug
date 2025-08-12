package user

import (
	"log"
)

// testing struct desserialization
type CreateUserInput struct {
	Username string `json:"username"`
}

// mug:handler POST /user/register
func CreateUser(input CreateUserInput) {
	log.Println(input.Username)
}
