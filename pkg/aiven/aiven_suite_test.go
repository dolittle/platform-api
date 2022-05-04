package aiven

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAiven(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aiven Suite")
}
