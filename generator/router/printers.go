package gen_router

import (
	_ "embed"
	"fmt"
	"go/types"
	"log"
	"strings"
	"text/template"
)

func printBasicRouter(w *strings.Builder, path string, handler HandlerDecl) {
	fmt.Fprintf(w, "http.HandleFunc(\"%s\", %s.%s)\n", path, handler.Package, handler.Fn.Name.Name)
}

//go:embed inject_router.go.tmpl
var injectRouterTmpl string

type InjectRouterValues struct {
	Path    string
	Package string
	Type    string
	FnName  string
}

func printInjectRouter(w *strings.Builder, path string, handler HandlerDecl) {
	handlerArgs := handler.Fn.Type.Params.List
	argType := types.ExprString(handlerArgs[0].Type)

	tmpl, err := template.New("injectRouter").Parse(injectRouterTmpl)
	if err != nil {
		log.Fatalf("Template parse error: %v\n", err)
	}

	err = tmpl.Execute(w, InjectRouterValues{
		Path:    path,
		Package: handler.Package,
		Type:    argType,
		FnName:  handler.Fn.Name.Name,
	})
	if err != nil {
		log.Fatalf("Template execution error: %v", err)
	}
}
