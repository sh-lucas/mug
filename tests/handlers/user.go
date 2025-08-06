package handlers

import (
	"fmt"
	"net/http"
)

// mug:handler POST /user/create
func CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Test function in generator package")
}
