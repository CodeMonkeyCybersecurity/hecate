/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/CodeMonkeyCybersecurity/hecate/pkg/utils"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore configuration and files from backup",
	Long:  `Restore the configuration directory, certificates, and docker-compose file from their backups.`,
	Run: func(cmd *cobra.Command, args []string) {
		runRestore()
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

func runRestore() {
	const (
		BACKUP_CONF    = "conf.d.bak"
		BACKUP_CERTS   = "certs.bak"
		BACKUP_COMPOSE = "docker-compose.yml.bak"

		DST_CONF    = "conf.d"
		DST_CERTS   = "certs"
		DST_COMPOSE = "docker-compose.yml"
	)

	// Restore conf.d directory.
	info, err := os.Stat(BACKUP_CONF)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error: Backup directory '%s' does not exist.\n", BACKUP_CONF)
		os.Exit(1)
	}
	if err := utils.RemoveIfExists(DST_CONF); err != nil {
		fmt.Printf("Error removing %s: %v\n", DST_CONF, err)
		os.Exit(1)
	}
	if err := utils.CopyDir(BACKUP_CONF, DST_CONF); err != nil {
		fmt.Printf("Error during restore of %s: %v\n", BACKUP_CONF, err)
		os.Exit(1)
	}
	fmt.Printf("Restore complete: '%s' has been restored to '%s'.\n", BACKUP_CONF, DST_CONF)

	// Restore certs directory.
	info, err = os.Stat(BACKUP_CERTS)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error: Backup directory '%s' does not exist.\n", BACKUP_CERTS)
		os.Exit(1)
	}
	if err := utils.RemoveIfExists(DST_CERTS); err != nil {
		fmt.Printf("Error removing %s: %v\n", DST_CERTS, err)
		os.Exit(1)
	}
	if err := utils.CopyDir(BACKUP_CERTS, DST_CERTS); err != nil {
		fmt.Printf("Error during restore of %s: %v\n", BACKUP_CERTS, err)
		os.Exit(1)
	}
	fmt.Printf("Restore complete: '%s' has been restored to '%s'.\n", BACKUP_CERTS, DST_CERTS)

	// Restore docker-compose.yml file.
	info, err = os.Stat(BACKUP_COMPOSE)
	if err != nil || info.IsDir() {
		fmt.Printf("Error: Backup file '%s' does not exist.\n", BACKUP_COMPOSE)
		os.Exit(1)
	}
	if err := utils.RemoveIfExists(DST_COMPOSE); err != nil {
		fmt.Printf("Error removing %s: %v\n", DST_COMPOSE, err)
		os.Exit(1)
	}
	if err := utils.CopyFile(BACKUP_COMPOSE, DST_COMPOSE); err != nil {
		fmt.Printf("Error during restore of %s: %v\n", BACKUP_COMPOSE, err)
		os.Exit(1)
	}
	fmt.Printf("Restore complete: '%s' has been restored to '%s'.\n", BACKUP_COMPOSE, DST_COMPOSE)
}
