package rawdatalog_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRawdatalog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rawdatalog Suite")
}
