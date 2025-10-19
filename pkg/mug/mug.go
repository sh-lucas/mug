package mug

import (
	"net/http"
)

type M map[string]any

// Pour into a mug so your requests are always served hot! ☕️
type Muggable interface {
	Pour(w http.ResponseWriter, r *http.Request) (ok bool)
}
