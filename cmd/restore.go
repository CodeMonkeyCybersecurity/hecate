/*
Copyright © 2025 NAME HERE
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/CodeMonkeyCybersecurity/hecate/pkg/utils"
	"github.com/spf13/cobra"
)

var timestampFlag string

// restoreCmd represents the restore command.
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore configuration and files from backup",
	Long: `Restore the configuration directory, certificates, and docker-compose file from their backups.

If --timestamp is provided (e.g. --timestamp 20250325-101010), then restore will look for:
  conf.d.20250325-101010.bak
  certs.20250325-101010.bak
  docker-compose.yml.20250325-101010.bak

If no --timestamp is given, it automatically selects the most recent (lexicographically greatest) backup
for each of these items.`,
	Run: func(cmd *cobra.Command, args []string) {
		runRestore()
	},
}

func init() {
	// Attach restoreCmd to createCmd so that you can run: hecate create restore
	createCmd.AddCommand(restoreCmd)

	// Define a timestamp flag for optional backup selection.
	restoreCmd.Flags().StringVarP(&timestampFlag, "timestamp", "t", "",
		"Timestamp used by backup (format: YYYYMMDD-HHMMSS). If omitted, the most recent backup is used.")
}

func runRestore() {
	const (
		SRC_CONF    = "conf.d"
		SRC_CERTS   = "certs"
		SRC_COMPOSE = "docker-compose.yml"
	)

	var backupConf, backupCerts, backupCompose string
	var err error

	if timestampFlag != "" {
		backupConf = fmt.Sprintf("%s.%s.bak", SRC_CONF, timestampFlag)
		backupCerts = fmt.Sprintf("%s.%s.bak", SRC_CERTS, timestampFlag)
		backupCompose = fmt.Sprintf("%s.%s.bak", SRC_COMPOSE, timestampFlag)
	} else {
		// Automatically detect the most recent backup.
		if backupConf, err = utils.FindLatestBackup(fmt.Sprintf("%s.", SRC_CONF)); err != nil {
			fmt.Printf("Error finding latest backup for %s: %v\n", SRC_CONF, err)
			os.Exit(1)
		}
		if backupCerts, err = utils.FindLatestBackup(fmt.Sprintf("%s.", SRC_CERTS)); err != nil {
			fmt.Printf("Error finding latest backup for %s: %v\n", SRC_CERTS, err)
			os.Exit(1)
		}
		if backupCompose, err = utils.FindLatestBackup(fmt.Sprintf("%s.", SRC_COMPOSE)); err != nil {
			fmt.Printf("Error finding latest backup for %s: %v\n", SRC_COMPOSE, err)
			os.Exit(1)
		}
	}

	// Restore the backups.
	utils.RestoreDir(backupConf, SRC_CONF)
	utils.RestoreDir(backupCerts, SRC_CERTS)
	utils.RestoreFile(backupCompose, SRC_COMPOSE)
}
