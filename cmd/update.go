/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
// update.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/CodeMonkeyCybersecurity/hecate/pkg/config"
	"github.com/CodeMonkeyCybersecurity/hecate/pkg/utils"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command.
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update various resources",
	Long: `This command updates various configurations for Hecate:
  
  1) Update Certificates
  2) Update docker-compose file
  3) Update Eos backend web apps configuration
  4) Update Nginx defaults
  5) Update all configurations

You can choose to update one or all of these resources interactively.`,
	Run: func(cmd *cobra.Command, args []string) {
		runUpdateConfig()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

// runUpdateConfig presents an interactive menu for update actions.
func runUpdateConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== Update Configurations ===")
	fmt.Println("Select the resource you want to update:")
	fmt.Println("1) Update Certificates")
	fmt.Println("2) Update docker-compose file")
	fmt.Println("3) Update Eos backend web apps configuration")
	fmt.Println("4) Update Nginx defaults")
	fmt.Println("5) Update all configurations")
	fmt.Print("Enter choice (1-5): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		updateCertificates()
	case "2":
		updateDockerCompose()
	case "3":
		updateEosConfig()
	case "4":
		updateNginxDefaults()
	case "5":
		updateCertificates()
		updateDockerCompose()
		updateEosConfig()
		updateNginxDefaults()
	default:
		fmt.Println("Invalid choice. Exiting.")
		os.Exit(1)
	}
}

func updateCertificates() {
	fmt.Println("\n--- Updating Certificates ---")
	// For example, re-run certificate generation logic or allow updating certificate parameters.
	// Here, we simply call runCerts() as a placeholder (or you could add more refined update logic).
	runCerts()
	fmt.Println("Certificates updated.")
}

func updateDockerCompose() {
	fmt.Println("\n--- Updating docker-compose file ---")
	// As an update operation, you might want to re-run the compose update logic.
	// For now, we'll call RunComposeInteractive() from utils as a placeholder.
	utils.RunComposeInteractive()
	fmt.Println("docker-compose file updated.")
}

func updateEosConfig() {
	fmt.Println("\n--- Updating Eos backend web apps configuration ---")
	// You might re-run your Eos configuration logic (or update specific settings).
	// For now, we'll call runEos() as a placeholder.
	runEos()
	fmt.Println("Eos backend web apps configuration updated.")
}

func updateNginxDefaults() {
	fmt.Println("\n--- Updating Nginx defaults ---")
	// Similarly, you can re-run your Nginx configuration updater (or update specific settings).
	// Here we call runHttp() as a placeholder.
	runHttp()
	fmt.Println("Nginx defaults updated.")
}
