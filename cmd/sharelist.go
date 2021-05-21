package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sort"

	"github.com/Azure/azure-storage-file-go/azfile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ListResponse struct {
	AccountName string   `json:"account_name"`
	ShareName   string   `json:"share_name"`
	Prefix      string   `json:"prefix"`
	Files       []string `json:"files"`
}

// https://docs.microsoft.com/en-us/azure/storage/blobs/storage-blob-event-overview

var shareListCmd = &cobra.Command{
	Use:   "share-list",
	Short: "Get list of databases for a specific customer and share name",
	Long: `

	ACCOUNT_NAME=$(cat customers.json | jq -r '.[]|select(.tenant.name=="Customer-Chris")|.azure_storage_account.name') \
	ACCOUNT_KEY=$(cat customers.json | jq -r '.[]|select(.tenant.name=="Customer-Chris")|.azure_storage_account.key') \
	SHARE_NAME=taco-dev-backup \
	go run main.go share-list

	`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.BindEnv("tools.share.accountName", "ACCOUNT_NAME")
		viper.BindEnv("tools.share.accountKey", "ACCOUNT_KEY")
		viper.BindEnv("tools.share.file.shareName", "SHARE_NAME")

		accountName := viper.GetString("tools.share.accountName")
		accountKey := viper.GetString("tools.share.accountKey")
		shareName := viper.GetString("tools.share.file.shareName")

		credential, err := azfile.NewSharedKeyCredential(accountName, accountKey)
		if err != nil {
			log.Fatal(err)
		}

		p := azfile.NewPipeline(credential, azfile.PipelineOptions{})
		u, _ := url.Parse(fmt.Sprintf("https://%s.file.core.windows.net", accountName))
		serviceURL := azfile.NewServiceURL(*u, p)

		ctx := context.Background()
		shareURL := serviceURL.NewShareURL(shareName)
		directoryURL := shareURL.NewDirectoryURL("mongo")

		found := make([]string, 0)
		for marker := (azfile.Marker{}); marker.NotDone(); {
			// Prefix is to strict
			listResponse, err := directoryURL.ListFilesAndDirectoriesSegment(ctx, marker, azfile.ListFilesAndDirectoriesOptions{})

			if err != nil {
				log.Fatal(err)
			}

			marker = listResponse.NextMarker

			for _, fileEntry := range listResponse.FileItems {
				found = append(found, fmt.Sprintf("%s/%s", directoryURL.URL().Path, fileEntry.Name))
			}
		}

		// Get the latest files, due to timestamp
		last := 20
		size := len(found)
		if size < last {
			last = size
		}
		latest := found[size-last:]
		sort.Sort(sort.Reverse(sort.StringSlice(latest)))

		b, _ := json.Marshal(ListResponse{
			AccountName: accountName,
			ShareName:   shareName,
			Prefix:      serviceURL.String(),
			Files:       latest,
		})
		fmt.Println(string(b))
	},
}

func init() {
	rootCmd.AddCommand(shareListCmd)
	viper.BindEnv("tools.share.accountName", "ACCOUNT_NAME")
	viper.BindEnv("tools.share.accountKey", "ACCOUNT_KEY")
	viper.BindEnv("tools.share.file.shareName", "SHARE_NAME")
}
