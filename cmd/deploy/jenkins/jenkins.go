// cmd/deploy/jenkins/jenkins.go

package jenkins

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

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

			fmt.Println("🎉 Jenkins reverse proxy deployed successfully.")
		},
	}
	return cmd
}
