package mug

import (
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

type Espresso struct {
	C struct {
		Writer  http.ResponseWriter `json:"-" bson:"-"`
		Request *http.Request       `json:"-" bson:"-"`
	}
}

func (e *Espresso) Pour(w http.ResponseWriter, r *http.Request, parent any) bool {
	e.C.Writer = w
	e.C.Request = r

	// Handle JSON Body
	if bodyable, ok := parent.(Bodyable); ok {
		bodyPtr := bodyable.GetBodyPtr()
		err := jsoniter.NewDecoder(r.Body).Decode(bodyPtr)
		if err != nil {
			fmt.Println("Error decoding body:", err)
		}
	}

	// Handle Authentication
	if auth, ok := parent.(Authable); ok {
		if !auth.Authenticate(w, r) {
			return false
		}
	}

	return true
}
