// cmd/backup/backup.go

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/

package backup

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"hecate/pkg/config"
	"hecate/pkg/logger"
	"hecate/pkg/utils"
)

var log = logger.GetLogger()

// backupCmd represents the backup command.
var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup configuration and files",
	Long:  `Backup important configuration directories and files.`,
	Run: func(cmd *cobra.Command, args []string) {
		runBackup()
	},
}

// runBackup is called when the user runs "hecate create backup".
func runBackup() {
	log := logger.GetLogger()
	if log == nil {
		println("⚠️ Logger is nil. Defaulting to console output.")
	}

	timestamp := time.Now().Format("20060102-150405")
	backupConf := config.DefaultConfDir + "." + timestamp + ".bak"
	backupCerts := config.DefaultCertsDir + "." + timestamp + ".bak"
	backupCompose := config.DefaultComposeYML + "." + timestamp + ".bak"

	// conf.d
	if info, err := os.Stat(config.DefaultConfDir); err != nil || !info.IsDir() {
		log.Error("Missing or invalid conf.d", zap.String("dir", config.DefaultConfDir), zap.Error(err))
		os.Exit(1)
	}
	if err := utils.RemoveIfExists(backupConf); err != nil {
		log.Error("Failed to remove existing backup", zap.String("path", backupConf), zap.Error(err))
		os.Exit(1)
	}

	if err := utils.CopyDir(config.DefaultConfDir, backupConf); err != nil {
		log.Error("Backup failed", zap.String("src", config.DefaultConfDir), zap.Error(err))
		os.Exit(1)
	}
	log.Info("✅ conf.d backed up", zap.String("dest", backupConf))

	// certs
	if info, err := os.Stat(config.DefaultCertsDir); err != nil || !info.IsDir() {
		log.Error("Missing or invalid certs", zap.String("dir", config.DefaultCertsDir), zap.Error(err))
		os.Exit(1)
	}
	if err := utils.RemoveIfExists(backupCerts); err != nil {
		log.Error("Failed to remove existing backup", zap.String("path", backupCerts), zap.Error(err))
		os.Exit(1)
	}
	if err := utils.CopyDir(config.DefaultCertsDir, backupCerts); err != nil {
		log.Error("Backup failed", zap.String("src", config.DefaultCertsDir), zap.Error(err))
		os.Exit(1)
	}
	log.Info("✅ certs backed up", zap.String("dest", backupCerts))

	// docker-compose.yml
	if info, err := os.Stat(config.DefaultComposeYML); err != nil || info.IsDir() {
		log.Error("Missing or invalid compose file", zap.String("file", config.DefaultComposeYML), zap.Error(err))
		os.Exit(1)
	}
	if err := utils.RemoveIfExists(backupCompose); err != nil {
		log.Error("Failed to remove existing backup", zap.String("path", backupCompose), zap.Error(err))
		os.Exit(1)
	}
	if err := utils.CopyFile(config.DefaultComposeYML, backupCompose); err != nil {
		log.Error("Backup failed", zap.String("src", config.DefaultComposeYML), zap.Error(err))
		os.Exit(1)
	}
	log.Info("✅ docker-compose.yml backed up", zap.String("dest", backupCompose))
	log.Info("🎉 All backup tasks completed successfully", zap.String("timestamp", timestamp))
}
