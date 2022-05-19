package m3connector_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestM3connector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "M3connector Suite")
}
