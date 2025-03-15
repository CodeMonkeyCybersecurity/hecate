/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/CodeMonkeyCybersecurity/hecate/pkg/utils"
	"github.com/spf13/cobra"
)

// configCmd represents the subcommand that runs under "hecate create ...".
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Update configuration variables",
	Long: `Prompts the user for BACKEND_IP, PERS_BACKEND_IP, DELPHI_BACKEND_IP, and BASE_DOMAIN,
and recursively updates placeholders in all .conf files under the conf.d directory.
A backup is created for any file that is modified.

Usage: hecate create config
`,
	Run: func(cmd *cobra.Command, args []string) {
		runConfig()
	},
}

// init adds this command to createCmd so you can call it via `hecate create config`.
func init() {
	createCmd.AddCommand(configCmd)
}

// runConfig contains the logic formerly in createConfigVariables.go.
func runConfig() {
	fmt.Println("=== Recursive conf.d Variable Updater ===\n")

	// 1. Load previous values if available, using the utility function.
	lastValues, err := utils.LoadLastValues()
	if err != nil {
		fmt.Printf("Error loading previous values: %v\n", err)
		os.Exit(1)
	}

	// 2. Prompt user for variables (using the utility function).
	BACKEND_IP := utils.PromptInput("BACKEND_IP", "Enter the backend IP address", lastValues["BACKEND_IP"])
	PERS_BACKEND_IP := utils.PromptInput("PERS_BACKEND_IP", "Enter the backend IP address for your Persephone backups", lastValues["PERS_BACKEND_IP"])
	DELPHI_BACKEND_IP := utils.PromptInput("DELPHI_BACKEND_IP", "Enter the backend IP address for your Delphi install", lastValues["DELPHI_BACKEND_IP"])
	BASE_DOMAIN := utils.PromptInput("BASE_DOMAIN", "Enter the base domain for your services", lastValues["BASE_DOMAIN"])

	// 3. Save the values for future runs.
	newValues := map[string]string{
		"BACKEND_IP":        BACKEND_IP,
		"PERS_BACKEND_IP":   PERS_BACKEND_IP,
		"DELPHI_BACKEND_IP": DELPHI_BACKEND_IP,
		"BASE_DOMAIN":       BASE_DOMAIN,
	}
	if err := utils.SaveLastValues(newValues); err != nil {
		fmt.Printf("Error saving values: %v\n", err)
		os.Exit(1)
	}

	// 4. Check that the conf.d directory exists (the default in utils is also "conf.d", but
	//    you can override or keep your own local check).
	info, err := os.Stat(utils.ConfDir)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error: Directory '%s' not found in the current directory.\n", utils.ConfDir)
		os.Exit(1)
	}

	// 5. Process all .conf files in conf.d using the utility function.
	if err := utils.ProcessConfDirectory(utils.ConfDir, BACKEND_IP, PERS_BACKEND_IP, DELPHI_BACKEND_IP, BASE_DOMAIN); err != nil {
		fmt.Printf("Error processing configuration files: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nDone updating configuration files in the conf.d directory.")
}
