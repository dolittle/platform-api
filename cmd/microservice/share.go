package microservice

import (
	"fmt"

	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func initGit(logContext logrus.FieldLogger) gitStorage.GitStorageConfig {
	gitDirectoryOnly := viper.GetBool("tools.server.gitRepo.directoryOnly")
	gitRepoURL := ""
	gitSshKeysFolder := ""

	gitRepoBranch := viper.GetString("tools.server.gitRepo.branch")
	if gitRepoBranch == "" {
		logContext.WithFields(logrus.Fields{
			"error": "GIT_REPO_BRANCH required",
		}).Fatal("start up")
	}

	gitLocalDirectory := viper.GetString("tools.server.gitRepo.directory")
	if gitLocalDirectory == "" {
		logContext.WithFields(logrus.Fields{
			"error": "GIT_REPO_DIRECTORY required",
		}).Fatal("start up")
	}
	fmt.Println(gitLocalDirectory)
	/*
			GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev" \
		GIT_REPO_DIRECTORY_ONLY="true" \
		GIT_REPO_BRANCH=main \
		LISTEN_ON="localhost:8080" \
		HEADER_SECRET="FAKE" \
		AZURE_SUBSCRIPTION_ID="e7220048-8a2c-4537-994b-6f9b320692d7" \
		go run main.go microservice server --kube-config=$(k3d kubeconfig write dolittle-dev)
	*/
	if !gitDirectoryOnly {
		gitRepoURL = viper.GetString("tools.server.gitRepo.url")
		if gitRepoURL == "" {
			logContext.WithFields(logrus.Fields{
				"error": "GIT_REPO_URL required",
			}).Fatal("start up")
		}

		gitSshKeysFolder = viper.GetString("tools.server.gitRepo.sshKey")
		if gitSshKeysFolder == "" {
			logContext.WithFields(logrus.Fields{
				"error": "GIT_REPO_SSH_KEY required",
			}).Fatal("start up")
		}
	}

	return gitStorage.GitStorageConfig{
		URL:            gitRepoURL,
		Branch:         gitRepoBranch,
		PrivateKey:     gitSshKeysFolder,
		LocalDirectory: gitLocalDirectory,
		DirectoryOnly:  gitDirectoryOnly,
	}
}
