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
	Name    *ast.Ident // Name of the function
	Package string
	Doc     *ast.CommentGroup // Documentation comment
	Path    string
}

func GenerateRouter() {
	decls, err := parseHandlersFolder()
	if err != nil {
		panic(err)
	}
	if len(decls) == 0 {
		log.Println(global.Yellow + "⚠️  No handlers found. Skipping router generation." + global.Reset)
		return
	}
	global.Logf("Generating cup_router package")

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
		fmt.Fprintf(content, "http.HandleFunc(\"%s\", %s.%s)\n", path, handler.Package, handler.Name.Name)
	}

	Generate(routerTemplate, content, "router", "router.go")
}

func parseHandlersFolder() (decls []HandlerDecl, err error) {
	// gets the path of the handlers directory
	execPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Fail getting executable file path: %v", err)
	}

	// parse the /handlers folder
	handlersDir := filepath.Join(execPath, "handlers")
	decls, err = getCommentsFromFolder(handlersDir)
	if err != nil {
		log.Println("Could not parse /handlers")
	}

	// parse the subfolders
	subHandlers, err := os.ReadDir(handlersDir)
	if err != nil {
		panic(err)
	}

	for _, entry := range subHandlers {
		if !entry.IsDir() {
			continue
		}

		subHandler := filepath.Join(handlersDir, entry.Name())

		if !global.ValidatePath(subHandler) {
			continue // skips things the watcher is not tracking.
		}

		handlerDecls, err := getCommentsFromFolder(subHandler)
		if err != nil {
			log.Printf("Error parsing handler %s: %v", entry.Name(), err)
		}

		decls = append(decls, handlerDecls...)
	}

	return decls, err
}

func getCommentsFromFolder(handlersDir string) (decls []HandlerDecl, err error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, handlersDir, nil, parser.ParseComments)
	if err != nil || len(pkgs) != 1 {
		return []HandlerDecl{}, nil // empty folder, ignore it
	}

	// there should be 1 package inside
	for pkgName, pkg := range pkgs {
		global.Logf("Found package %s", pkgName)
		for _, file := range pkg.Files {
			// for every declaration in the file
			for _, decl := range file.Decls {
				// if the declaration is a function declaration
				if funcDecl, ok := decl.(*ast.FuncDecl); ok {
					ParseFunc(pkgName, funcDecl, &decls)
				}
			}
		}
	}
	return decls, nil
}

func ParseFunc(pkgName string, funcDecl *ast.FuncDecl, decls *[]HandlerDecl) {
	// skips functions without comments
	if funcDecl.Doc == nil || len(funcDecl.Doc.List) == 0 {
		return
	}

	for _, comment := range funcDecl.Doc.List {
		if verifyPrefix(comment.Text) {
			*decls = append(*decls, HandlerDecl{
				Name:    funcDecl.Name,
				Package: pkgName,
				Doc:     funcDecl.Doc,
				Path:    comment.Text,
			})
		}
	}
}

func verifyPrefix(comment string) bool {
	return strings.HasPrefix(comment, "// mug:handler") ||
		strings.HasPrefix(comment, "//mug:handler")
}
