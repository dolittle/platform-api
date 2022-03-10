package create

import (
	"github.com/dolittle/platform-api/pkg/cmd"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "create",
	Short: "Commands to create microservices",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(privateCMD)

	cmd.SetupStringConfiguration(RootCMD, "tools.microservice.create.applicationName", "application-name", "APPLICATION_NAME", "", "Microservices applications name")

	cmd.SetupStringConfiguration(RootCMD, "tools.microservice.create.microserviceName", "microservice-name", "MICROSERVICE_NAME", "", "Microservices name")

	cmd.SetupStringConfiguration(RootCMD, "tools.microservice.create.tenant.name", "tenant-name", "TENANT_NAME", "", "Microservices tenants name")
	cmd.SetupStringConfiguration(RootCMD, "tools.microservice.create.tenant.id", "tenant-id", "TENANT_ID", uuid.New().String(), "Microservices tenants name, (defaults to a randomly generated one)")

	cmd.SetupStringConfiguration(RootCMD, "tools.microservice.create.customerID", "customer-id", "CUSTOMER_ID", "", "Microservices customers ID (the owner of the microservice)")

	cmd.SetupStringConfiguration(RootCMD, "tools.microservice.create.headImage", "head-image", "HEAD_IMAGE", "", "Head image")
	cmd.SetupStringConfiguration(RootCMD, "tools.microservice.create.runtimeImage", "runtime-image", "RUNTIME_IMAGE", "", "Runtime image")
}
