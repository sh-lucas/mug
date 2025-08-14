package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/joho/godotenv"
	gen_envs "github.com/sh-lucas/mug/generator/envs"
	gen_router "github.com/sh-lucas/mug/generator/router"
	"github.com/sh-lucas/mug/global"
	"github.com/sh-lucas/mug/watcher"
)

func main() {
	if global.Config.Features.AutoTidy {
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

	if err == nil && global.Config.Features.InjEnvs {
		for k, v := range envs {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
		global.Logf(global.Green + "✅ Injecting .env file" + global.Reset)
	}

	// generates code for .env and handlers
	generateCode()

	if !global.Config.Features.Watch {
		// exits without running the application
		os.Exit(0)
	}

	// runs "go run ."
	if err := cmd.Start(); err != nil {
		log.Printf(global.Red+"❌ Failed to start new process: %v", err)
	}

	return cmd
}

// logic for generating all the code before executing the command
func generateCode() {
	funcs := []func(){}
	if global.Config.Features.GenRouter {
		funcs = append(funcs, gen_router.GenerateRouter)
	}
	if global.Config.Features.GenEnvs {
		funcs = append(funcs, gen_envs.GenerateEnvs)
	}

	global.WaitMany(funcs...)
}
