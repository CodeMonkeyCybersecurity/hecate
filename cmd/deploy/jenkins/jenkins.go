// cmd/deploy/jenkins/jenkins.go

package jenkins

import (
	"fmt"

	"github.com/spf13/cobra"
	"hecate/pkg/utils"
)

// NewDeployJenkinsCmd returns the Jenkins-specific deploy command.
func NewDeployJenkinsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jenkins",
		Short: "Deploy reverse proxy for Jenkins",
		Long: `Deploy the reverse proxy configuration for Jenkins using Hecate.

This command is designed to be extendable. For now, it primarily tests the asset organization step:
it organizes assets by moving all files in the "assets" directory that are not relevant to Jenkins
into the "other" directory (located at the project root).`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🚀 Deploying reverse proxy for Jenkins...")
			if err := utils.OrganizeAssetsForDeployment("jenkins"); err != nil {
				fmt.Printf("Failed to organize assets: %v\n", err)
				return
			}
			fmt.Println("🎉 Jenkins reverse proxy deployed successfully.")
		},
	}
	return cmd
}
