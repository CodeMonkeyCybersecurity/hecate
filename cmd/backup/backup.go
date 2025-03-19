/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/

package cmd

import (
    "fmt"
    "os"
    "time"

    "hecate/pkg/utils"
    "github.com/spf13/cobra"
)

// backupCmd represents the backup command.
var backupCmd = &cobra.Command{
    Use:   "backup",
    Short: "Backup configuration and files",
    Long:  `Backup important configuration directories and files.`,
    Run: func(cmd *cobra.Command, args []string) {
        runBackup()
    },
}

func init() {
    // Add backupCmd as a subcommand of createCmd
    createCmd.AddCommand(backupCmd)
}

// runBackup is called when the user runs "hecate create backup".
func runBackup() {
    // 1) Define your source directories / files.
    const (
        SRC_CONF    = "conf.d"
        SRC_CERTS   = "certs"
        SRC_COMPOSE = "docker-compose.yml"
    )

    // 2) Build a timestamp string
    timestamp := time.Now().Format("20060102-150405")

    // 3) Generate unique backup names
    backupConf    := fmt.Sprintf("%s.%s.bak", SRC_CONF, timestamp)
    backupCerts   := fmt.Sprintf("%s.%s.bak", SRC_CERTS, timestamp)
    backupCompose := fmt.Sprintf("%s.%s.bak", SRC_COMPOSE, timestamp)

    // === Backup conf.d directory ===
    confInfo, err := os.Stat(SRC_CONF)
    if err != nil || !confInfo.IsDir() {
        fmt.Printf("Error: Source directory '%s' does not exist.\n", SRC_CONF)
        os.Exit(1)
    }
    if err := utils.RemoveIfExists(backupConf); err != nil {
        fmt.Printf("Error removing backup directory '%s': %v\n", backupConf, err)
        os.Exit(1)
    }
    if err := utils.CopyDir(SRC_CONF, backupConf); err != nil {
        fmt.Printf("Error during backup of %s: %v\n", SRC_CONF, err)
        os.Exit(1)
    }
    fmt.Printf("Backup complete: '%s' has been backed up to '%s'.\n", SRC_CONF, backupConf)

    // === Backup certs directory ===
    certsInfo, err := os.Stat(SRC_CERTS)
    if err != nil || !certsInfo.IsDir() {
        fmt.Printf("Error: Source directory '%s' does not exist.\n", SRC_CERTS)
        os.Exit(1)
    }
    if err := utils.RemoveIfExists(backupCerts); err != nil {
        fmt.Printf("Error removing backup directory '%s': %v\n", backupCerts, err)
        os.Exit(1)
    }
    if err := utils.CopyDir(SRC_CERTS, backupCerts); err != nil {
        fmt.Printf("Error during backup of %s: %v\n", SRC_CERTS, err)
        os.Exit(1)
    }
    fmt.Printf("Backup complete: '%s' has been backed up to '%s'.\n", SRC_CERTS, backupCerts)

    // === Backup docker-compose.yml file ===
    composeInfo, err := os.Stat(SRC_COMPOSE)
    if err != nil || composeInfo.IsDir() {
        fmt.Printf("Error: Source file '%s' does not exist.\n", SRC_COMPOSE)
        os.Exit(1)
    }
    if err := utils.RemoveIfExists(backupCompose); err != nil {
        fmt.Printf("Error removing backup file '%s': %v\n", backupCompose, err)
        os.Exit(1)
    }
    if err := utils.CopyFile(SRC_COMPOSE, backupCompose); err != nil {
        fmt.Printf("Error during backup of %s: %v\n", SRC_COMPOSE, err)
        os.Exit(1)
    }
    fmt.Printf("Backup complete: '%s' has been backed up to '%s'.\n", SRC_COMPOSE, backupCompose)
}
