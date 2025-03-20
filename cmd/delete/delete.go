/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
// delete.go
package delete

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"path/filepath"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command.
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete various resources",
	Long: `This command deletes various resources for Hecate:

  1) Delete Certificates
  2) Delete docker-compose modifications/backups
  3) Delete Eos backend web apps configuration files
  4) Delete (or revert) Nginx defaults
  5) Delete all specified resources

You can choose to delete one or all of these resources interactively.`,
	Run: func(cmd *cobra.Command, args []string) {
		runDeleteConfig()
	},
}

// runDeleteConfig presents an interactive menu for delete actions.
func runDeleteConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== Delete Resources ===")
	fmt.Println("Select the resource you want to delete:")
	fmt.Println("1) Delete Certificates")
	fmt.Println("2) Delete docker-compose modifications/backups")
	fmt.Println("3) Delete Eos backend web apps configuration files")
	fmt.Println("4) Delete (or revert) Nginx defaults")
	fmt.Println("5) Delete all specified resources")
	fmt.Print("Enter choice (1-5): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		deleteCertificates()
	case "2":
		deleteDockerCompose()
	case "3":
		deleteEosConfig()
	case "4":
		deleteNginxDefaults()
	case "5":
		deleteCertificates()
		deleteDockerCompose()
		deleteEosConfig()
		deleteNginxDefaults()
	default:
		fmt.Println("Invalid choice. Exiting.")
		os.Exit(1)
	}
}

func deleteCertificates() {
	fmt.Println("\n--- Deleting Certificates ---")
	// Example: Remove files from the certificates directory.
	certsDir := "certs" // Adjust as necessary.
	err := os.RemoveAll(certsDir)
	if err != nil {
		fmt.Printf("Error deleting certificates directory: %v\n", err)
	} else {
		fmt.Printf("Certificates deleted (directory '%s' removed).\n", certsDir)
	}
}

func deleteDockerCompose() {
	fmt.Println("\n--- Deleting docker-compose modifications/backups ---")
	// Example: Remove backup files (you might have a naming pattern for backups).
	// Here we assume backups have a .bak extension in the current directory.
	matches, err := filepath.Glob("*_docker-compose.yml.bak")
	if err != nil {
		fmt.Printf("Error searching for backups: %v\n", err)
		return
	}
	for _, file := range matches {
		if err := os.Remove(file); err != nil {
			fmt.Printf("Error removing backup file %s: %v\n", file, err)
		} else {
			fmt.Printf("Removed backup file: %s\n", file)
		}
	}
}

func deleteEosConfig() {
	fmt.Println("\n--- Deleting Eos backend web apps configuration files ---")
	// Example: Delete all .conf files in the conf.d directory that match a pattern.
	confDir := "conf.d" // Adjust as necessary.
	err := os.RemoveAll(confDir)
	if err != nil {
		fmt.Printf("Error deleting Eos configuration directory %s: %v\n", confDir, err)
	} else {
		fmt.Printf("Eos backend configuration files deleted (directory '%s' removed).\n", confDir)
	}
}

func deleteNginxDefaults() {
	fmt.Println("\n--- Deleting (or reverting) Nginx defaults ---")
	// Example: Remove or revert http.conf to a backup if one exists.
	configFile := "http.conf"
	backupFile := "http.conf.bak"
	// Check if a backup exists, and if so, restore it (i.e. delete the current file and rename the backup)
	if _, err := os.Stat(backupFile); err == nil {
		if err := os.Remove(configFile); err != nil {
			fmt.Printf("Error removing current %s: %v\n", configFile, err)
		} else if err := os.Rename(backupFile, configFile); err != nil {
			fmt.Printf("Error restoring backup %s to %s: %v\n", backupFile, configFile, err)
		} else {
			fmt.Printf("Nginx defaults reverted by restoring %s.\n", configFile)
		}
	} else {
		// If no backup exists, simply delete the file.
		if err := os.Remove(configFile); err != nil {
			fmt.Printf("Error removing %s: %v\n", configFile, err)
		} else {
			fmt.Printf("Deleted %s.\n", configFile)
		}
	}
}
