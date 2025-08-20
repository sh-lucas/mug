package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sh-lucas/mug/internal/generator/envs"
	"github.com/sh-lucas/mug/internal/generator/router"
	"github.com/sh-lucas/mug/internal/helpers"
	"github.com/sh-lucas/mug/internal/watcher"
)

func main() {
	if helpers.Config.Features.AutoTidy {
		_ = exec.Command("go", "mod", "tidy").Run()
	}
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

	if err == nil && helpers.Config.Features.InjEnvs {
		for k, v := range envs {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
		helpers.Logf(helpers.Green + "✅ Injecting .env file" + helpers.Reset)
	}

	// generates code for .env and handlers
	generateCode()

	if !helpers.Config.Features.Watch {
		// exits without running the application
		os.Exit(0)
	}

	// runs "go run ."
	if err := cmd.Start(); err != nil {
		log.Printf(helpers.Red+"❌ Failed to start new process: %v", err)
	}

	return cmd
}

// logic for generating all the code before executing the command
func generateCode() {
	funcs := []func(){}
	if helpers.Config.Features.GenRouter {
		funcs = append(funcs, router.GenerateRouter)
	}
	if helpers.Config.Features.GenEnvs {
		funcs = append(funcs, envs.GenerateEnvs)
	}

	helpers.WaitMany(funcs...)
}
