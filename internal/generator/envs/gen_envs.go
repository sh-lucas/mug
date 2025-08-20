package envs

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sh-lucas/mug/internal/generator"
	"github.com/sh-lucas/mug/internal/helpers"
)

//go:embed envs.go.tmpl
var envsTemplate string

var lastEnvUpdate time.Time

func GenerateEnvs() {
	envs, err := godotenv.Read(".env")
	if err != nil {
		return // do nothing
	}

	info, err := os.Stat(".env")
	if err != nil || lastEnvUpdate.After(info.ModTime()) {
		return
	}

	helpers.Logf("Generating envs package")

	var content = &strings.Builder{}
	for k := range envs {
		// gets the env as a variable
		fmt.Fprintf(content, "var %s = os.Getenv(\"%s\")\n", k, k)
	}

	generator.Generate(envsTemplate, content.String(), "", "envs.go")
	lastEnvUpdate = time.Now()
}
