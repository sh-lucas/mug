package gen_router

import (
	_ "embed"
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/sh-lucas/mug/global"
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

var injectRouterSyntaxTmpl = global.Red + "Function %s needs to return (int, any), being 'any' the returned body after json marshalling" + global.Reset

func printInjectRouter(w *strings.Builder, path string, handler HandlerDecl) {
	handlerArgs := handler.Fn.Type.Params.List
	argType := types.ExprString(handlerArgs[0].Type)

	// checks the return types if it makes sense
	if handler.Fn.Type.Results == nil || len(handler.Fn.Type.Results.List) != 2 {
		fmt.Printf(injectRouterSyntaxTmpl+"\n", handler.Fn.Name)
		os.Exit(1)
	}
	results := handler.Fn.Type.Results.List
	identCode, frst := results[0].Type.(*ast.Ident)
	_, scnd := results[1].Type.(*ast.Ident)
	if !frst || !scnd || identCode.Name != "int" {
		log.Fatalf(injectRouterSyntaxTmpl, handler.Fn.Name)
	}

	// truly generates the code using another template
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
