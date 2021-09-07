package microservice_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMicroservice(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Microservice Suite")
}
