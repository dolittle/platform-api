package create

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dolittle/platform-api/pkg/aiven"
)

var environmentCMD = &cobra.Command{
	Use:   "environment",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		apiToken := viper.GetString("tools.m3connector.aiven.apiToken")
		if apiToken == "" {
			fmt.Println("you have to provide an Aiven api token")
			return
		}
		project := "dolittle-test-env"
		service := "kafka-test-env"
		client := aiven.NewClient(apiToken, project, service)
		createUserResponse, err := client.CreateUser("joel-throwaway-test")

		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(createUserResponse)

		// addACLResponse, err := client.CreateACL("joel-throwaway-test-topic", "joel-throwaway-test", aiven.Admin)

		// if err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Println(addACLResponse)
	},
}
