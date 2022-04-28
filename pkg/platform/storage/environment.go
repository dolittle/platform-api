package storage

import (
	"github.com/thoas/go-funk"
)

func EnvironmentExists(environments []JSONEnvironment, environment string) bool {
	return funk.Contains(environments, func(item JSONEnvironment) bool {
		return item.Name == environment
	})
}

func GetEnvironment(environments []JSONEnvironment, environment string) (JSONEnvironment, error) {
	found := funk.Find(environments, func(item JSONEnvironment) bool {
		return item.Name == environment
	})

	if found == nil {
		return JSONEnvironment{}, ErrNotFound
	}
	return found.(JSONEnvironment), nil
}
