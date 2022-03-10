package create

import (
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCMD = &cobra.Command{
	Use:   "create",
	Short: "Commands to create microservices",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(privateCMD)

	setupStringConfiguration(RootCMD, "tools.microservice.create.application.id", "application-id", "APPLICATION_ID", "", "Microservices applications ID")
	setupStringConfiguration(RootCMD, "tools.microservice.create.application.name", "application-name", "APPLICATION_NAME", "", "Microservices applications name")

	setupStringConfiguration(RootCMD, "tools.microservice.create.microservice.id", "microservice-id", "MICROSERVICE_ID", uuid.New().String(), "Microservices ID (defaults to a randomly generated one)")
	setupStringConfiguration(RootCMD, "tools.microservice.create.microservice.name", "microservice-name", "MICROSERVICE_NAME", "", "Microservices name")

	setupStringConfiguration(RootCMD, "tools.microservice.create.environment", "environment", "ENVIRONMENT", "", "Microservices environment")

	setupStringConfiguration(RootCMD, "tools.microservice.create.tenant.name", "tenant-name", "TENANT_NAME", "", "Microservices tenants name")
	setupStringConfiguration(RootCMD, "tools.microservice.create.tenant.id", "tenant-id", "TENANT_ID", uuid.New().String(), "Microservices tenants name, (defaults to a randomly generated one)")

	setupStringConfiguration(RootCMD, "tools.microservice.create.customerID", "customer-id", "CUSTOMER_ID", "", "Microservices customers ID (the owner of the microservice)")

	setupStringConfiguration(RootCMD, "tools.microservice.create.headImage", "head-image", "HEAD_IMAGE", "", "Head image")
	setupStringConfiguration(RootCMD, "tools.microservice.create.runtimeImage", "runtime-image", "RUNTIME_IMAGE", "", "Runtime image")
}

func setupStringConfiguration(cmd *cobra.Command, key, flag, envVarName, defaultValue, description string) {
	viper.SetDefault(key, defaultValue)
	viper.BindEnv(key, envVarName)
	cmd.PersistentFlags().String(flag, defaultValue, description)
	viper.BindPFlag(key, cmd.PersistentFlags().Lookup(flag))
}
