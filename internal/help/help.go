package help

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/qawatake/kakucho/internal/config"
)

// Helper handles help output augmentation
type Helper struct {
	Config config.Config
}

// New creates a new Helper
func New(cfg config.Config) *Helper {
	return &Helper{Config: cfg}
}

// ExecuteHelp runs the help command and appends extension information
func (h *Helper) ExecuteHelp(target string, args []string) error {
	// Get original help output
	originalHelp, exitCode, err := h.getOriginalHelp(target, args)
	if err != nil {
		return err
	}

	// Print original help
	fmt.Print(originalHelp)

	// Only append extensions if original help succeeded
	if exitCode == 0 {
		h.printExtensions(target)
	}

	// Exit with the same code as original
	os.Exit(exitCode)
	return nil
}

// getOriginalHelp runs the original help command and captures output
func (h *Helper) getOriginalHelp(target string, args []string) (string, int, error) {
	path, err := exec.LookPath(target)
	if err != nil {
		return "", -1, err
	}

	cmd := exec.Command(path, args...)
	output, err := cmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return "", -1, err
		}
	}

	return string(output), exitCode, nil
}

// printExtensions prints the extension help section
func (h *Helper) printExtensions(target string) {
	extensions := h.Config.GetExtensions(target)
	if len(extensions) == 0 {
		return
	}

	// Sort extensions by name
	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Name < extensions[j].Name
	})

	// Find max name length for alignment
	maxLen := 0
	for _, ext := range extensions {
		if len(ext.Name) > maxLen {
			maxLen = len(ext.Name)
		}
	}

	fmt.Println()
	fmt.Println("EXTENSIONS (via kakucho):")
	for _, ext := range extensions {
		padding := strings.Repeat(" ", maxLen-len(ext.Name)+2)
		fmt.Printf("  %s%s%s\n", ext.Name, padding, ext.Description)
	}
}

// ExecuteExtensionHelp runs help for an extension command
func (h *Helper) ExecuteExtensionHelp(target, subcommand string) error {
	ext := h.Config.GetExtension(target, subcommand)
	if ext == nil {
		return fmt.Errorf("extension '%s' not found for target '%s'", subcommand, target)
	}

	// Get the extension path
	path, err := config.GetExtensionPath(target, subcommand)
	if err != nil {
		// Command not found, just show description
		fmt.Printf("%s: %s\n", subcommand, ext.Description)
		return nil
	}

	// Try to run the extension with -h
	cmd := exec.Command(path, "-h")
	output, err := cmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Command execution error, just show description
			fmt.Printf("%s: %s\n", subcommand, ext.Description)
			return nil
		}
	}

	if exitCode == 0 {
		fmt.Print(string(output))
	} else {
		// Extension -h returned error, just show description
		fmt.Printf("%s: %s\n", subcommand, ext.Description)
	}

	return nil
}
