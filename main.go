package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sh-lucas/mug/generator"
	"github.com/sh-lucas/mug/global"
	"github.com/sh-lucas/mug/watcher"
)

func main() {
	godotenv.Read(".env")
	watcher.Start(rebuild)
}

// the process is already killed. Must return new process to track.
func rebuild() *exec.Cmd {
	// auto generates code
	generator.GenerateRouter()

	// prepares the statement
	cmd := exec.Command("go", "run", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	// if there is a .env, generate the code for it
	// and inject in the prepared process
	envs, err := godotenv.Read(".env")
	if err == nil {
		generator.GenerateEnvs(envs)
		for k := range envs {
			// injects in the format KEY=VALUE, hope this works well =)
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, envs[k]))
		}
	}

	// groups all processes and the current application
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		log.Printf(global.Red+"❌ Failed to start new process: %v", err)
	}

	return cmd
}
