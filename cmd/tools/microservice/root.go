package microservice

import (
	"github.com/dolittle/platform-api/cmd/tools/microservice/create"
	"github.com/dolittle/platform-api/pkg/cmd"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "microservice",
	Short: "Commands to manage microservices",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(create.RootCMD)
	cmd.SetupStringConfiguration(RootCMD, "tools.microservice.microserviceId", "microservice-id", "MICROSERVICE_ID", uuid.New().String(), "Microservices ID")
	cmd.SetupStringConfiguration(RootCMD, "tools.microservice.applicationID", "application-id", "APPLICATION_ID", "", "Microservcies applications ID")
	cmd.SetupStringConfiguration(RootCMD, "tools.microservice.environment", "environment", "ENVIRONMNET", "", "Microservcies environment")
}
