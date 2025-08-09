package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sh-lucas/mug/generator"
	"github.com/sh-lucas/mug/global"
	"github.com/sh-lucas/mug/watcher"
)

func main() {
	watcher.Start(rebuild)
}

// the process is already killed. Must return new process to track.
func rebuild() *exec.Cmd {
	// prepares the statement nicy (lower priority)
	cmd := exec.Command("nice", "-n", "15", "go", "run", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// groups the processes
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// before generateCode() so then it overrides it
	cmd.Env = os.Environ()

	generateCode(cmd)

	if err := cmd.Start(); err != nil {
		log.Printf(global.Red+"❌ Failed to start new process: %v", err)
	}

	return cmd
}

var lastEnvUpdate time.Time

// logic for generating all the code before executing the command
func generateCode(cmd *exec.Cmd) {
	// auto generates code
	generator.GenerateRouter()

	// avoid rebuilding this if it weren't modified.
	info, err := os.Stat(".env")
	if err != nil || info.ModTime().After(lastEnvUpdate) {
		return
	}

	envs, err := godotenv.Read(".env")
	if err == nil {
		generator.GenerateEnvs(envs)
		for k := range envs {
			// injects in the format KEY=VALUE, hope this works well =)
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, envs[k]))
		}
		lastEnvUpdate = time.Now()
	}
}
