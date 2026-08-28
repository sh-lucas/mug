package router

import (
	_ "embed"
	"fmt"
	"go/ast"
	"log"
	"strings"

	"github.com/sh-lucas/mug/pkg"
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

var injectRouterSyntaxTmpl = pkg.Red + "Function %s needs to return (int, any), being 'any' the returned body after json marshalling" + pkg.Reset

func printInjectRouter(w *strings.Builder, path string, handler HandlerDecl) {
	// type checks
	results := handler.Fn.Type.Results.List
	identCode, frst := results[0].Type.(*ast.Ident)
	_, scnd := results[1].Type.(*ast.Ident)
	if !frst || !scnd || identCode.Name != "int" {
		log.Fatalf(injectRouterSyntaxTmpl, handler.Fn.Name)
	}

	// if the last comment contains middleware names, append as last arg
	mws := strings.Builder{}
	for _, mw := range getMiddlewares(handler.Doc) {
		fmt.Fprintf(&mws, "middlewares.%s, ", mw)
	}

	// code generated route path =)
	fmt.Fprintf(w, "fmt.Println(\"[%s] %s\")\n", handler.Fn.Name.Name, path)

	// code generated new router =)
	fmt.Fprintf(
		w, "spout.MakeHandler(router, \"%s\", %s.%s, %s)\n",
		path, handler.Package, handler.Fn.Name, mws.String(),
	)
}

func getMiddlewares(comment *ast.CommentGroup) []string {
	comments := comment.List
	last := comments[len(comments)-1].Text
	// verifies and cuts off the comment part
	if strings.HasPrefix(last, "// > ") || strings.HasPrefix(last, "//> ") {
		last = strings.TrimPrefix(last, "// >")
		last = strings.TrimPrefix(last, "//>")
		return strings.Split(last, " > ")
	}
	return []string{}
}
