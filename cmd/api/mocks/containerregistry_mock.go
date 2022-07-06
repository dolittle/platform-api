package mocks

import "github.com/dolittle/platform-api/pkg/platform/containerregistry"

type ContainerRegistryMock struct {
	imagesResult []string
	tagResult    []string
	tagResult2   []containerregistry.ImageTag
}

func (m *ContainerRegistryMock) StubAndReturnImages(result []string) {
	m.imagesResult = result
}

func (m *ContainerRegistryMock) StubAndReturnTags(result []string) {
	m.tagResult = result
}

func (m *ContainerRegistryMock) StubAndReturnTags2(result []containerregistry.ImageTag) {
	m.tagResult2 = result
}

func (m *ContainerRegistryMock) GetImages(credentials containerregistry.ContainerRegistryCredentials) ([]string, error) {
	return m.imagesResult, nil
}

func (m *ContainerRegistryMock) GetTags(credentials containerregistry.ContainerRegistryCredentials, image string) ([]string, error) {
	return m.tagResult, nil
}

func (m *ContainerRegistryMock) GetTags2(credentials containerregistry.ContainerRegistryCredentials, image string) ([]containerregistry.ImageTag, error) {
	return m.tagResult2, nil
}
