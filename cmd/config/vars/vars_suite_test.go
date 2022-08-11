package vars_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVars(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vars Suite")
}
