package tools

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dolittle/platform-api/pkg/git"
	"github.com/dolittle/platform-api/pkg/platform/manual"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var getMicoservicesCMD = &cobra.Command{
	Use:   "get-microservices",
	Short: "Get all microservices",
	Long: `Search the git repo for all microservices

	GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev" \
	go run main.go tools get-microservices
	`,
	Run: func(cmd *cobra.Command, args []string) {

		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()

		gitLocalDirectory := viper.GetString("tools.server.gitRepo.directory")
		if gitLocalDirectory == "" {
			logContext.WithFields(logrus.Fields{
				"error": "GIT_REPO_DIRECTORY required",
			}).Fatal("start up")
		}

		// TODO ignore platform-api
		files, err := manual.GetMicroservicePaths(gitLocalDirectory)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("failed getting microservice paths")
		}

		for _, filePath := range files {
			info := manual.GetMicroserviceInfo(filePath)
			b, _ := json.Marshal(info)
			fmt.Println(string(b))
		}
	},
}

func init() {
	git.SetupViper()
}
