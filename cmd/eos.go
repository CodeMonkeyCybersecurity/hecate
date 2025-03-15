/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
// cmd/eos.go
package cmd

import (
	"fmt"
	"os"

	"github.com/CodeMonkeyCybersecurity/hecate/pkg/config"
	"github.com/CodeMonkeyCybersecurity/hecate/pkg/utils"
	"github.com/spf13/cobra"
)

func runEos() {
	fmt.Println("=== EOS Backend Web Apps Selector ===\n")
	// Load the previous configuration from .hecate.conf.
	lastValues, err := utils.LoadLastValues()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	// Use previously saved selection if available.
	defaultApps := lastValues["APPS_SELECTION"]

	// Display available app options using the config package.
	config.DisplayOptions()
	// Get user selection from the config package.
	selectedApps, selectionStr := config.GetUserSelection(defaultApps)

	// Build a map of allowed configuration filenames.
	allowedFiles := make(map[string]bool)
	// Always preserve essential files.
	essential := []string{"http.conf", "stream.conf", "fallback.conf"}
	for _, fname := range essential {
		allowedFiles[fname] = true
	}
	// For each selected app, mark its configuration file as allowed.
	for _, app := range selectedApps {
		allowedFiles[app.ConfFile] = true
	}

	fmt.Println("\nYou have selected the following configuration files to keep:")
	for file := range allowedFiles {
		if file == "http.conf" || file == "stream.conf" || file == "fallback.conf" {
			fmt.Printf(" - Essential file: %s\n", file)
		} else {
			// Look up the app name from config.Apps.
			for _, opt := range config.Apps {
				if opt.ConfFile == file {
					fmt.Printf(" - %s (%s)\n", opt.Name, file)
					break
				}
			}
		}
	}

	fmt.Println("\nNow scanning the conf.d directory and removing files not in your selection...")
	utils.RemoveUnwantedConfFiles(allowedFiles)

	// Save the new selection back to the configuration.
	lastValues["APPS_SELECTION"] = selectionStr
	if err := utils.SaveLastValues(lastValues); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
	}
}

var eosAppsCmd = &cobra.Command{
	Use:   "eos",
	Short: "Select and clean up EOS backend web apps configuration",
	Long: `Interactively choose which EOS backend web apps should remain active and 
remove unwanted configuration files from the conf.d directory.

This command performs the following:
  - Loads any previous selection from .hecate.conf.
  - Displays a list of supported web apps with their corresponding configuration files.
  - Prompts for a new selection (or uses the previous selection by default).
  - Ensures essential files (http.conf, stream.conf, fallback.conf) are always preserved.
  - Scans the conf.d directory and removes .conf files not in the selected set.

Examples:
  hecate create eos       // Runs in interactive mode using previous selections if available.
`,
	Run: func(cmd *cobra.Command, args []string) {
		runEos()
	},
}

func init() {
	// Assumes createCmd is defined in your application under the "hecate create" command.
	createCmd.AddCommand(eosCmd)
}
