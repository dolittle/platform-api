package storage

import (
	"github.com/thoas/go-funk"
)

func EnvironmentExists(environments []JSONEnvironment2, environment string) bool {
	return funk.Contains(environments, func(item JSONEnvironment2) bool {
		return item.Name == environment
	})
}
