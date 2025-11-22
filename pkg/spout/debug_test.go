// small test to debug extractBodyType
// mostly used by AI 😳

package spout

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/invopop/jsonschema"
	"github.com/sh-lucas/mug/pkg/mug"
)

type TestBody struct {
	Owner      string `json:"owner"`
	CoffeeType string `json:"coffee_type"`
}

type TestInput struct {
	mug.JsonBodyT[TestBody]
}

type TestInputAnonymous struct {
	mug.JsonBodyT[struct {
		Owner      string `json:"owner"`
		CoffeeType string `json:"coffee_type"`
	}]
}

func TestExtractBodyType(t *testing.T) {
	tests := []struct {
		name string
		typ  reflect.Type
	}{
		{"Named Struct", reflect.TypeOf(TestInput{})},
		{"Anonymous Struct", reflect.TypeOf(TestInputAnonymous{})},
		{"ShortBrew", reflect.TypeOf(mug.ShortBrew[TestBody, jwt.RegisteredClaims]{})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("Inspecting %s\n", tt.name)
			got := extractBodyType(tt.typ)
			if got == nil {
				t.Errorf("extractBodyType() returned nil")
			} else {
				fmt.Printf("Found body type: %v\n", got)
				// Check fields of the found type
				if got.Kind() == reflect.Struct {
					for i := 0; i < got.NumField(); i++ {
						fmt.Printf("  Field: %s %s\n", got.Field(i).Name, got.Field(i).Type)
					}
				}

				// Test Schema Generation
				reflector := jsonschema.Reflector{
					AllowAdditionalProperties: false,
					DoNotReference:            true,
				}
				schema := reflector.Reflect(got)
				jsonBytes, _ := schema.MarshalJSON()
				fmt.Printf("Schema JSON: %s\n", string(jsonBytes))
			}
		})
	}
}
