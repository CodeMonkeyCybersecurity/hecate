package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// httpCmd represents the http subcommand of "create"
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Create HTTP configuration",
	Long:  `Interactively generate an HTTP configuration file for your web server.`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile := "http.conf"
		backupFile := "http.conf.bak"

		fmt.Println("Welcome to the HTTP block configuration updater for your http.conf file.")
		fmt.Println("Below you'll see a description for each setting along with its current default value.")
		fmt.Println("Press Enter to keep the default value or type a new one as desired.\n")

		serverTokens := PromptWithDefault(
			"Hide NGINX version (server_tokens)",
			"off",
			"server_tokens: When set to 'off', NGINX will not display its version in error pages and headers.",
		)

		includeMime := PromptWithDefault(
			"Path to MIME types file (include)",
			"mime.types",
			"include: This file defines MIME types for various file extensions.",
		)

		defaultType := PromptWithDefault(
			"Default MIME type (default_type)",
			"application/octet-stream",
			"default_type: The default MIME type for files with unknown extensions.",
		)

		errorLogPath := PromptWithDefault(
			"Error log path",
			"/var/log/nginx/error.log",
			"error_log: The file path where NGINX will log error messages.",
		)

		errorLogLevel := PromptWithDefault(
			"Error log level (e.g., warn, debug)",
			"warn",
			"Error log level: Determines the minimum severity of messages to be logged (e.g., 'warn' or 'debug').",
		)

		accessLogPath := PromptWithDefault(
			"Access log path",
			"/var/log/nginx/access.log",
			"access_log: The file path where NGINX will log access details for incoming requests.",
		)

		sendfile := PromptWithDefault(
			"Sendfile (on/off)",
			"on",
			"sendfile: When 'on', NGINX uses the sendfile system call to transfer files efficiently.",
		)

		includeServers := PromptWithDefault(
			"Include server blocks directory/pattern",
			"/etc/nginx/conf.d/servers/*.conf",
			"include: Specifies the path or pattern for including server block configuration files.",
		)

		// Build the configuration content.
		configContent := fmt.Sprintf(`###
# HTTP BLOCK
###

http {
    # Hide NGINX version
    server_tokens   %s;
    include         %s;
    default_type    %s;

    error_log       %s %s;    # change warn to debug if installing a development server
    access_log      %s;       # enable access logging 
    sendfile        %s;
    
    ###
    # SERVER BLOCKS
    ###
    include          %s;
}
`, serverTokens, includeMime, defaultType, errorLogPath, errorLogLevel, accessLogPath, sendfile, includeServers)

		// Check if http.conf exists and back it up if necessary.
		if _, err := os.Stat(configFile); err == nil {
			// Remove existing backup if present.
			if _, err := os.Stat(backupFile); err == nil {
				if err := os.Remove(backupFile); err != nil {
					fmt.Printf("Error removing existing backup %s: %v\n", backupFile, err)
					os.Exit(1)
				}
			}
			// Rename current config file to backup.
			if err := os.Rename(configFile, backupFile); err != nil {
				fmt.Printf("Error backing up %s to %s: %v\n", configFile, backupFile, err)
				os.Exit(1)
			}
			fmt.Printf("\nExisting %s has been backed up to %s\n", configFile, backupFile)
		}

		// Write the new configuration to http.conf.
		f, err := os.Create(configFile)
		if err != nil {
			fmt.Printf("Error creating %s: %v\n", configFile, err)
			os.Exit(1)
		}
		defer f.Close()

		_, err = f.WriteString(configContent)
		if err != nil {
			fmt.Printf("Error writing to %s: %v\n", configFile, err)
			os.Exit(1)
		}

		fmt.Printf("\nNew configuration has been written to %s\n", configFile)
	},
}

func init() {
	// Attach this command to the existing "create" command.
	// Ensure that createCmd is defined in your project.
	createCmd.AddCommand(httpCmd)
}
