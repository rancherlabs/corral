package vars_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancherlabs/corral/cmd/config/vars"
	pkgcmd "github.com/rancherlabs/corral/pkg/cmd"
	"github.com/rancherlabs/corral/pkg/config"
)

var _ = Describe("Vars", func() {
	When("only one variable is passed", func() {
		When("the output format is table", func() {
			It("returns only one variable", func() {
				v, err := vars.ListVars(config.Config{Vars: map[string]any{"test": "test1"}}, pkgcmd.OutputFormatTable, "test")
				Expect(err).To(BeNil())
				Expect(v).To(Equal("test1"))
			})
		})
		When("the output format is json", func() {
			It("returns only one variable", func() {
				v, err := vars.ListVars(config.Config{Vars: map[string]any{"test": "test1"}}, pkgcmd.OutputFormatJSON, "test")
				Expect(err).To(BeNil())
				Expect(v).To(Equal("test1"))
			})
		})
		When("the output format is yaml", func() {
			It("returns only one variable", func() {
				v, err := vars.ListVars(config.Config{Vars: map[string]any{"test": "test1"}}, pkgcmd.OutputFormatYAML, "test")
				Expect(err).To(BeNil())
				Expect(v).To(Equal("test1"))
			})
		})
	})
	When("an unsupported output format is passed", func() {
		It("returns an error", func() {
			_, err := vars.ListVars(config.Config{Vars: map[string]any{}}, "a")
			Expect(err).Should(MatchError(pkgcmd.ErrUnknownOutputFormat))
		})
	})
	When("no variables are passed", func() {
		When("the output format is table", func() {
			It("has the expected output", func() {
				v, err := vars.ListVars(config.Config{Vars: map[string]any{"test": "test1"}}, pkgcmd.OutputFormatTable)
				Expect(err).To(BeNil())
				Expect(v).To(Equal("+------+-------+\n| NAME | VALUE |\n+------+-------+\n| test | test1 |\n+------+-------+"))
			})
		})
		When("the output format is json", func() {
			It("has the expected output", func() {
				v, err := vars.ListVars(config.Config{Vars: map[string]any{"test": "test1"}}, pkgcmd.OutputFormatJSON)
				Expect(err).To(BeNil())
				Expect(v).To(Equal(`{"test":"test1"}`))
			})
		})
		When("the output format is yaml", func() {
			It("has the expected output", func() {
				v, err := vars.ListVars(config.Config{Vars: map[string]any{"test": "test1"}}, pkgcmd.OutputFormatYAML)
				Expect(err).To(BeNil())
				Expect(v).To(Equal("test: test1"))
			})
		})
	})
})
