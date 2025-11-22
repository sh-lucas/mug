package router

import (
	_ "embed"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sh-lucas/mug/internal/config"
	"github.com/sh-lucas/mug/internal/generator"
	"github.com/sh-lucas/mug/internal/helpers"
	"github.com/sh-lucas/mug/pkg"
)

//go:embed router.go.tmpl
var routerTemplate string

type HandlerDecl struct {
	Fn      *ast.FuncDecl
	Package string
	Doc     *ast.CommentGroup // Documentation comment
	Path    string
}

type genData struct {
	Handlers string
	Swagger  bool
}

func GenerateRouter() {
	decls, err := parseHandlersFolder()
	if err != nil {
		panic(err)
	}
	if len(decls) == 0 {
		if config.Global.Debug {
			helpers.Logf(pkg.Yellow + "⚠️  No handlers found. Skipping router generation." + pkg.Reset)
		}
		return
	}
	helpers.Logf("Generating router package")

	var content = &strings.Builder{}

	for _, handler := range decls {
		path, f := strings.CutPrefix(handler.Path, "// mug:handler ")
		if !f {
			path, f = strings.CutPrefix(handler.Path, "//mug:handler ")
			if !f {
				log.Fatalf("Invalid handler comment format: %s", handler.Path)
			}
		}
		// fmt.Printf(helpers.Yellow+"[%s] - %s%s\n"+helpers.Reset, handler.Fn.Name.Name, helpers.Cyan, path)

		handlerArgs := handler.Fn.Type.Params.List
		if len(handlerArgs) > 0 && isResponseWriter(handlerArgs[0]) {
			printBasicRouter(content, path, handler)
		} else {
			printInjectRouter(content, path, handler)
		}
	}

	data := genData{
		Handlers: content.String(),
		Swagger:  config.Global.Gen.Swagger,
	}

	err = generator.Generate(routerTemplate, data, "router", "router.go")
	if err != nil {
		panic(err)
	}
}

func isResponseWriter(field *ast.Field) bool {
	selector, ok := field.Type.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkgIdent, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}
	return pkgIdent.Name == "http" && selector.Sel.Name == "ResponseWriter"
}

func parseHandlersFolder() (decls []HandlerDecl, err error) {
	// gets the path of the handlers directory
	execPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Fail getting executable file path: %v", err)
	}

	handlersDir := filepath.Join(execPath, "handlers")

	// parse the subfolders
	helpers.Walk(handlersDir, func(filepath string) {
		handlerDecls, err := getCommentsFromFolder(filepath)
		if err != nil {
			log.Printf("Error parsing handler %s: %v", filepath, err)
		}

		decls = append(decls, handlerDecls...)
	})

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
		helpers.Logf("Found handler package %s", pkgName)
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
				Fn:      funcDecl,
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
