package _package

import (
	"github.com/santhosh-tekuri/jsonschema/v5"
)

type VarSet map[string]string

type Schema struct {
	*jsonschema.Schema

	Sensitive   bool
	Optional    bool
	Description string
}
