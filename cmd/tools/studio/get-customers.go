package studio

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dolittle/platform-api/pkg/git"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// TODO should we deprecate this? or make it reusable in terms of add the "name" and it will hook up the developer rbac
var getCustomersCMD = &cobra.Command{
	Use:   "get-customers",
	Short: "Get customer info from studio storage",
	Long: `
	Attempts to get customer.json

	GIT_REPO_BRANCH=dev \
	GIT_REPO_DRY_RUN=true \
	GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev" \
	GIT_REPO_DIRECTORY_ONLY=true \
	go run main.go tools studio get-customers
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		logContext := logrus.StandardLogger()
		platformEnvironment, _ := cmd.Flags().GetString("platform-environment")
		gitRepoConfig := git.InitGit(logContext, platformEnvironment)

		storageRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		customers, err := storageRepo.GetCustomers()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		b, _ := json.Marshal(customers)
		fmt.Println(string(b))
	},
}

func init() {
	getCustomersCMD.Flags().String("platform-environment", "dev", "Platform environment (dev or prod), not linked to application environment")
}
