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
)

func Generate() {
	comments := parseHandlersFolder()
	for _, comment := range comments {
		fmt.Println("Handler: ", comment)
	}
}

func parseHandlersFolder() []string {
	var comments []string

	// gets the path of the handlers directory
	execPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Falha ao obter o caminho do executável: %v", err)
	}
	handlersDir := filepath.Join(execPath, "handlers")

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, handlersDir, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Erro ao analisar o diretório de handlers: %v", err)
	}
	if len(pkgs) == 0 || pkgs["handlers"] == nil {
		log.Fatal("Package 'handlers' not found.")
	}

	pkg := pkgs["handlers"]

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				if funcDecl.Doc != nil && len(funcDecl.Doc.List) > 0 {
					for _, comment := range funcDecl.Doc.List {
						if strings.HasPrefix(comment.Text, "// mug:handler") {
							comments = append(comments, comment.Text)
						}
					}
				}
			}
		}
	}

	return comments
}
