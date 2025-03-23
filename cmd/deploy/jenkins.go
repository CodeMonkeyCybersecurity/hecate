// cmd/deploy/jenkins.go

package deploy

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var deployJenkinsCmd = &cobra.Command{
	Use:   "jenkins",
	Short: "Deploy reverse proxy for Jenkins",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🚀 Deploying reverse proxy for Jenkins...")

		// 1. Ensure configuration values exist (from .hecate.conf) or ask the user.
		baseDomain, backendIP, err := ensureHecateConfig()
		if err != nil {
			return fmt.Errorf("failed to retrieve configuration: %w", err)
		}

		// 2. Replace placeholders in Jenkins config files (HTTP and Stream).
		if err := replacePlaceholders(filepath.Join("assets", "servers", "jenkins.conf"), baseDomain, backendIP); err != nil {
			return fmt.Errorf("failed to update servers config: %w", err)
		}
		if err := replacePlaceholders(filepath.Join("assets", "stream", "jenkins.conf"), baseDomain, backendIP); err != nil {
			return fmt.Errorf("failed to update stream config: %w", err)
		}
		fmt.Println("✅ Configurations updated.")

		// 3. Run docker compose up -d to deploy the reverse proxy (and Jenkins)
		cmdStr := "docker compose up -d"
		fmt.Printf("Running: %s\n", cmdStr)
		parts := strings.Split(cmdStr, " ")
		cmdExec := exec.Command(parts[0], parts[1:]...)
		cmdExec.Stdout = os.Stdout
		cmdExec.Stderr = os.Stderr

		if err := cmdExec.Run(); err != nil {
			return fmt.Errorf("failed to run docker compose: %w", err)
		}

		fmt.Println("🎉 Jenkins reverse proxy deployed successfully.")
		return nil
	},
}

// ensureHecateConfig checks for .hecate.conf, reads the BASE_DOMAIN and backendIP values if present,
// and asks the user to confirm or update them. It then writes the final values back to the file.
func ensureHecateConfig() (string, string, error) {
	const configPath = ".hecate.conf"
	var baseDomain, backendIP string

	// Check if the file exists.
	if _, err := os.Stat(configPath); err == nil {
		// File exists. Read its contents.
		file, err := os.Open(configPath)
		if err != nil {
			return "", "", fmt.Errorf("unable to open %s: %w", configPath, err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "BASE_DOMAIN=") {
				baseDomain = strings.TrimSpace(strings.TrimPrefix(line, "BASE_DOMAIN="))
			} else if strings.HasPrefix(line, "backendIP=") {
				backendIP = strings.TrimSpace(strings.TrimPrefix(line, "backendIP="))
			}
		}
		if err := scanner.Err(); err != nil {
			return "", "", fmt.Errorf("error reading %s: %w", configPath, err)
		}

		fmt.Printf("Found existing configuration:\n  BASE_DOMAIN: %s\n  backendIP: %s\n", baseDomain, backendIP)
		// Ask user if they want to keep the values or update them.
		if yesOrNo("Do you want to keep these values? (Y/n): ") {
			// User confirmed; nothing to change.
		} else {
			// Ask for new values.
			baseDomain = prompt("Enter new BASE_DOMAIN: ")
			backendIP = prompt("Enter new backendIP: ")
		}
	} else {
		// File does not exist; ask for new values.
		baseDomain = prompt("Enter BASE_DOMAIN: ")
		backendIP = prompt("Enter backendIP: ")
	}

	// Write the values back to .hecate.conf.
	content := fmt.Sprintf("BASE_DOMAIN=%s\nbackendIP=%s\n", baseDomain, backendIP)
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write %s: %w", configPath, err)
	}
	return baseDomain, backendIP, nil
}

// prompt reads a single line from standard input after displaying the provided message.
func prompt(message string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

// yesOrNo asks the user a yes/no question and returns true if the answer is yes (default yes).
func yesOrNo(message string) bool {
	response := prompt(message)
	// Default to yes if response is empty.
	if response == "" {
		return true
	}
	response = strings.ToLower(response)
	return response == "y" || response == "yes"
}

// replacePlaceholders opens the file at filePath, replaces ${BASE_DOMAIN} and ${backendIP} placeholders
// with the provided values, and writes the updated content back to the same file.
func replacePlaceholders(filePath, baseDomain, backendIP string) error {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	content := string(contentBytes)
	content = strings.ReplaceAll(content, "${BASE_DOMAIN}", baseDomain)
	content = strings.ReplaceAll(content, "${backendIP}", backendIP)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing file %s: %w", filePath, err)
	}
	return nil
}

// NewDeployJenkinsCmd exposes this command to be added to the root command.
func NewDeployJenkinsCmd() *cobra.Command {
	return deployJenkinsCmd
}
