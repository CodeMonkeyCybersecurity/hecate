package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"hecate/pkg/utils" // Ensure this exists
	"hecate/pkg/config"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// DeployCmd represents the deploy command
var DeployCmd = &cobra.Command{
	Use:   "deploy [app]",
	Short: "Deploy an application behind the Hecate reverse proxy",
	Long: `Deploys an Nginx reverse proxy configuration for a specified application.

Example:
  hecate deploy nextcloud
  hecate deploy wazuh
  hecate deploy jenkins`,
	Args: cobra.ExactArgs(1), // Ensure exactly one argument is provided
	Run:  runDeploy,
}

func init() {
	rootCmd.AddCommand(DeployCmd)
}
