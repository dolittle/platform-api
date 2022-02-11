package storage

import "github.com/dolittle/platform-api/pkg/platform"

func DefaultStudioConfig() platform.StudioConfig {
	return platform.StudioConfig{
		BuildOverwrite:       true,
		DisabledEnvironments: make([]string, 0),
		CanCreateApplication: true,
	}
}
