// cmd/inspect.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// inspectCmd represents the "inspect" command.
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect various resources",
	Long:  `Inspect commands allow you to view existing configurations and resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use a subcommand (e.g. 'inspect config') to inspect a specific resource.")
	},
}

// inspectConfigCmd represents the "inspect config" subcommand.
var inspectConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Display the current configuration",
	Long:  `This command displays the current configuration (e.g., the http.conf file) so you can review its settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		// For example, read and print the http.conf file:
		configFile := "http.conf"
		data, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", configFile, err)
			return
		}
		fmt.Printf("Contents of %s:\n\n%s\n", configFile, string(data))
	},
}

func init() {
	// Attach the inspect command to the root command.
	rootCmd.AddCommand(inspectCmd)
	// Attach the inspectConfigCmd as a subcommand of inspectCmd.
	inspectCmd.AddCommand(inspectConfigCmd)
}
