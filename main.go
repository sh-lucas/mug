package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/sh-lucas/mug/generator"
	"github.com/sh-lucas/mug/watcher"
)

func main() {
	watcher.Start(rebuild)
}

// kills the process and rebuilds the application
func rebuild() *exec.Cmd {
	cmd := exec.Command("go", "run", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// groups all processes and the current application
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// auto generates code
	generator.Generate()

	if err := cmd.Start(); err != nil {
		log.Printf("❌ Failed to start new process: %v", err)
		return cmd
	}

	return cmd
}
