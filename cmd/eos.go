/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
// cmd/eos.go
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// runEos performs the main logic for selecting EOS backend web apps and cleaning up config files.
func runEos() {
	fmt.Println("=== EOS Backend Web Apps Selector ===\n")
	lastValues, err := loadLastValues()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	// Use previously saved selection if available.
	defaultApps := lastValues["APPS_SELECTION"]
	displayOptions()
	allowedFiles, selectionStr := getUserSelection(defaultApps)
	// Always preserve essential files.
	essential := []string{"http.conf", "stream.conf", "fallback.conf"}
	for _, fname := range essential {
		allowedFiles[fname] = true
	}
	fmt.Println("\nYou have selected the following configuration files to keep:")
	for file := range allowedFiles {
		if file == "http.conf" || file == "stream.conf" || file == "fallback.conf" {
			fmt.Printf(" - Essential file: %s\n", file)
		} else {
			// Look up the app name.
			for _, opt := range APPS_SELECTION {
				if opt.ConfFile == file {
					fmt.Printf(" - %s (%s)\n", opt.AppName, file)
					break
				}
			}
		}
	}
	fmt.Println("\nNow scanning the conf.d directory and removing files not in your selection...")
	removeUnwantedConfFiles(allowedFiles)
	// Save the new selection.
	lastValues["APPS_SELECTION"] = selectionStr
	if err := saveLastValues(lastValues); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
	}
}

// eosAppsCmd represents the "eosApps" subcommand under "hecate create".
var eosAppsCmd = &cobra.Command{
	Use:   "eosApps",
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
  hecate create eosApps       // Runs in interactive mode using previous selections if available.
`,
	Run: func(cmd *cobra.Command, args []string) {
		runEos()
	},
}

func init() {
	// Assumes createCmd is defined in your application under the "hecate create" command.
	createCmd.AddCommand(eosAppsCmd)
}
