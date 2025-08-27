package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/sh-lucas/mug/internal/config"
	"github.com/sh-lucas/mug/internal/watcher"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mug",
	Short: "Get a mug of coffee and relax.",
	Long:  "Mug is a router generator, code rebuilder, .env loader, and backend framework. Simply use `mug` to run in default settings, or configure the settings using `mug init`",
	Run: func(cmd *cobra.Command, args []string) {
		watcher.Start(func() *exec.Cmd {
			cmd := getBuildCmd()

			generateCode()
			if config.Global.Watch.Tidy {
				_ = exec.Command("go", "mod", "tidy").Run()
			}

			injectEnvs(cmd)

			if err := cmd.Start(); err != nil {
				fmt.Printf("❌ Could not start server: %s\n", err)
			}
			return cmd
		},
		)
	},
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watches for file changes and rebuilds",
	Long:  "Start the application and automatically rebuild with file changes. add `-gen` to also generate code.",
	Run: func(cmd *cobra.Command, args []string) {
		watcher.Start(func() *exec.Cmd {
			cmd := getBuildCmd()

			injectEnvs(cmd)

			if err := cmd.Start(); err != nil {
				fmt.Printf("❌ Could not start server: %s\n", err)
			}
			return cmd
		})
	},
}

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Only generates code.",
	Long:  "Generates the specified router/env packages for rapid easier development.",
	Run: func(cmd *cobra.Command, args []string) {
		generateCode()
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Dump default settings; empty variables will get defaults too.",
	Run: func(cmd *cobra.Command, args []string) {
		config.DumpConfig()
	},
}

// build not yet implemented
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Compila o projeto (não implementado)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚧 Command 'build' not yet implemented.")
	},
}

// make not yet implemented
var makeCmd = &cobra.Command{
	Use:   "make",
	Short: "Executa tarefas de automação (não implementado)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚧 Command 'make' not yet implemented.")
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(genCmd)
	rootCmd.AddCommand(initCmd)
	// rootCmd.AddCommand(buildCmd)
	// rootCmd.AddCommand(makeCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Ocorreu um erro: '%s'", err)
		os.Exit(1)
	}
}
