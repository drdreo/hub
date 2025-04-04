package runner

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"os/exec"
	"path/filepath"
	"strings"

	"demo-cli/internal/config"
)

// ExecuteStep runs a single demo step based on its type
func ExecuteStep(step config.Step) error {
	switch step.Type {
	case "generate":
		return generateFile(step)
	case "modify":
		return modifyFile(step)
	case "execute":
		return executeCommand(step)
	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}
}

// generateFile creates a new file from a template
func generateFile(step config.Step) error {
	if step.Template == "" {
		return errors.New("template content is required for generate steps")
	}
	if step.Target == "" {
		return errors.New("target path is required for generate steps")
	}

	// Ensure the directory exists
	dir := filepath.Dir(step.Target)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write the file
	return os.WriteFile(step.Target, []byte(step.Template), 0644)
}

// modifyFile changes content in an existing file
func modifyFile(step config.Step) error {
	if step.Target == "" {
		return errors.New("target path is required for modify steps")
	}
	if step.Match == "" {
		return errors.New("match pattern is required for modify steps")
	}
	if step.Replace == "" {
		return errors.New("replacement text is required for modify steps")
	}

	// Read the file
	content, err := os.ReadFile(step.Target)
	if err != nil {
		return err
	}

	// Replace content
	newContent := strings.Replace(string(content), step.Match, step.Replace, 1)

	// Write back to the file
	return os.WriteFile(step.Target, []byte(newContent), 0644)
}

// executeCommand runs a shell command
func executeCommand(step config.Step) error {
	if step.Command == "" {
		return errors.New("command is required for execute steps")
	}

	// Split the command string
// 	parts := strings.Fields(step.Command)
// 	if len(parts) == 0 {
// 		return errors.New("invalid command")
// 	}

	// Use shell to properly handle command operators like &&, |, etc.
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", step.Command)
	} else {
		cmd = exec.Command("sh", "-c", step.Command)
	}


	// Get the command and arguments
// 	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
