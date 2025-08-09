package user

import (
	"fmt"
	"net/http"
)

// mug:handler POST /user/login
func TestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "User logged in")
}
