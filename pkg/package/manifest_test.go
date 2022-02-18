package _package_test

import (
	"embed"
	"github.com/rancherlabs/corral/pkg/vars"
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
		assert.Equal(t, "valid description", res.Description)

		assert.NotNil(t, res.Annotations)
		assert.Equal(t, res.Annotations["foo"], "bar")
		assert.Equal(t, res.Annotations["baz"], "1")

		assert.Len(t, res.Commands, 1)
		assert.Equal(t, "whoami", res.Commands[0].Command)

		assert.Len(t, res.Commands[0].NodePoolNames, 1)
		assert.Equal(t, "foo", res.Commands[0].NodePoolNames[0])

		assert.Len(t, res.VariableSchemas, 5)

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
		vs := vars.VarSet{
			"a": "aval",
			"c": "cval",
			"d": "true",
		}

		res := manifest.ValidateVarSet(vs, true)

		assert.NoError(t, res)
	}

	{ // read only
		vs := vars.VarSet{
			"a": "aval",
			"b": "12",
			"c": "cval",
			"d": "true",
		}

		res := manifest.ValidateVarSet(vs, true)

		assert.Error(t, res)
	}

	{ // read only no write
		vs := vars.VarSet{
			"a": "aval",
			"b": "12",
			"c": "cval",
			"d": "true",
		}

		res := manifest.ValidateVarSet(vs, false)

		assert.NoError(t, res)
	}

	{ // optional
		vs := vars.VarSet{}

		res := manifest.ValidateVarSet(vs, true)

		assert.Error(t, res)
	}

	{ // schema
		vs := vars.VarSet{
			"a": "aval",
			"b": "five",
			"c": "cval",
		}

		res := manifest.ValidateVarSet(vs, true)

		assert.Error(t, res)
	}
}

func TestFilterVars(t *testing.T) {
	manifest, _ := _package.LoadManifest(_fs, "tests/valid.yaml")

	vs := vars.VarSet{
		"a": "",
		"b": "",
		"c": "",
		"d": "",
		"e": "",
		"f": "",
	}

	res := manifest.FilterVars(vs)

	assert.Equal(t, res, vars.VarSet{"a": "", "b": "", "c": "", "d": "", "e": ""})
}

func TestFilterSensitiveVars(t *testing.T) {
	manifest, _ := _package.LoadManifest(_fs, "tests/valid.yaml")

	vs := vars.VarSet{
		"a": "",
		"b": "",
		"c": "",
		"d": "",
		"e": "",
	}

	res := manifest.FilterSensitiveVars(vs)

	assert.Equal(t, res, vars.VarSet{"a": "", "b": "", "d": "", "e": ""})
}
