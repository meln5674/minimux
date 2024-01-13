package minimux_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMinimux(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Minimux Suite")
}
