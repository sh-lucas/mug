package spout

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	jsoniter "github.com/json-iterator/go"
	"github.com/sh-lucas/mug/pkg"
	"github.com/sh-lucas/mug/pkg/mug"
)

type kegHandler[T any, U any] func(input T) (code int, body U)

type middleware func(http.Handler) http.Handler

// validator v10 initialized
var validate = validator.New(validator.WithRequiredStructEnabled())

var translator ut.Translator

func init() {
	// setup validator json parser
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	english := en.New()
	uni := ut.New(english, english)
	translator, _ = uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, translator)
}

var internalErrorMsg = `{
	"error": "Internal server error",
	"message": "The issue must be reported to the system administrator."
}`

// Defines a new kegHandler in r (router), at path, with middlewares before handler.
func MakeHandler[T any, U any](
	r chi.Router,
	path string, handler func(input T) (code int, body U),
	middlewares ...middleware,
) {
	chained := chain(middlewares, ConvertHandler(handler))

	r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		// crash recovery
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf(pkg.Red+"panic: %v\n"+pkg.Reset, r)
				http.Error(w, internalErrorMsg, 500)
				return
			}
		}()
		chained.ServeHTTP(w, r)
	})

	// Register for Swagger
	parts := strings.SplitN(path, " ", 2)
	method := "GET"
	url := path
	if len(parts) == 2 {
		method = parts[0]
		url = parts[1]
	}

	registry = append(registry, RouteSpec{
		Method:     method,
		Path:       url,
		InputType:  reflect.TypeOf((*T)(nil)).Elem(),
		OutputType: reflect.TypeOf((*U)(nil)).Elem(),
	})
}

// chain chains middlewares before a finalHandler.
func chain(middlewares []middleware, finalHandler http.Handler) http.Handler {
	// If there are no middlewares, just return the final handler.
	if len(middlewares) == 0 {
		return finalHandler
	}

	// Start with the final handler as the innermost item.
	wrapped := finalHandler

	// Loop backwards through the middlewares, wrapping the handler.
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}

	return wrapped
}

// converts a personalized handler (kegHandler) to an http.Handler
func ConvertHandler[T any, U any](handler kegHandler[T, U]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		// unmarshal into T and check if something is missing.
		// errors are ignored because of the validation that do it's job.
		var payload T

		// if it is muggable, pour it!
		// responsible for auth, unmarshalling, etc.
		if m, ok := any(&payload).(mug.Muggable); ok {
			ok := m.Pour(w, r, &payload)
			if !ok {
				return
			}
		} else {
			// if not muggable, simply try to unmarshal the body
			_ = jsoniter.NewDecoder(r.Body).Decode(&payload)
		}

		// validation happens after pouring =)
		err := validate.Struct(&payload)
		if err != nil {
			errMsg := formatValidationErrors(err, translator)
			http.Error(w, string(errMsg), http.StatusBadRequest)
			return
		}

		code, body := handler(payload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		// marshal response
		err = jsoniter.NewEncoder(w).Encode(body)
		if err != nil {
			log.Println("Unsmarshable content returned from handler!")
			http.Error(w, internalErrorMsg, http.StatusInternalServerError)
			return
		}
	})
}

// formats validation errors as json and marshals it
// so your api is easy to consume.
func formatValidationErrors(err error, trans ut.Translator) []byte {

	response := make(map[string]string)
	var validationErrors validator.ValidationErrors

	if errors.As(err, &validationErrors) {
		for _, fieldErr := range validationErrors {
			response[fieldErr.Field()] = fieldErr.Translate(trans)
		}
	} else {
		response["error"] = "invalid input provided"
	}

	jsonResponse, _ := jsoniter.Marshal(response)
	return jsonResponse
}
