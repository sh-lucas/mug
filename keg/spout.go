package keg

import (
	"encoding/json"
	"errors"
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
)

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

func MakeHandler[T any, U any](r chi.Router, path string, handler func(input T) (code int, body U)) {
	r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		// guarantees valid response
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, internalErrorMsg, 500)
				return
			}
		}()

		// unmarshal into T and check if something is missing.
		// errors are ignored
		var payload T
		_ = jsoniter.NewDecoder(r.Body).Decode(&payload)

		err := validate.Struct(&payload)
		if err != nil {
			errMsg := formatValidationErrors(err, translator)
			w.WriteHeader(400)
			w.Write(errMsg)
			return
		}

		code, body := handler(payload)
		w.WriteHeader(code)
		err = jsoniter.NewEncoder(w).Encode(body)
		if err != nil {
			log.Println("Unsmarshable content returned from handler!")
			http.Error(w, internalErrorMsg, 500)
			return
		}
	})
}

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
	jsonResponse, _ := json.Marshal(response)
	return jsonResponse
}
