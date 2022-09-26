package vars_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancherlabs/corral/cmd/config/vars"
	"github.com/rancherlabs/corral/pkg/config"
)

var _ = Describe("Set", func() {
	When("a string is passed", func() {
		It("returns that string", func() {
			cfg := &config.Config{Vars: map[string]any{}}
			err := vars.CreateVar(cfg, "test", "test")
			Expect(err).To(BeNil())
			Expect(cfg.Vars["test"]).To(Equal("test"))
		})
	})
	When("a number is passed", func() {
		It("returns a number", func() {
			cfg := &config.Config{Vars: map[string]any{}}
			err := vars.CreateVar(cfg, "test", "1")
			Expect(err).To(BeNil())
			Expect(cfg.Vars["test"]).To(Equal(1.))
		})
	})
	When("a number is passed in quotes", func() {
		It("returns a string", func() {
			cfg := &config.Config{Vars: map[string]any{}}
			err := vars.CreateVar(cfg, "test", `"1"`)
			Expect(err).To(BeNil())
			Expect(cfg.Vars["test"]).To(Equal("1"))
		})
	})
	When("a json list is passed", func() {
		It("returns a slice", func() {
			cfg := &config.Config{Vars: map[string]any{}}
			err := vars.CreateVar(cfg, "test", "[1,2,3]")
			Expect(err).To(BeNil())
			Expect(cfg.Vars["test"]).To(Equal([]any{1., 2., 3.}))
		})
	})
	When("a json object is passed", func() {
		It("returns a map", func() {
			cfg := &config.Config{Vars: map[string]any{}}
			err := vars.CreateVar(cfg, "test", `{"a":"1","b":2,"c":[1,2,3]}`)
			Expect(err).To(BeNil())
			Expect(cfg.Vars["test"]).To(Equal(map[string]any{"a": "1", "b": 2., "c": []any{1., 2., 3.}}))
		})
	})
})
