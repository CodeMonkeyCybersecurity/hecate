package deploy

import (
	"fmt"

	"hecate/pkg/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// DeployCmd represents the deploy command
var DeployCmd = &cobra.Command{
	Use:   "deploy [app]",
	Short: "Deploy an application behind the Hecate reverse proxy",
	Args:  cobra.ExactArgs(1),
	Run:   runDeploy,
}

func runDeploy(cmd *cobra.Command, args []string) {
	app := args[0]
	logger := utils.GetLogger()

	logger.Info("Starting deployment", zap.String("app", app))

	// Deploy the application using utils
	if err := utils.DeployApp(app, cmd); err != nil {
		logger.Error("Deployment failed", zap.Error(err))
		fmt.Println("Error:", err)
	}
}
