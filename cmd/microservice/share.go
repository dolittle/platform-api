package microservice

import (
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
		URL:           gitRepoURL,
		Branch:        gitRepoBranch,
		PrivateKey:    gitSshKeysFolder,
		RepoRoot:      gitLocalDirectory,
		DirectoryOnly: gitDirectoryOnly,
	}
}
