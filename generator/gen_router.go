package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sh-lucas/mug/global"
)

type HandlerDecl struct {
	Name *ast.Ident        // Name of the function
	Doc  *ast.CommentGroup // Documentation comment
	Path string
}

func GenerateRouter() {
	decls := parseHandlersFolder()

	var content = &strings.Builder{}

	for _, handler := range decls {
		path, f := strings.CutPrefix(handler.Path, "// mug:handler ")
		if !f {
			path, f = strings.CutPrefix(handler.Path, "//mug:handler ")
			if !f {
				log.Fatalf("Invalid handler comment format: %s", handler.Path)
			}
		}
		fmt.Printf("%s[%s] - %s%s\n%s", global.Yellow, handler.Name.Name, global.Cyan, path, global.Reset)
		fmt.Fprintf(content, "http.HandleFunc(\"%s\", handlers.%s)\n", path, handler.Name.Name)
	}

	Generate(routerTemplate, content, "router.go")

}

func parseHandlersFolder() (decls []HandlerDecl) {
	// gets the path of the handlers directory
	execPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Falha ao obter o caminho do executável: %v", err)
	}
	handlersDir := filepath.Join(execPath, "handlers")
	return getCommentsFromFolder(handlersDir)
}

func getCommentsFromFolder(handlersDir string) (decls []HandlerDecl) {

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, handlersDir, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Erro ao analisar o diretório de handlers: %v", err)
	}
	if len(pkgs) == 0 || pkgs["handlers"] == nil {
		log.Fatal("Package 'handlers' not found.")
	}

	pkg := pkgs["handlers"]

	// for every file on the package
	for _, file := range pkg.Files {
		// for every declaration in the file
		for _, decl := range file.Decls {
			// if the declaration is a function declaration
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				// if it has a comment
				if funcDecl.Doc != nil && len(funcDecl.Doc.List) > 0 {
					for _, comment := range funcDecl.Doc.List {
						if strings.HasPrefix(comment.Text, "// mug:handler") ||
							strings.HasPrefix(comment.Text, "//mug:handler") {
							decls = append(decls, HandlerDecl{
								Name: funcDecl.Name,
								Doc:  funcDecl.Doc,
								Path: comment.Text,
							})
						}
					}
				}
			}
		}
	}
	return decls
}
