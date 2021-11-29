package git

import (
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func SetupViper() {
	viper.BindEnv("tools.server.gitRepo.sshKey", "GIT_REPO_SSH_KEY")
	viper.BindEnv("tools.server.gitRepo.branch", "GIT_REPO_BRANCH")
	viper.BindEnv("tools.server.gitRepo.url", "GIT_REPO_URL")
	viper.BindEnv("tools.server.gitRepo.directory", "GIT_REPO_DIRECTORY")
	viper.BindEnv("tools.server.gitRepo.directoryOnly", "GIT_REPO_DIRECTORY_ONLY")
	viper.BindEnv("tools.server.gitRepo.dryRun", "GIT_REPO_DRY_RUN")

	viper.SetDefault("tools.server.gitRepo.sshKey", "")
	viper.SetDefault("tools.server.gitRepo.url", "")
	viper.SetDefault("tools.server.gitRepo.directory", "/tmp/dolittle-k8s")
	viper.SetDefault("tools.server.gitRepo.directoryOnly", false)
	viper.SetDefault("tools.server.gitRepo.dryRun", false)
}

func InitGit(logContext logrus.FieldLogger) gitStorage.GitStorageConfig {
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

	gitDryRun := viper.GetBool("tools.server.gitRepo.dryRun")

	return gitStorage.GitStorageConfig{
		URL:           gitRepoURL,
		Branch:        gitRepoBranch,
		PrivateKey:    gitSshKeysFolder,
		RepoRoot:      gitLocalDirectory,
		DirectoryOnly: gitDirectoryOnly,
		DryRun:        gitDryRun,
	}
}
