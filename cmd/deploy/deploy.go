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
	Run:   runDeploy,
}

// ✅ New `force` flag (kept here because it's specific to this command)
var force bool

func runDeploy(cmd *cobra.Command, args []string) {
	log := logger.GetSafeLogger()
	if log == nil {
		fmt.Println("⚠️ Warning: Logger is nil. Defaulting to console output.")
	}
	
	app := strings.ToLower(args[0]) // ✅ Normalize input to lowercase

	// ✅ Validate application before proceeding
	if !utils.IsValidApp(app) {
		errMsg := fmt.Sprintf("❌ Invalid application: %s. Supported: %v", app, config.GetSupportedAppNames())
		utils.PrintError(log, errMsg)
		return
	}	

	utils.SafeLog(log, "🚀 Deploying application", zap.String("app", app), zap.Bool("force", utils.GetForce()))
	
	// ✅ Proceed with deployment
	if err := deployApplication(app, cmd); err != nil {
		utils.PrintError(log, fmt.Sprintf("❌ Deployment failed for '%s': %v", app, err))
		return
	}

	utils.SafeLog(log, "✅ Deployment completed successfully", zap.String("app", app))
}


// ✅ Deployment wrapper function
func deployApplication(app string, cmd *cobra.Command) error {
	if err := utils.DeployApp(app, utils.GetForce()); err != nil {
		return fmt.Errorf("❌ Deployment failed for '%s': %w", app, err)
	}
	return nil
}

// ✅ Ensure `force` flag is handled correctly
func init() {
	DeployCmd.Flags().BoolVarP(&force, "force", "f", false, "Force redeployment (overwrite existing)")
	DeployCmd.PreRun = func(cmd *cobra.Command, args []string) {
		utils.SetForce(force)
		if force { // ✅ Only log when force is actually enabled
			utils.SafeLog(logger.GetLogger(), "⚠️ Force mode enabled: Existing deployments may be overwritten.")
		}
	}
}
