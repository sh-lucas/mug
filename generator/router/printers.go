package gen_router

import (
	_ "embed"
	"fmt"
	"go/ast"
	"log"
	"strings"

	"github.com/sh-lucas/mug/global"
)

func printBasicRouter(w *strings.Builder, path string, handler HandlerDecl) {
	fmt.Fprintf(w, "router.HandleFunc(\"%s\", %s.%s)\n", path, handler.Package, handler.Fn.Name.Name)
}

type InjectRouterValues struct {
	Path    string
	Package string
	Type    string
	FnName  string
}

var injectRouterSyntaxTmpl = global.Red + "Function %s needs to return (int, any), being 'any' the returned body after json marshalling" + global.Reset

func printInjectRouter(w *strings.Builder, path string, handler HandlerDecl) {
	// type checks
	results := handler.Fn.Type.Results.List
	identCode, frst := results[0].Type.(*ast.Ident)
	_, scnd := results[1].Type.(*ast.Ident)
	if !frst || !scnd || identCode.Name != "int" {
		log.Fatalf(injectRouterSyntaxTmpl, handler.Fn.Name)
	}

	// code generated route path =)
	fmt.Fprintf(w, "fmt.Println(\"[%s] %s\")\n", handler.Fn.Name.Name, path)
	// code generated new router =)
	fmt.Fprintf(w, "keg.MakeHandler(router, \"%s\", %s.%s)", path, handler.Package, handler.Fn.Name)
}
