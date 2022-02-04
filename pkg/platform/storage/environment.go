package storage

import (
	"github.com/thoas/go-funk"
)

func EnvironmentExists(environments []JSONEnvironment, environment string) bool {
	return funk.Contains(environments, func(item JSONEnvironment) bool {
		return item.Name == environment
	})
}
