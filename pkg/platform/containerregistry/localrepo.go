package containerregistry

type localContainerRegistryRepo struct {
}

func NewLocalRepo() localContainerRegistryRepo {
	return localContainerRegistryRepo{}
}

func (r localContainerRegistryRepo) GetImages(credentials ContainerRegistryCredentials) ([]string, error) {
	return []string{
		"hello",
	}, nil
}

func (r localContainerRegistryRepo) GetTags(credentials ContainerRegistryCredentials, image string) ([]string, error) {
	return []string{
		"latest",
	}, nil
}
