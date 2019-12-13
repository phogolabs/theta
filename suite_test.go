package theta_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTheta(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Theta Suite")
}
