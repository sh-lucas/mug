package main

import (
	"flag"
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

var codeGen = flag.Bool("gen", false, "Enables code generation")
var injEnvs = flag.Bool("env", true, "Disables .env file injection")

func main() {
	flag.Parse()
	watcher.Start(rebuild)
}

// the process is already killed. Must return new process to track.
func rebuild() *exec.Cmd {
	// prepares the statement nicely (lower priority)
	cmd := exec.Command("nice", "-n", "15", "go", "run", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// groups the processes
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// injects environment variables priorizating .env file (as last)
	cmd.Env = os.Environ()
	envs, err := godotenv.Read(".env")
	if *injEnvs && err == nil {
		for k, v := range envs {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	// generates code for .env and handlers
	if *codeGen {
		generateCode(envs)
	}

	// runs "go run ."
	if err := cmd.Start(); err != nil {
		log.Printf(global.Red+"❌ Failed to start new process: %v", err)
	}

	return cmd
}

var lastEnvUpdate time.Time

// logic for generating all the code before executing the command
func generateCode(envs map[string]string) {
	// auto generates code
	generator.GenerateRouter()

	// avoid rebuilding this if it weren't modified.
	info, err := os.Stat(".env")

	if err != nil || lastEnvUpdate.After(info.ModTime()) {
		return
	}

	generator.GenerateEnvs(envs)
	lastEnvUpdate = time.Now()
}
