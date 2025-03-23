// cmd/deploy/deploy.go

package deploy

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"hecate/pkg/utils"
	"hecate/pkg/logger"
	"hecate/pkg/config"
)

// DeployCmd represents the deploy command
var DeployCmd = &cobra.Command{
	Use:   "deploy [app]",
	Short: "Deploy an application behind the Hecate reverse proxy",
	Long: `Deploy applications behind Hecate’s reverse proxy.

This command allows you to deploy pre-configured applications such as Nextcloud, Jenkins, Wazuh, and others.
Hecate will automatically configure Nginx and deploy any necessary services.

Supported applications:
  - Nextcloud
  - Jenkins
  - Wazuh
  - Mailcow
  - Grafana
  - Mattermost
  - MinIO
  - Wiki.js
  - ERPNext
  - Persephone

Examples:

  # Deploy Nextcloud
  hecate deploy nextcloud

  # Force redeploy Jenkins (overwrite existing)
  hecate deploy jenkins --force
`,
	Args:  cobra.ExactArgs(1),
	Run:   runDeploy, // This generic function is used for non-specific deployments.
}


func runDeploy(cmd *cobra.Command, args []string) {
	app := strings.ToLower(args[0])

	// Validate the application.
	if !utils.IsValidApp(app) {
		fmt.Printf("❌ Invalid application: %s. Supported: %v\n", app, config.GetSupportedAppNames())
		return
	}

	fmt.Printf("🚀 Deploying application %s (force: %v)\n", app, force)

	// Proceed with deployment.
	if err := deployApplication(app); err != nil {
		fmt.Printf("❌ Deployment failed for '%s': %v\n", app, err)
		return
	}

	fmt.Printf("✅ Deployment completed successfully for %s\n", app)
}

func deployApplication(app string) error {
	if err := utils.DeployApp(app, force); err != nil {
		return fmt.Errorf("Deployment failed for '%s': %w", app, err)
	}
	return nil
}

func init() {
	DeployCmd.AddCommand(jenkins.NewDeployJenkinsCmd())
}
