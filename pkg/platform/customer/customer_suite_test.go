package customer_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCustomer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Customer Suite")
}
