package validate_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancherlabs/corral/pkg/config"
	_package "github.com/rancherlabs/corral/pkg/package"
)

var _ = Describe("Validate packages", func() {
	BeforeEach(func() {
		config.InitializeRootPath(GinkgoT().TempDir())
	})
	When("the package does not have a manifest", func() {
		It("should not be validated", func() {
			Expect(_package.Validate("./testdata/no_manifest")).Should(MatchError(os.ErrNotExist))
		})
	})
	When("the package does not have an overlay folder", func() {
		It("should not be validated", func() {
			Expect(_package.Validate("./testdata/no_overlay")).Should(MatchError(_package.ErrOverlayNotFound))
		})
	})
	When("the package uses a module that is not present", func() {
		It("should not be validated", func() {
			Expect(_package.Validate("./testdata/no_manifest")).Should(MatchError(os.ErrNotExist))
		})
	})
	When("the package is valid", func() {
		It("should be validated", func() {
			Expect(_package.Validate("./testdata/valid")).To(BeNil())
		})
	})
})
