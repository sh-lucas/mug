package handlers

import (
	"fmt"
	"net/http"

	cup "github.com/sh-lucas/mug/tests/mug_generated"
)

// mug:handler GET /
func CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Follow me on my website: %s", cup.MY_WEBSITE)
}
