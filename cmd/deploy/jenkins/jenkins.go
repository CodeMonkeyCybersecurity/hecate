// cmd/deploy/jenkins/jenkins.go

package jenkins

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"hecate/pkg/config"
	"hecate/pkg/docker"
	"hecate/pkg/logger"
	"hecate/pkg/utils"
)

var log = logger.GetLogger()

// NewDeployJenkinsCmd returns the Jenkins-specific deploy command.
func NewDeployJenkinsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jenkins",
		Short: "Deploy reverse proxy for Jenkins",
		Long: `Deploy the reverse proxy configuration for Jenkins using Hecate.

This command stops the Hecate container (if running) and then organizes assets by moving files 
that are not relevant to Jenkins into the "other" directory at the project root.`,
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("Starting Jenkins deployment")

			// Stop the container if it's running.
			if err := docker.StopContainersBySubstring("hecate"); err != nil {
				log.Error("Error stopping container", zap.String("substring", "hecate"), zap.Error(err))
				fmt.Printf("Error stopping container: %v\n", err)
				return
			}
			log.Info("Containers with 'hecate' in the name stopped successfully")

			// Organize assets for Jenkins.
			if err := utils.OrganizeAssetsForDeployment("jenkins"); err != nil {
				log.Error("Failed to organize assets", zap.Error(err))
				fmt.Printf("Failed to organize assets: %v\n", err)
				return
			}
			log.Info("Assets organized successfully for Jenkins")

			// Load configuration from .hecate.conf.
			cfg, err := config.LoadConfig("jenkins")
			if err != nil {
			    log.Error("Configuration error", zap.Error(err))
			    fmt.Printf("Configuration error: %v\n", err)
			    return
			}
			log.Info("Configuration loaded", zap.Any("config", cfg))
			fmt.Println("Configuration loaded:")
			fmt.Printf("  Base Domain: %s\n", cfg.BaseDomain)
			fmt.Printf("  Backend IP: %s\n", cfg.BackendIP)
			fmt.Printf("  Subdomain: %s\n", cfg.Subdomain)
			fmt.Printf("  Email: %s\n", cfg.Email)

			// Here you could add additional certificate logic if desired.
			// For example:
			// if err := certs.EnsureCertificates(cfg.Subdomain, cfg.Email); err != nil {
			//     log.Error("Certificate error", zap.Error(err))
			//     fmt.Printf("Certificate error: %v\n", err)
			//     return
			// }

			fmt.Println("🎉 Jenkins reverse proxy deployed successfully.")
		},
	}
	return cmd
}
