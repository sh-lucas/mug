package user

import (
	"fmt"
	"net/http"
)

// mug:handler POST /user/login
func CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "User logged in")
}
