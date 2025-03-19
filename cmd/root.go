/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"hecate/cmd/create"
	"hecate/cmd/delete"
	"hecate/cmd/inspect"
	"hecate/cmd/update"
	"hecate/cmd/deploy"
	
	"os"

	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var log *zap.Logger // Global logger instance


// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hecate",
	Short: Hecate CLI: Manage and configure your reverse proxy",
	Long: `Hecate is a command-line tool designed to simplify the management and configuration 
of your reverse proxy setup. It provides a unified interface for tasks such as deploying, 
updating, and monitoring your proxy environment.

Examples:
	# Display the help information for Hecate
  	hecate --help

  	# Deploy a new reverse proxy configuration for a specific application
  	hecate deploy jenkins
   	hecate deploy nextcloud
    	hecate deploy delphi

 	# Check the status of your reverse proxy services
  	hecate status

Use Hecate to quickly generate configuration files, manage certificates, and streamline 
the deployment process for your reverse proxy setup.`,

	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Eos CLI started successfully.")

		if !utils.CheckSudo() {
			log.Error("Sudo privileges are required to create a backup.")
			return
		}

		// Example: Process the config path
		configPath := filepath.Join(".", "config", "default.yaml")
		log.Info("Loaded configuration", zap.String("path", configPath))
	},
}


// Register all subcommands in the init function
func init() {
	rootCmd.AddCommand(create.CreateCmd)
	rootCmd.AddCommand(inspect.InspectCmd)
	rootCmd.AddCommand(update.UpdateCmd)
	rootCmd.AddCommand(delete.DeleteCmd)
}

// Execute starts the CLI
func Execute() {
	// Initialize the logger once globally
	logger.Initialize()
	defer logger.Sync()

	// Assign the logger instance globally for reuse
	log = logger.GetLogger()

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Error("CLI execution error", zap.Error(err))
		os.Exit(1)
	}
}
