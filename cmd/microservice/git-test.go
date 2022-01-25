package microservice

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dolittle/platform-api/pkg/git"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var gitTestCMD = &cobra.Command{
	Use:   "git-test",
	Short: "Test git",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		platformEnvironment := viper.GetString("tools.server.platformEnvironment")
		gitRepoConfig := git.InitGit(logContext, platformEnvironment)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		dir := "/tmp/dolittle-k8s/Source/V3/platform-api/453e04a7-4f9d-42f2-b36c-d51fa2c83fa3/11b6cf47-5d9f-438f-8116-0d9828654657/dev"
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println(err)
			return
		}
		microserviceID := "test"
		data := []byte(`hi 1`)
		filename := filepath.Join(dir, fmt.Sprintf("ms_%s.json", microserviceID))
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			fmt.Println("writeFile")
			fmt.Println(err)
			return
		}

		path := strings.TrimPrefix(filename, "/tmp/dolittle-k8s/")

		err = gitRepo.CommitPathAndPush(path, "upsert microservice")
		if err != nil {
			fmt.Println("CommitPathAndPush")
			fmt.Println(err)
			return
		}
	},
}
