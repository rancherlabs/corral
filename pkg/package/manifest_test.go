package _package_test

import (
	"embed"
	"testing"

	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/stretchr/testify/assert"
)

//go:embed tests
var _fs embed.FS

func TestLoadManifest(t *testing.T) {
	{ // valid
		res, err := _package.LoadManifest(_fs, "tests/valid.yaml")

		assert.NoError(t, err)

		assert.Equal(t, "valid", res.Name)
		assert.Equal(t, "0.1.0", res.Version)
		assert.Equal(t, "valid description", res.Description)

		assert.Len(t, res.Commands, 1)
		assert.Equal(t, "whoami", res.Commands[0].Command)

		assert.Len(t, res.Commands[0].NodePoolNames, 1)
		assert.Equal(t, "foo", res.Commands[0].NodePoolNames[0])

		assert.Len(t, res.VariableSchemas, 4)

		assert.NotNil(t, res.VariableSchemas["a"])
		assert.False(t, res.VariableSchemas["a"].Sensitive)
		assert.False(t, res.VariableSchemas["a"].Optional)
		assert.False(t, res.VariableSchemas["a"].ReadOnly)
		assert.Equal(t, []string{"string"}, res.VariableSchemas["a"].Types)

		assert.NotNil(t, res.VariableSchemas["b"])
		assert.False(t, res.VariableSchemas["b"].Sensitive)
		assert.False(t, res.VariableSchemas["b"].Optional)
		assert.True(t, res.VariableSchemas["b"].ReadOnly)
		assert.Equal(t, []string{"integer"}, res.VariableSchemas["b"].Types)

		assert.NotNil(t, res.VariableSchemas["c"])
		assert.True(t, res.VariableSchemas["c"].Sensitive)
		assert.False(t, res.VariableSchemas["c"].Optional)
		assert.False(t, res.VariableSchemas["c"].ReadOnly)
		assert.Equal(t, []string{"string"}, res.VariableSchemas["c"].Types)

		assert.NotNil(t, res.VariableSchemas["d"])
		assert.False(t, res.VariableSchemas["d"].Sensitive)
		assert.True(t, res.VariableSchemas["d"].Optional)
		assert.False(t, res.VariableSchemas["d"].ReadOnly)
		assert.Equal(t, []string{"boolean"}, res.VariableSchemas["d"].Types)
	}

	{ // bad schema
		_, err := _package.LoadManifest(_fs, "tests/bad-schema.yaml")

		assert.Error(t, err)
	}
}

func TestValidateVarSet(t *testing.T) {
	manifest, _ := _package.LoadManifest(_fs, "tests/valid.yaml")

	{ // valid
		vars := _package.VarSet{
			"a": "aval",
			"c": "cval",
			"d": "true",
		}

		res := manifest.ValidateVarSet(vars, true)

		assert.NoError(t, res)
	}

	{ // read only
		vars := _package.VarSet{
			"a": "aval",
			"b": "12",
			"c": "cval",
			"d": "true",
		}

		res := manifest.ValidateVarSet(vars, true)

		assert.Error(t, res)
	}

	{ // read only no write
		vars := _package.VarSet{
			"a": "aval",
			"b": "12",
			"c": "cval",
			"d": "true",
		}

		res := manifest.ValidateVarSet(vars, false)

		assert.NoError(t, res)
	}

	{ // optional
		vars := _package.VarSet{}

		res := manifest.ValidateVarSet(vars, true)

		assert.Error(t, res)
	}

	{ // schema
		vars := _package.VarSet{
			"a": "aval",
			"b": "five",
			"c": "cval",
		}

		res := manifest.ValidateVarSet(vars, true)

		assert.Error(t, res)
	}
}

func TestFilterVars(t *testing.T) {
	manifest, _ := _package.LoadManifest(_fs, "tests/valid.yaml")

	vars := _package.VarSet{
		"a": "",
		"b": "",
		"c": "",
		"d": "",
		"e": "",
	}

	res := manifest.FilterVars(vars)

	assert.Equal(t, res, _package.VarSet{"a": "", "b": "", "c": "", "d": ""})
}

func TestFilterSensitiveVars(t *testing.T) {
	manifest, _ := _package.LoadManifest(_fs, "tests/valid.yaml")

	vars := _package.VarSet{
		"a": "",
		"b": "",
		"c": "",
		"d": "",
		"e": "",
	}

	res := manifest.FilterSensitiveVars(vars)

	assert.Equal(t, res, _package.VarSet{"a": "", "b": "", "d": "", "e": ""})
}
