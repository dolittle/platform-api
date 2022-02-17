package customertenant

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "customer-tenant",
	Short: "Customer tenants specific tooling",
	Long: `Customer tenants specific tooling
- Reducing manual steps in working with configs and the runtime`,
}

func init() {
	RootCmd.AddCommand(addCMD)
}
