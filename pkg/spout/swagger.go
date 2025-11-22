package spout

import (
	_ "embed"
	"net/http"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/invopop/jsonschema"
	jsoniter "github.com/json-iterator/go"
)

// Swagger Implementation

type RouteSpec struct {
	Method      string
	Path        string
	InputType   reflect.Type
	OutputType  reflect.Type
	Summary     string
	Description string
}

var registry []RouteSpec

//go:embed swagger.template.html
var swaggerTemplate []byte

// ServeDocs serves the Swagger UI and the generated OpenAPI spec.
func ServeDocs(r chi.Router) {
	// Serve swagger.json
	r.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		spec := generateOpenAPI()
		jsoniter.NewEncoder(w).Encode(spec)
	})

	// Serve Swagger UI
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(swaggerTemplate)
	})
}

func generateOpenAPI() *openapi3.T {
	spec := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Mug API",
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{},
		Components: &openapi3.Components{
			Schemas: make(openapi3.Schemas),
		},
	}

	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}

	for _, route := range registry {
		// Try to extract actual body type from JsonBodyT
		actualBodyType := extractBodyType(route.InputType)

		// Build request body
		var requestBody *openapi3.RequestBodyRef
		if actualBodyType != nil {
			// Generate schema for the extracted body type
			bodySchema := reflector.ReflectFromType(actualBodyType)

			// Convert to OpenAPI schema via JSON (preserves all fields)
			schemaBytes, _ := bodySchema.MarshalJSON()
			var schemaRef openapi3.SchemaRef
			_ = schemaRef.UnmarshalJSON(schemaBytes)

			requestBody = &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &schemaRef,
						},
					},
				},
			}
		} else if route.InputType.Kind() == reflect.Struct {
			// Fallback: use the input type directly if no JsonBodyT found
			inputSchema := reflector.ReflectFromType(route.InputType)

			// Convert to OpenAPI schema via JSON (preserves all fields)
			schemaBytes, _ := inputSchema.MarshalJSON()
			var schemaRef openapi3.SchemaRef
			_ = schemaRef.UnmarshalJSON(schemaBytes)

			requestBody = &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &schemaRef,
						},
					},
				},
			}
		}

		// Build responses
		responses := &openapi3.Responses{}
		if route.OutputType.Kind() == reflect.Struct {
			// Generate inline schema for output
			outputSchema := reflector.ReflectFromType(route.OutputType)

			// Convert to OpenAPI schema via JSON (preserves all fields)
			schemaBytes, _ := outputSchema.MarshalJSON()
			var schemaRef openapi3.SchemaRef
			_ = schemaRef.UnmarshalJSON(schemaBytes)

			responses.Set("200", &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Description: ptr("Successful response"),
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &schemaRef,
						},
					},
				},
			})
		} else if route.OutputType.Kind() == reflect.Interface {
			// For interface{} outputs, just show generic response
			responses.Set("200", &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Description: ptr("Successful response"),
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"object"},
								},
							},
						},
					},
				},
			})
		} else {
			// Default response if no struct
			responses.Set("200", &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Description: ptr("Successful response"),
				},
			})
		}

		// Create operation
		op := &openapi3.Operation{
			Summary:     route.Summary,
			Description: route.Description,
			RequestBody: requestBody,
			Responses:   responses,
		}

		// Get or create path item
		pathItem := spec.Paths.Find(route.Path)
		if pathItem == nil {
			pathItem = &openapi3.PathItem{}
			spec.Paths.Set(route.Path, pathItem)
		}

		// Assign operation to the correct method
		switch strings.ToUpper(route.Method) {
		case "GET":
			pathItem.Get = op
		case "POST":
			pathItem.Post = op
		case "PUT":
			pathItem.Put = op
		case "DELETE":
			pathItem.Delete = op
		case "PATCH":
			pathItem.Patch = op
		}
	}

	return spec
}

// extractBodyType attempts to extract the type parameter from JsonBodyT[T]
func extractBodyType(t reflect.Type) reflect.Type {
	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}

	// Look for embedded JsonBodyT or a direct Body field
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Check if this field is the Body from JsonBodyT (direct)
		if field.Name == "Body" && field.Tag.Get("json") == "body" {
			return field.Type
		}

		// Check if this is an embedded struct (like JsonBodyT)
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			// Recursively check the embedded struct for Body field
			if bodyType := extractBodyType(field.Type); bodyType != nil {
				return bodyType
			}
		}
	}

	return nil
}

func ptr[T any](v T) *T {
	return &v
}
