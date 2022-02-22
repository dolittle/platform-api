package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/spf13/cobra"
)

var deleteCustomerCMD = &cobra.Command{
	Use:   "delete-customer",
	Short: "Shows commands to aid in deleting a customer from the cluster",
	Long: `
	go run main.go tools helpers commands delete-customer --directory="/Users/freshteapot/dolittle/git/Operations" 6677c2f0-9e2f-4d2b-beb5-50014fc8ad0c
	`,
	Run: func(cmd *cobra.Command, args []string) {
		rootDir, _ := cmd.Flags().GetString("directory")
		rootDir = strings.TrimSuffix(rootDir, string(os.PathSeparator))

		platformApiDir := filepath.Join(rootDir, "Source", "V3", "platform-api")
		azureDir := filepath.Join(rootDir, "Source", "V3", "Azure")

		customerID := args[0]

		type storageInfo struct {
			CustomerID          string
			ApplicationID       string
			PlatformEnvironment string
			Path                string
		}
		items := make([]storageInfo, 0)

		err := filepath.Walk(platformApiDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				return nil
			}

			_path := strings.TrimPrefix(path, platformApiDir)
			_path = strings.TrimPrefix(_path, string(os.PathSeparator))

			parts := strings.Split(_path, string(os.PathSeparator))

			if len(parts) > 3 {
				return filepath.SkipDir
			}

			if !strings.Contains(path, customerID) {
				return nil
			}

			if len(parts) != 3 {
				return nil
			}

			items = append(items, storageInfo{
				PlatformEnvironment: parts[0],
				CustomerID:          parts[1],
				ApplicationID:       parts[2],
				Path:                path,
			})
			return nil
		})

		if err != nil {
			panic(err)
		}

		for _, item := range items {
			namespace := platformK8s.GetApplicationNamespace(item.ApplicationID)
			tfPrefix := fmt.Sprintf("customer_%s_%s", item.CustomerID, item.ApplicationID)
			cmds := []string{
				// Delete from cluster
				fmt.Sprintf(`kubectl delete namespace %s`, namespace),
				// Delete from storage
				fmt.Sprintf(`rm -r %s`, item.Path),
				// Delete from terraform
				fmt.Sprintf("cd %s", azureDir),
				"TODO delete from terraform",
				fmt.Sprintf(`terraform destroy -target="module.%s"`, tfPrefix),
				fmt.Sprintf(`rm -r %s/%s.tf"`, azureDir, tfPrefix),
				"terraform apply",
			}
			output := strings.Join(cmds, "\n")
			fmt.Println(output)
		}
	},
}

func init() {
	deleteCustomerCMD.Flags().String("directory", "", "Path to git repo")
}
