package generator

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
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
	fmt.Println(global.Green + "✅ Injecting .env file" + global.Reset)
	for k := range envs {
		// gets the env as a variable
		fmt.Fprintf(content, "var %s = os.Getenv(\"%s\")\n", k, k)
	}

	Generate(envsTemplate, content, "", "envs.go")
	lastEnvUpdate = time.Now()
}
