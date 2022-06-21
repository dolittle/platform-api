package microservice

import (
	"github.com/dolittle/platform-api/cmd/tools/microservice/copy"
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "microservice",
	Short: "Tooling microservices",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(copy.RootCMD)
}
