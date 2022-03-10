package microservice

import (
	"github.com/dolittle/platform-api/cmd/tools/microservice/create"
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "microservice",
	Short: "Commands to manage microservices",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(create.RootCMD)
}
