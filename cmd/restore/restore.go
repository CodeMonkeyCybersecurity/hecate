package restore

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hecate/pkg/utils"

	"github.com/spf13/cobra"
)

var timestampFlag string

// RestoreCmd represents the restore command.
var RestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore configuration and files from backup",
	Long: `Restore configuration files, certificates, and docker-compose file from backups.

If --timestamp is provided (e.g. --timestamp 20250325-101010), then restore will look for:
  conf.d.<timestamp>.bak
  certs.<timestamp>.bak
  docker-compose.yml.<timestamp>.bak

If no --timestamp is given, the command enters interactive mode to choose which resources to restore.`,
	Run: func(cmd *cobra.Command, args []string) {
		if timestampFlag != "" {
			runAutoRestore(timestampFlag)
		} else {
			runInteractiveRestore()
		}
	},
}

func init() {
	root.RootCmd.AddCommand(RestoreCmd) // ✅ Attach to RootCmd, NOT createCmd

	// Define timestamp flag
	RestoreCmd.Flags().StringVarP(&timestampFlag, "timestamp", "t", "",
		"Timestamp for backup (format: YYYYMMDD-HHMMSS). If omitted, interactive mode is used.")
}

// runAutoRestore automatically restores resources using the provided timestamp.
func runAutoRestore(ts string) {
	const (
		SRC_CONF    = "conf.d"
		SRC_CERTS   = "certs"
		SRC_COMPOSE = "docker-compose.yml"
	)

	backupConf := fmt.Sprintf("%s.%s.bak", SRC_CONF, ts)
	backupCerts := fmt.Sprintf("%s.%s.bak", SRC_CERTS, ts)
	backupCompose := fmt.Sprintf("%s.%s.bak", SRC_COMPOSE, ts)

	fmt.Printf("Restoring backups with timestamp %s...\n", ts)
	utils.RestoreDir(backupConf, SRC_CONF)
	utils.RestoreDir(backupCerts, SRC_CERTS)
	utils.RestoreFile(backupCompose, SRC_COMPOSE)
}

// runInteractiveRestore presents a menu to choose which resource(s) to restore.
func runInteractiveRestore() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== Interactive Restore ===")
	fmt.Println("Select the resource you want to restore:")
	fmt.Println("1) Restore configuration (conf.d)")
	fmt.Println("2) Restore certificates (certs)")
	fmt.Println("3) Restore docker-compose file")
	fmt.Println("4) Restore all resources")
	fmt.Print("Enter choice (1-4): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		restoreConf()
	case "2":
		restoreCerts()
	case "3":
		restoreCompose()
	case "4":
		restoreConf()
		restoreCerts()
		restoreCompose()
	default:
		fmt.Println("Invalid choice. Exiting.")
		os.Exit(1)
	}
}

func restoreConf() {
	const SRC_CONF = "conf.d"
	backupConf, err := utils.FindLatestBackup(fmt.Sprintf("%s.", SRC_CONF))
	if err != nil {
		fmt.Printf("Error finding backup for %s: %v\n", SRC_CONF, err)
		return
	}
	fmt.Printf("Restoring configuration from backup: %s\n", backupConf)
	utils.RestoreDir(backupConf, SRC_CONF)
}

func restoreCerts() {
	const SRC_CERTS = "certs"
	backupCerts, err := utils.FindLatestBackup(fmt.Sprintf("%s.", SRC_CERTS))
	if err != nil {
		fmt.Printf("Error finding backup for %s: %v\n", SRC_CERTS, err)
		return
	}
	fmt.Printf("Restoring certificates from backup: %s\n", backupCerts)
	utils.RestoreDir(backupCerts, SRC_CERTS)
}

func restoreCompose() {
	const SRC_COMPOSE = "docker-compose.yml"
	backupCompose, err := utils.FindLatestBackup(fmt.Sprintf("%s.", SRC_COMPOSE))
	if err != nil {
		fmt.Printf("Error finding backup for %s: %v\n", SRC_COMPOSE, err)
		return
	}
	fmt.Printf("Restoring docker-compose file from backup: %s\n", backupCompose)
	utils.RestoreFile(backupCompose, SRC_COMPOSE)
}
