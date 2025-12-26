package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"syscall"

	"github.com/qawatake/kakucho/internal/config"
	"github.com/qawatake/kakucho/internal/help"
	"github.com/qawatake/kakucho/internal/shell"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "kakucho",
		Short: "Extend any CLI with custom subcommands",
		Long:  `kakucho allows you to add custom subcommands to any CLI tool.`,
	}

	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(execCmd())
	rootCmd.AddCommand(lsCmd())
	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(rmCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <shell>",
		Short: "Generate shell integration script",
		Long:  `Generate shell integration script for bash or zsh.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			script, err := shell.GenerateInit(args[0], cfg)
			if err != nil {
				return err
			}

			fmt.Print(script)
			return nil
		},
	}
}

func execCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "exec <target> [args...]",
		Short:              "Execute a command through kakucho",
		Long:               `Execute a command through kakucho, routing to extensions when appropriate.`,
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			target := args[0]
			cmdArgs := args[1:]

			return executeCommand(cfg, target, cmdArgs)
		},
	}
}

func executeCommand(cfg config.Config, target string, args []string) error {
	if len(args) == 0 {
		return executeOriginal(target, args)
	}

	subcommand := args[0]

	// Check if this is a help command
	if isHelpCommand(subcommand) {
		helper := help.New(cfg)
		return helper.ExecuteHelp(target, args)
	}

	// Check if subcommand is an extension
	ext := cfg.GetExtension(target, subcommand)
	if ext != nil {
		// Check if asking for help on extension
		if len(args) > 1 && isHelpCommand(args[1]) {
			helper := help.New(cfg)
			return helper.ExecuteExtensionHelp(target, subcommand)
		}
		return executeExtension(target, ext, args[1:])
	}

	// Fall back to original command
	return executeOriginal(target, args)
}

func isHelpCommand(arg string) bool {
	return arg == "help" || arg == "-h" || arg == "--help"
}

func executeOriginal(target string, args []string) error {
	path, err := exec.LookPath(target)
	if err != nil {
		return fmt.Errorf("command not found: %s", target)
	}

	cmdArgs := append([]string{target}, args...)
	return syscall.Exec(path, cmdArgs, os.Environ())
}

func executeExtension(target string, ext *config.Extension, args []string) error {
	path, err := config.GetExtensionPath(target, ext.Name)
	if err != nil {
		return fmt.Errorf("extension executable not found: %s-%s", target, ext.Name)
	}

	cmdArgs := append([]string{path}, args...)
	return syscall.Exec(path, cmdArgs, os.Environ())
}

func lsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls [target]",
		Short: "List registered extensions",
		Long:  `List all registered extensions, optionally filtered by target.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if len(args) == 1 {
				// List extensions for specific target
				target := args[0]
				extensions, ok := cfg[target]
				if !ok {
					fmt.Printf("No extensions registered for '%s'\n", target)
					return nil
				}
				printExtensions(target, extensions)
			} else {
				// List all targets and extensions
				targets := cfg.GetTargetNames()
				sort.Strings(targets)

				if len(targets) == 0 {
					fmt.Println("No extensions registered")
					return nil
				}

				for i, name := range targets {
					printExtensions(name, cfg[name])
					if i < len(targets)-1 {
						fmt.Println()
					}
				}
			}
			return nil
		},
	}
}

func printExtensions(name string, extensions map[string]string) {
	fmt.Printf("%s:\n", name)
	if len(extensions) == 0 {
		fmt.Println("  (no extensions)")
		return
	}

	names := make([]string, 0, len(extensions))
	for n := range extensions {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, n := range names {
		desc := extensions[n]
		execPath := name + "-" + n
		fmt.Printf("  %s\t%s\t(%s)\n", n, desc, execPath)
	}
}

func addCmd() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "add <target> <name>",
		Short: "Add a new extension",
		Long: `Add a new extension command to a target CLI.
The extension executable must be named {target}-{name} and be in your PATH.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			target := args[0]
			name := args[1]

			cfg.AddExtension(target, name, description)

			if err := cfg.Save(); err != nil {
				return err
			}

			fmt.Printf("Added extension '%s' to '%s' (executable: %s-%s)\n", name, target, target, name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Description of the extension")

	return cmd
}

func rmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <target> <name>",
		Short: "Remove an extension",
		Long:  `Remove an extension command from a target CLI.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			target := args[0]
			name := args[1]

			if !cfg.RemoveExtension(target, name) {
				return fmt.Errorf("extension '%s' not found for target '%s'", name, target)
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			fmt.Printf("Removed extension '%s' from '%s'\n", name, target)
			return nil
		},
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("kakucho version %s\n", version)
		},
	}
}
