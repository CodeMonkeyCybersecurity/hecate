// cmd/create/create.go

package create

import (
	"github.com/spf13/cobra"
)

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

func init() {
	root.RootCmd.AddCommand(CreateCmd) // ✅ Attach CreateCmd to rootCmd
}
