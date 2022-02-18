package _package

import (
	"github.com/santhosh-tekuri/jsonschema/v5"
)

type Schema struct {
	*jsonschema.Schema

	Sensitive   bool
	Optional    bool
	Description string
	Default     string
}
