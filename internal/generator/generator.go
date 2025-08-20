package generator

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"golang.org/x/tools/imports"
)

var outputDir = "mug_generated"

func Generate(templ string, input any, subdir, fileName string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}
	// Copies ../template/generated.go.tmpl to mug_generated/generated.go
	if subdir != "" {
		// creates the subdirectory if it does not exist
		if err := os.MkdirAll(filepath.Join(outputDir, subdir), 0755); err != nil {
			return fmt.Errorf("failed to create subdirectory: %v", err)
		}
		fileName = filepath.Join(subdir, fileName)
	}
	generatedFilePath := filepath.Join(outputDir, fileName)

	// uses text/template to generate the new file content
	tmpl, err := template.New("generated.go.tmpl").Parse(templ)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// uses the template to generate the content and formats it.
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, input); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}
	codigoFormatado, err := imports.Process(fileName, buf.Bytes(), nil)
	if err != nil {
		log.Printf("Failed code:\n %s\n", buf.String())
		return fmt.Errorf("failed to format code: %v", err)
	}

	// write out (needs permissions to write)
	err = os.WriteFile(generatedFilePath, codigoFormatado, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	return nil
}
