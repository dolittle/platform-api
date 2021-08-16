package microservice

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var gitTestCMD = &cobra.Command{
	Use:   "git-test",
	Short: "Test git",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		gitRepoConfig := initGit(logContext)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		w, err := gitRepo.Repo.Worktree()
		if err != nil {
			fmt.Println(err)
			return
		}

		// TODO actually build structure
		// `{tenantID}/{applicationID}/{environment}/{microserviceID}.json`
		dir := "/tmp/dolittle-k8s/453e04a7-4f9d-42f2-b36c-d51fa2c83fa3/11b6cf47-5d9f-438f-8116-0d9828654657/dev"
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println(err)
			return
		}
		microserviceID := "test"
		data := []byte(`hi 1`)
		filename := fmt.Sprintf("%s/ms_%s.json", dir, microserviceID)
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			fmt.Println("writeFile")
			fmt.Println(err)
			return
		}

		// Adds the new file to the staging area.
		// Need to remove the prefix
		err = w.AddWithOptions(&git.AddOptions{
			Path: strings.TrimPrefix(filename, "/tmp/dolittle-k8s/"),
		})

		if err != nil {
			fmt.Println("w.Add")
			fmt.Println(err)
			return
		}

		_, err = w.Status()
		if err != nil {
			fmt.Println("w.Status")
			fmt.Println(err)
			return
		}

		err = gitRepo.CommitAndPush(w, "upsert microservice")
		if err != nil {
			fmt.Println("w.Status")
			fmt.Println(err)
			return
		}
	},
}
