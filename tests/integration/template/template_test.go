package template_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancherlabs/corral/pkg/config"
	_package "github.com/rancherlabs/corral/pkg/package"
)

var _ = Describe("Template and package tests", func() {
	BeforeEach(func() {
		config.InitializeRootPath(GinkgoT().TempDir())
	})
	When("a local package does not exist", func() {
		It("should report an error", func() {
			Expect(_package.Template("test", "", "./doesnotexist")).Should(MatchError(os.ErrNotExist))
		})
	})
	When("a remote package does not exist", func() {
		It("should report an error", func() {
			Expect(_package.Template("test", "", "doesnotexist:latest")).ToNot(BeNil())
		})
	})
	When("the templates are valid", func() {
		It("should create successfully", func() {
			Expect(_package.Template("test", "", "testdata/template/template1", "testdata/template/template2")).To(BeNil())
			pkg, err := _package.LoadPackage("test")
			Expect(err).To(BeNil())
			Expect(pkg.Description).To(Equal("template1 input\ntemplate2 input"))
		})
	})
	When("description is not empty", func() {
		It("should utilize the description", func() {
			Expect(_package.Template("test", "test description", "testdata/template/template1", "testdata/template/template2")).To(BeNil())
			pkg, err := _package.LoadPackage("test")
			Expect(err).To(BeNil())
			Expect(pkg.Description).To(Equal("test description"))
		})
	})
})
