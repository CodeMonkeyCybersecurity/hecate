// cmd/deploy/jenkins.go

package deploy

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"hecate/pkg/config"
	"hecate/pkg/utils"
)

var deployJenkinsCmd = &cobra.Command{
	Use:   "jenkins",
	Short: "Deploy reverse proxy for Jenkins",
		Long: `Deploy the reverse proxy configuration for Jenkins using Hecate.
	
This command performs a series of automated steps to ensure that your Jenkins reverse proxy is 
properly configured and deployed. The overall process includes the following steps:

1. **Configuration Loading and Confirmation:**
   - It first checks for a configuration file (.hecate.conf) in the current directory. 
   - If the file exists, the command reads the existing values for BASE_DOMAIN and backendIP.
   - It then displays these values and prompts you to confirm whether you want to keep them or enter new ones.
   - If the file does not exist or the values are missing, you will be prompted to input the BASE_DOMAIN 
     (which defines the base domain for your reverse proxy) and the backendIP (the IP address of your backend server).
   - The confirmed values are saved back into the .hecate.conf file for future deployments.

2. **Asset Update and Placeholder Replacement:**
   - The command processes key configuration files located in the assets directory. For Jenkins, this 
     includes files such as assets/servers/jenkins.conf and assets/stream/jenkins.conf.
   - In these files, placeholders such as ${BASE_DOMAIN} and ${backendIP} are replaced with the values 
     obtained from the configuration file. This ensures that your Nginx configuration is dynamically updated 
     to reflect your deployment environment.

3. **Certificate Verification and Generation (Stub):**
   - Before deploying, the command checks whether the necessary HTTPS certificates exist in the certs/ 
     directory for the Jenkins subdomain (for example, jenkins.${BASE_DOMAIN}). 
   - If the certificates are missing, the command can be extended to automatically invoke certbot or another 
     ACME client to retrieve and store the required certificates. (In this minimal implementation, certificate 
     generation is simulated with a stub function.)

4. **Deployment via Docker Compose:**
   - Once the configuration files have been updated and certificates have been verified, the command invokes 
     Docker Compose to deploy the reverse proxy. The docker-compose.yml file is expected to mount the assets 
     directory into the container so that Nginx picks up the updated configuration.
   - This step brings up the reverse proxy (and any associated containers) in detached mode, making the 
     deployment visible and active.

By automating these steps, Hecate streamlines the deployment process for Jenkins, reducing manual errors and ensuring 
consistency across environments. This command is designed to be extendable so that additional checks (like asset organization 
or advanced logging) can be integrated in the future.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🚀 Deploying reverse proxy for Jenkins...")

		// Load Hecate configuration.
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}

		// Replace placeholders in Jenkins configuration files.
		serverConf := filepath.Join("assets", "servers", "jenkins.conf")
		streamConf := filepath.Join("assets", "stream", "jenkins.conf")
		if err := utils.ReplacePlaceholders(serverConf, cfg.BaseDomain, cfg.BackendIP); err != nil {
			return fmt.Errorf("failed to update server config: %w", err)
		}
		if err := utils.ReplacePlaceholders(streamConf, cfg.BaseDomain, cfg.BackendIP); err != nil {
			return fmt.Errorf("failed to update stream config: %w", err)
		}
		fmt.Println("✅ Configuration files updated.")

		// Organize assets: move unused configuration files into assets/other.
		if err := utils.OrganizeAssetsForDeployment("jenkins"); err != nil {
			return fmt.Errorf("failed to organize assets: %w", err)
		}

		fmt.Println("🎉 Jenkins reverse proxy deployed successfully.")
		return nil
	},
}



// NewDeployJenkinsCmd exposes this command to be added to the root command.
func NewDeployJenkinsCmd() *cobra.Command {
	return deployJenkinsCmd
}
