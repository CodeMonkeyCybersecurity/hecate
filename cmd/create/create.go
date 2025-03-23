// cmd/create/create.go
package create

import (
    "github.com/spf13/cobra"
    "go.uber.org/zap"

    "hecate/cmd/create/hetzner"
    "hecate/pkg/logger"
)

var log = logger.GetLogger()

// CreateCmd represents the create command
var CreateCmd = &cobra.Command{
    Use:   "create",
    Short: "Create resources for Hecate",
    Long: `The create command allows you to create specific resources
needed for your Hecate deployment, such as certificates, proxy configurations, etc.`,
    Run: func(cmd *cobra.Command, args []string) {
        cmd.Println("Create command executed!")
    },
}

// init gets called automatically at package load time
func init() {
    // Attach the hetzner-wildcard subcommand here
    CreateCmd.AddCommand(hetzner.NewCreateHetznerWildcardCmd())
}
