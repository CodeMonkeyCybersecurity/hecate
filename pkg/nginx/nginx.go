// pkg/nginx/nginx.go

package nginx

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"
	"github.com/spf13/cobra"
	
	"hecate/pkg/logger"
	"hecate/pkg/utils"
	"hecate/pkg/config"
	"hecate/pkg/docker"
	
)

//
//---------------------------- NGINX FUNCTIONS ---------------------------- //
//

// DeployApp deploys an application by copying necessary configs and restarting services
func DeployApp(app string, cmd *cobra.Command) error {
	logger.Info("Starting deployment", zap.String("app", app))  // ✅ Use logger.Info directly

	// Check if the required HTTP config exists
	httpConfig := filepath.Join(assetsPath, "servers", app+".conf")
	if !utils.FileExists(httpConfig) {
		logger.Error("Missing HTTP config file", zap.String("file", httpConfig))
		return fmt.Errorf("missing Nginx HTTP config for %s", app)
	}

	// Copy HTTP config
	if err := utils.CopyFile(httpConfig, filepath.Join(nginxConfPath, app+".conf")); err != nil {
		return fmt.Errorf("failed to copy HTTP config: %w", err)
	}

	// Copy Stream config if available
	streamConfig := filepath.Join(assetsPath, "stream", app+".conf")
	if utils.FileExists(streamConfig) {
		if err := utils.CopyFile(streamConfig, filepath.Join(nginxStreamPath, app+".conf")); err != nil {
			return fmt.Errorf("failed to copy Stream config: %w", err)
		}
	}

	// Handle NextCloud Coturn deployment
	if app == "nextcloud" {
		noTalk, _ := cmd.Flags().GetBool("without-talk")
		if !noTalk {
			logger.Info("Deploying Coturn for NextCloud Talk")
			if err := docker.RunDockerComposeService(dockerComposeFile, "coturn"); err != nil {
				return fmt.Errorf("failed to deploy Coturn: %w", err)
			}
		} else {
			logger.Info("Skipping Coturn deployment")
		}
	}

	// Validate and restart Nginx
	if err := ValidateNginx(); err != nil {
		return fmt.Errorf("invalid Nginx configuration: %w", err)
	}

	if err := RestartNginx(); err != nil {
		return fmt.Errorf("failed to restart Nginx: %w", err)
	}

	logger.Info("Deployment successful", zap.String("app", app))
	fmt.Printf("Successfully deployed %s!\n", app)
	return nil
}


// ValidateNginx runs `nginx -t` to check configuration validity
func ValidateNginx() error {
	logger.Info("Validating Nginx configuration...")
	cmd := exec.Command("nginx", "-t")
	err := cmd.Run()
	if err != nil {
		logger.Error("Nginx configuration validation failed", zap.Error(err))
	}
	return err
}

// RestartNginx reloads the Nginx service
func RestartNginx() error {
	logger.Info("Restarting Nginx...")
	cmd := exec.Command("systemctl", "reload", "nginx")
	err := cmd.Run()
	if err != nil {
		logger.Error("Failed to restart Nginx", zap.Error(err))
	}
	return err
}
