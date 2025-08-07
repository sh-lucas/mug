package generator

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/sh-lucas/mug/global"
)

//go:embed envs.go.tmpl
var envsTemplate string

func GenerateEnvs(envs map[string]string) {

	var content = &strings.Builder{}
	fmt.Println(global.Green + "✅ Loaded .env file" + global.Reset)
	for k := range envs {
		// gets the env as a variable
		fmt.Fprintf(content, "var %s = os.Getenv(\"%s\")\n", k, k)
	}

	Generate(envsTemplate, content, "envs.go")
}
