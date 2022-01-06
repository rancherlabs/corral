package _package

import (
	_ "embed"
	"fmt"

	"bytes"
	"encoding/json"
	"io"
	"io/fs"

	"github.com/pkg/errors"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v2"
)

type Command struct {
	NodePoolNames []string `yaml:"node_pools"`
	Command       string   `yaml:"command"`
}

type Manifest struct {
	Name            string            `yaml:"name"`
	Version         string            `yaml:"version"`
	Description     string            `yaml:"description"`
	Commands        []Command         `yaml:"commands"`
	VariableSchemas map[string]Schema `yaml:"-"`
}

//go:embed package-manifest.schema.json
var manifestSchemaBytes []byte

var manifestSchema *jsonschema.Schema
var schemaCompiler *jsonschema.Compiler

func init() {
	schemaCompiler = jsonschema.NewCompiler()
	schemaCompiler.Draft = jsonschema.Draft7

	_ = schemaCompiler.AddResource("manifest", bytes.NewReader(manifestSchemaBytes))
	manifestSchema = schemaCompiler.MustCompile("manifest")
	manifestSchema.Location = "Package Manifest"
}

// LoadManifest reads a manifest file and validates it is a valid manifest.
func LoadManifest(_fs fs.FS, path string) (Manifest, error) {
	var manifest Manifest

	f, err := _fs.Open(path)
	if err != nil {
		return manifest, err
	}
	defer func(f fs.File) { _ = f.Close() }(f)

	buf, err := io.ReadAll(f)
	if err != nil {
		return manifest, err
	}
	_ = f.Close()

	err = validateManifest(buf)
	if err != nil {
		return manifest, err
	}

	err = yaml.Unmarshal(buf, &manifest)
	if err != nil {
		return manifest, err
	}

	manifest.VariableSchemas = parseVariableSchema(buf)

	return manifest, err
}

// ValidateVarSet returns an error if the var set does not match the manifest variable schemas.
func (m *Manifest) ValidateVarSet(vars VarSet, write bool) error {
	for k, schema := range m.VariableSchemas {
		var parsedValue interface{}

		if _, ok := vars[k]; schema.ReadOnly && write && ok {
			return fmt.Errorf("[%s] may not be set", k)
		}

		if _, ok := vars[k]; !schema.Optional && !schema.ReadOnly && !ok {
			return fmt.Errorf("[%s] is required", k)
		}

		if err := json.Unmarshal([]byte(vars[k]), &parsedValue); err != nil {
			parsedValue = vars[k]
		}

		if err := schema.Validate(parsedValue); err != nil && parsedValue != "" {
			return errors.Wrap(err, k+":")
		}
	}

	return nil
}

// FilterVars returns the given VarSet without any variables not defined in the manifest
func (m *Manifest) FilterVars(vars VarSet) VarSet {
	rval := VarSet{}

	for k, v := range vars {
		if _, ok := m.VariableSchemas[k]; !ok {
			continue
		}

		rval[k] = v
	}

	return rval
}

// FilterSensitiveVars returns the given VarSet without any variables marked as sensitive in the manifest
func (m *Manifest) FilterSensitiveVars(vars VarSet) VarSet {
	rval := VarSet{}

	for k, v := range vars {
		schema := m.VariableSchemas[k]

		if schema.Sensitive {
			continue
		}

		rval[k] = v
	}

	return rval
}

// validateManifest returns an error of the manifest violates any rules defined in the package-manifest.schema.json
func validateManifest(manifest []byte) error {
	var yml interface{}

	_ = yaml.Unmarshal(manifest, &yml)

	j := toStringKeys(yml)

	return manifestSchema.Validate(j)
}

// toStringKeys converts any map[interface{}]interface{} to map[string]interface{} recursively.
func toStringKeys(val interface{}) interface{} {
	switch val := val.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, v := range val {
			k := k.(string)
			m[k] = toStringKeys(v)
		}
		return m
	case []interface{}:
		var l = make([]interface{}, len(val))
		for i, v := range val {
			l[i] = toStringKeys(v)
		}
		return l
	default:
		return val
	}
}

// parseVariableSchema accepts a manifest yaml as bytes and parses the schema map. Assumes manifest is a valid manifest
// and does not have any error handling!
func parseVariableSchema(manifest []byte) map[string]Schema {
	vars := struct {
		VariableSchemas map[string]interface{} `yaml:"variables,omitempty"`
	}{
		VariableSchemas: map[string]interface{}{},
	}
	_ = yaml.Unmarshal(manifest, &vars)

	rval := map[string]Schema{}
	for k, v := range vars.VariableSchemas {
		var buf bytes.Buffer
		var schema Schema

		_ = json.NewEncoder(&buf).Encode(toStringKeys(v))
		_ = schemaCompiler.AddResource(k, &buf)
		schema.Schema = schemaCompiler.MustCompile(k)
		schema.Location = k

		if vv, ok := v.(map[interface{}]interface{}); ok {
			if val, okk := vv["sensitive"].(bool); okk && val {
				schema.Sensitive = true
			}

			if val, okk := vv["optional"].(bool); okk && val {
				schema.Optional = true
			}

			if val, okk := vv["readOnly"].(bool); okk && val {
				schema.ReadOnly = true
			}

			if description, okk := vv["description"].(string); okk {
				schema.Description = description
			}
		}

		rval[k] = schema
	}

	return rval
}
