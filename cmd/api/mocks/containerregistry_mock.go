package mocks

import "github.com/dolittle/platform-api/pkg/platform/containerregistry"

type ContainerRegistryMock struct {
	imagesResult []string
	tagResult    []string
	imageTags    []containerregistry.ImageTag
}

func (m *ContainerRegistryMock) StubAndReturnImages(result []string) {
	m.imagesResult = result
}

func (m *ContainerRegistryMock) StubAndReturnTags(result []string) {
	m.tagResult = result
}

func (m *ContainerRegistryMock) StubAndReturnImageTags(result []containerregistry.ImageTag) {
	m.imageTags = result
}

func (m *ContainerRegistryMock) GetImages(credentials containerregistry.ContainerRegistryCredentials) ([]string, error) {
	return m.imagesResult, nil
}

func (m *ContainerRegistryMock) GetTags(credentials containerregistry.ContainerRegistryCredentials, image string) ([]string, error) {
	return m.tagResult, nil
}

func (m *ContainerRegistryMock) GetImageTags(credentials containerregistry.ContainerRegistryCredentials, image string) ([]containerregistry.ImageTag, error) {
	return m.imageTags, nil
}
