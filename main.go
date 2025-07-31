package main

import (
	"os"

	"github.com/sh-lucas/mug/generator"
	"github.com/sh-lucas/mug/watcher"
)

func main() {
	args := os.Args
	if len(args) > 1 && args[1] == "watch" {
		watcher.Start()
	}

	generator.Generate()
}
