package containerregistry

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/services/preview/containerregistry/runtime/2019-07/containerregistry"
	"github.com/Azure/go-autorest/autorest"
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

	username := credentials.Username
	password := credentials.Password
	basicAuthorizer := autorest.NewBasicAuthorizer(username, password)

	baseClient := containerregistry.New(credentials.URL)
	baseClient.Authorizer = basicAuthorizer

	ctx := context.Background()
	var n int32 = 200
	result, err := baseClient.GetRepositories(ctx, "", &n)
	if err != nil {
		repo.logContext.WithField("error", err).Error("failed to get images")
		return make([]string, 0), errors.New("failed to get images")
	}

	return *result.Names, nil
}

func (repo azureRepo) GetTags(credentials ContainerRegistryCredentials, image string) ([]string, error) {
	ctx := context.Background()

	username := credentials.Username
	password := credentials.Password
	basicAuthorizer := autorest.NewBasicAuthorizer(username, password)

	baseClient := containerregistry.New(credentials.URL)
	baseClient.Authorizer = basicAuthorizer
	tagResult, err := baseClient.GetTagList(ctx, image)

	if err != nil {
		repo.logContext.WithField("error", err).Error("failed to get tags")
		return make([]string, 0), errors.New("failed to get tags")
	}

	return *tagResult.Tags, nil
}
