// cmd/inspect/inspect.go

package inspect

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// InspectCmd is the top-level `inspect` command
var InspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect the current state of Hecate-managed services",
	Long: `Use this command to inspect the status, configuration, and health of 
reverse proxy applications deployed via Hecate.

Examples:
  hecate inspect config
  hecate inspect`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Please use a subcommand (e.g. 'inspect config') to inspect a resource.")
	},
}

// inspectConfigCmd represents the "inspect config" subcommand
var inspectConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Inspect configurations",
	Long: `This command lets you inspect various configuration resources for Hecate.
You can choose from:
  1) Inspect Certificates
  2) Inspect docker-compose file
  3) Inspect Eos backend web apps configuration
  4) Inspect Nginx defaults
  5) Inspect all configurations`,
	Run: func(cmd *cobra.Command, args []string) {
		runInspectConfig()
	},
}

// Register subcommands when the package is loaded
func init() {
	InspectCmd.AddCommand(inspectConfigCmd)
}

// runInspectConfig presents an interactive menu for inspection
func runInspectConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== Inspect Configurations ===")
	fmt.Println("Select the resource you want to inspect:")
	fmt.Println("1) Inspect Certificates")
	fmt.Println("2) Inspect docker-compose file")
	fmt.Println("3) Inspect Eos backend web apps configuration")
	fmt.Println("4) Inspect Nginx defaults")
	fmt.Println("5) Inspect all configurations")
	fmt.Print("Enter choice (1-5): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		inspectCertificates()
	case "2":
		inspectDockerCompose()
	case "3":
		inspectEosConfig()
	case "4":
		inspectNginxDefaults()
	case "5":
		inspectCertificates()
		inspectDockerCompose()
		inspectEosConfig()
		inspectNginxDefaults()
	default:
		fmt.Println("Invalid choice. Exiting.")
		os.Exit(1)
	}
}
