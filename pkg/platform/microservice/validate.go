package microservice

import (
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/thoas/go-funk"
)

func CheckIfIngressPathInUseInEnvironment(ingresses []platform.Ingress, environment string, ingressPath string) bool {
	pathExists := funk.Contains(ingresses, func(info platform.Ingress) bool {
		if info.Environment != environment {
			return false
		}
		return info.Path == ingressPath
	})

	return pathExists
}
