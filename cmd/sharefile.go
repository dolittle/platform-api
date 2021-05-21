package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-storage-file-go/azfile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var shareFileCmd = &cobra.Command{
	Use:   "share-file",
	Short: "Create link for a file in a share name linked to a specific customer",
	Long: `

	ACCOUNT_NAME=$(cat customers.json | jq -r '.[]|select(.tenant.name=="Customer-Chris")|.azure_storage_account.name') \
	ACCOUNT_KEY=$(cat customers.json | jq -r '.[]|select(.tenant.name=="Customer-Chris")|.azure_storage_account.key') \
	FILE_PATH="mongo/2021-01-20_16-07-04.gz.mongodump" \
	go run main.go share-file

	`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.BindEnv("tools.share.accountName", "ACCOUNT_NAME")
		viper.BindEnv("tools.share.accountKey", "ACCOUNT_KEY")
		viper.BindEnv("tools.share.file.shareName", "SHARE_NAME")
		viper.BindEnv("tools.share.file.path", "FILE_PATH")

		accountName := viper.GetString("tools.share.accountName")
		accountKey := viper.GetString("tools.share.accountKey")
		shareName := viper.GetString("tools.share.file.shareName")
		filePath := viper.GetString("tools.share.file.path")

		credential, err := azfile.NewSharedKeyCredential(accountName, accountKey)
		if err != nil {
			log.Fatal(err)
		}

		sasQueryParams, err := azfile.FileSASSignatureValues{
			Protocol:   azfile.SASProtocolHTTPS,
			ExpiryTime: time.Now().UTC().Add(48 * time.Hour),
			ShareName:  shareName,
			FilePath:   filePath,

			Permissions: azfile.FileSASPermissions{
				Read:  true,
				Write: false}.String(),
		}.NewSASQueryParameters(credential)

		if err != nil {
			log.Fatal(err)
		}

		qp := sasQueryParams.Encode()
		urlToSendToSomeone := fmt.Sprintf("https://%s.file.core.windows.net/%s/%s?%s",
			accountName, shareName, filePath, qp)
		fmt.Println(urlToSendToSomeone)
	},
}

func init() {
	rootCmd.AddCommand(shareFileCmd)
	viper.BindEnv("tools.share.accountName", "ACCOUNT_NAME")
	viper.BindEnv("tools.share.accountKey", "ACCOUNT_KEY")
	viper.BindEnv("tools.share.file.shareName", "SHARE_NAME")
	viper.BindEnv("tools.share.file.path", "FILE_PATH")
}
