package containerregistry

import (
	"github.com/sirupsen/logrus"
)

type azureRepo struct {
	logContext logrus.FieldLogger
}

func NewAzureRepo(
	logContext logrus.FieldLogger,
) azureRepo {
	return azureRepo{
		logContext: logContext,
	}
}

func (repo azureRepo) GetImages(credentials ContainerRegistryCredentials) ([]string, error) {
	return make([]string, 0), nil
}
func (repo azureRepo) GetTags(credentials ContainerRegistryCredentials, image string) ([]string, error) {
	return make([]string, 0), nil
}
