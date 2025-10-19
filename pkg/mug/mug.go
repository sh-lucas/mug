package mug

import (
	"net/http"
)

type M map[string]any

// Pour into a mug so your requests are always served hot! ☕️
type Muggable interface {
	Pour(w http.ResponseWriter, r *http.Request) (ok bool)
}

// simple and straightforward, just like a good cup of coffee.
// pass along the basic stuff from request using
type ShortBrew[AuthT any] struct {
	Writer  http.ResponseWriter `json:"-" bson:"-"`
	Request *http.Request       `json:"-" bson:"-"`
	Auth    AuthT               `json:"-" bson:"-"`
}
