package configFiles_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConfigFiles(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ConfigFiles Suite")
}
