package generator

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

type HandlerDecl struct {
	Name *ast.Ident        // Name of the function
	Doc  *ast.CommentGroup // Documentation comment
	Path string
}

//go:embed generated.go.tmpl
var templateContent string

func Generate() {
	decls := parseHandlersFolder()

	var output = &strings.Builder{}

	for _, handler := range decls {
		path, _ := strings.CutPrefix(handler.Path, "// mug:handler ")
		fmt.Println("Path:   ", path)
		fmt.Println("Handler:", handler.Name.Name)

		fmt.Fprintf(output, "http.HandleFunc(\"%s\", handlers.%s)\n", path, handler.Name.Name)
	}

	// Create the output directory if it doesn't exist
	outputDir := "mug_generated"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	// Copies ../template/generated.go.tmpl to mug_generated/generated.go
	generatedFilePath := filepath.Join(outputDir, "generated.go")

	// uses text/template to generate the new file content
	tmpl, err := template.New("generated.go.tmpl").Parse(templateContent)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, output.String()); err != nil {
		log.Fatalf("Falha ao executar o template: %v", err)
	}

	codigoFormatado, err := imports.Process("generated.go", buf.Bytes(), nil)
	if err != nil {
		log.Fatalf("Falha ao formatar o código: %v", err)
	}

	err = os.WriteFile(generatedFilePath, codigoFormatado, 0644)
	if err != nil {
		log.Fatalf("Falha ao escrever o arquivo: %v", err)
	}

	log.Println("Arquivo 'generated.go' criado com sucesso.")
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
						if strings.HasPrefix(comment.Text, "// mug:handler") {
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
