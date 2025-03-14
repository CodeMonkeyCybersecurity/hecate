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


// backupCmd represents the backup command.
var backupCmd = &cobra.Command{
    Use:   "backup",
    Short: "Backup configuration and files",
    Long:  `Backup important configuration directories and files.`,
    Run: func(cmd *cobra.Command, args []string) {
        const (
            SRC_CONF    = "conf.d"
            SRC_CERTS   = "certs"
            SRC_COMPOSE = "docker-compose.yml"

            BACKUP_CONF    = "conf.d.bak"
            BACKUP_CERTS   = "certs.bak"
            BACKUP_COMPOSE = "docker-compose.yml.bak"
        )

        // Backup conf.d directory
        confInfo, err := os.Stat(SRC_CONF)
        if err != nil || !confInfo.IsDir() {
            fmt.Printf("Error: Source directory '%s' does not exist.\n", SRC_CONF)
            os.Exit(1)
        }
        if err := utils.RemoveIfExists(BACKUP_CONF); err != nil {
            fmt.Printf("Error removing backup directory '%s': %v\n", BACKUP_CONF, err)
            os.Exit(1)
        }
        if err := utils.CopyDir(SRC_CONF, BACKUP_CONF); err != nil {
            fmt.Printf("Error during backup of %s: %v\n", SRC_CONF, err)
            os.Exit(1)
        }
        fmt.Printf("Backup complete: '%s' has been backed up to '%s'.\n", SRC_CONF, BACKUP_CONF)

        // Backup certs directory
        certsInfo, err := os.Stat(SRC_CERTS)
        if err != nil || !certsInfo.IsDir() {
            fmt.Printf("Error: Source directory '%s' does not exist.\n", SRC_CERTS)
            os.Exit(1)
        }
        if err := utils.RemoveIfExists(BACKUP_CERTS); err != nil {
            fmt.Printf("Error removing backup directory '%s': %v\n", BACKUP_CERTS, err)
            os.Exit(1)
        }
        if err := utils.CopyDir(SRC_CERTS, BACKUP_CERTS); err != nil {
            fmt.Printf("Error during backup of %s: %v\n", SRC_CERTS, err)
            os.Exit(1)
        }
        fmt.Printf("Backup complete: '%s' has been backed up to '%s'.\n", SRC_CERTS, BACKUP_CERTS)

        // Backup docker-compose.yml file
        composeInfo, err := os.Stat(SRC_COMPOSE)
        if err != nil || composeInfo.IsDir() {
            fmt.Printf("Error: Source file '%s' does not exist.\n", SRC_COMPOSE)
            os.Exit(1)
        }
        if err := utils.RemoveIfExists(BACKUP_COMPOSE); err != nil {
            fmt.Printf("Error removing backup file '%s': %v\n", BACKUP_COMPOSE, err)
            os.Exit(1)
        }
        if err := utils.CopyFile(SRC_COMPOSE, BACKUP_COMPOSE); err != nil {
            fmt.Printf("Error during backup of %s: %v\n", SRC_COMPOSE, err)
            os.Exit(1)
        }
        fmt.Printf("Backup complete: '%s' has been backed up to '%s'.\n", SRC_COMPOSE, BACKUP_COMPOSE)
    },
}

func init() {
    // Assuming createCmd is defined in cmd/create.go,
    // we add backupCmd as a subcommand of createCmd.
    // If createCmd is not in the same package, import it accordingly.
    createCmd.AddCommand(backupCmd)
}

