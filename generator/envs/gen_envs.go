package gen_envs

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sh-lucas/mug/generator"
	"github.com/sh-lucas/mug/global"
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

	global.Logf("Generating envs package")

	var content = &strings.Builder{}
	for k := range envs {
		// gets the env as a variable
		fmt.Fprintf(content, "var %s = os.Getenv(\"%s\")\n", k, k)
	}

	generator.Generate(envsTemplate, content.String(), "", "envs.go")
	lastEnvUpdate = time.Now()
}
