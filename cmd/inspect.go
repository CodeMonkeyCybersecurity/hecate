// cmd/inspect.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// inspectCmd represents the top-level inspect command.
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect (read) various resources",
	Long:  `The inspect command allows you to view current configurations and resources without modifying them.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Please use a subcommand (e.g. 'inspect config') to inspect a resource.")
	},
}

// inspectConfigCmd represents the "inspect config" subcommand.
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

func init() {
	// Attach inspectCmd to the root command.
	rootCmd.AddCommand(inspectCmd)
	// Attach inspectConfigCmd as a subcommand of inspectCmd.
	inspectCmd.AddCommand(inspectConfigCmd)
}

// runInspectConfig presents an interactive menu for inspection.
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

// inspectCertificates displays a list of certificates (for example, the files in a local "certs" directory).
func inspectCertificates() {
	certsDir := "certs" // Adjust if your certificates are stored elsewhere.
	fmt.Printf("\n--- Inspecting Certificates in '%s' ---\n", certsDir)
	files, err := os.ReadDir(certsDir)
	if err != nil {
		fmt.Printf("Error reading certificates directory: %v\n", err)
		return
	}
	if len(files) == 0 {
		fmt.Println("No certificates found.")
		return
	}
	for _, file := range files {
		fmt.Printf(" - %s\n", file.Name())
	}
}

// inspectDockerCompose reads and prints the contents of the docker-compose file.
func inspectDockerCompose() {
	configFile := "docker-compose.yml"
	fmt.Printf("\n--- Inspecting docker-compose file: %s ---\n", configFile)
	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", configFile, err)
		return
	}
	fmt.Println(string(data))
}

// inspectEosConfig lists the Eos backend configuration files (from the conf.d directory).
func inspectEosConfig() {
	confDir := "conf.d"
	fmt.Printf("\n--- Inspecting Eos backend web apps configuration in '%s' ---\n", confDir)
	files, err := os.ReadDir(confDir)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", confDir, err)
		return
	}
	found := false
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".conf") {
			fmt.Printf(" - %s\n", file.Name())
			found = true
		}
	}
	if !found {
		fmt.Println("No Eos configuration files found.")
	}
}

// inspectNginxDefaults reads and prints the contents of the http.conf file.
func inspectNginxDefaults() {
	configFile := "http.conf"
	fmt.Printf("\n--- Inspecting Nginx defaults in %s ---\n", configFile)
	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", configFile, err)
		return
	}
	fmt.Println(string(data))
}
