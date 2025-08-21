package main

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sh-lucas/mug/internal/generator/envs"
	"github.com/sh-lucas/mug/internal/generator/router"
	"github.com/sh-lucas/mug/internal/helpers"
)

// returns the basic build cmd to run the application.
// It is recommended to use set and generate code before running with cmd.Start().
func getBuildCmd() *exec.Cmd {
	cmd := exec.Command("nice", "-n", "15", "go", "run", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// groups the processes
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}

// if the config allows it, injects the environment variables from the envs map.
func injectEnvs(cmd *exec.Cmd) {
	cmd.Env = os.Environ()

	envs, err := godotenv.Read()
	if err == nil && helpers.Config.Features.InjEnvs {
		for k, v := range envs {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
		helpers.Logf(helpers.Green + "✅ Injecting .env file" + helpers.Reset)
	}
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
