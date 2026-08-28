package main

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sh-lucas/mug/internal/config"
	"github.com/sh-lucas/mug/internal/generator/envs"
	"github.com/sh-lucas/mug/internal/generator/router"
	"github.com/sh-lucas/mug/internal/helpers"
	"github.com/sh-lucas/mug/pkg"
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

	if config.Global.Watch.InjectEnvs != "" {
		envs, err := godotenv.Read(config.Global.Watch.InjectEnvs)
		if err == nil {
			for k, v := range envs {
				cmd.Env = append(cmd.Env, k+"="+v)
			}
			helpers.Logf(pkg.Green + "✅ Injecting .env file" + pkg.Reset)
		}
	}
}

// logic for generating all the code before executing the command
func generateCode() {
	funcs := []func(){}
	if config.Global.Gen.Router {
		funcs = append(funcs, router.GenerateRouter)
	}
	if config.Global.Gen.Envs {
		funcs = append(funcs, envs.GenerateEnvs)
	}

	helpers.WaitMany(funcs...)
}
