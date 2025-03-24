// cmd/deploy/jenkins/jenkins.go

package jenkins

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"hecate/pkg/certs"
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
		fmt.Printf("Configuration loaded:\n  Base Domain: %s\n  Backend IP: %s\n  Subdomain: %s\n  Email: %s\n",
		    cfg.BaseDomain, cfg.BackendIP, cfg.Subdomain, cfg.Email)

		// Define fullDomain using subdomain and base domain.
		fullDomain := fmt.Sprintf("%s.%s", cfg.Subdomain, cfg.BaseDomain)

		if err := certs.EnsureCertificates(cfg.Subdomain, cfg.BaseDomain, cfg.Email); err != nil {
		    log.Error("Certificate generation failed", zap.Error(err))
		    fmt.Printf("Certificate generation failed: %v\n", err)
		    return
		}
		log.Info("Certificate retrieved successfully", zap.String("domain", fullDomain))

		// Uncomment lines in docker-compose.yaml relevant to Jenkins.
		composeFile, err := UncommentSegment("docker-compose.yaml", "uncomment if using Jenkins behind Hecate")
		if err != nil {
		    log.Error("Failed to uncomment Jenkins section", zap.Error(err))
		    fmt.Printf("Failed to uncomment Jenkins section: %v\n", err)
		    return
		}
		log.Info("Successfully uncommented Jenkins lines", zap.String("composeFile", composeFile))

		
		// Now use the effective compose file for starting the services.
		err = docker.RunDockerComposeAllServices(composeFile)
		if err != nil {
		    log.Error("Failed to start Docker services", zap.Error(err))
		    fmt.Printf("Failed to run docker-compose up: %v\n", err)
		    return
		}
		
		fmt.Println("🎉 Jenkins reverse proxy deployed successfully.")
		},
	}
	return cmd
}
