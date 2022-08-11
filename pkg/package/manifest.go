package _package

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"

	"github.com/pkg/errors"
	"github.com/rancherlabs/corral/pkg/vars"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

type Command struct {
	// shell fields
	Command       string   `yaml:"command,omitempty"`
	NodePoolNames []string `yaml:"node_pools,omitempty"`
	Parallel      *bool    `yaml:"parallel,omitempty"`

	// terraform module fields
	Module      string `yaml:"module,omitempty"`
	SkipCleanup bool   `yaml:"skip_cleanup,omitempty"`
}

type VariableSchemas map[string]Schema

type Manifest struct {
	Name            string            `yaml:"name"`
	Annotations     map[string]string `yaml:"annotations,omitempty"`
	Description     string            `yaml:"description,omitempty"`
	Commands        []Command         `yaml:"commands"`
	Overlay         map[string]string `yaml:"overlay,omitempty"`
	VariableSchemas VariableSchemas   `yaml:"variables,omitempty"`
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

	err = ValidateManifest(buf)
	if err != nil {
		return manifest, err
	}

	err = yaml.Unmarshal(buf, &manifest)
	if err != nil {
		return manifest, err
	}

	if manifest.Annotations == nil {
		manifest.Annotations = map[string]string{}
	}

	return manifest, err
}

func (m *Manifest) ApplyDefaultVars(vs vars.VarSet) error {
	for k, schema := range m.VariableSchemas {
		if _, ok := vs[k]; !ok {
			if schema.Default == nil {
				continue
			}
			value := schema.Default
			vs[k] = value
		}
	}

	return nil
}

// ValidateDefaults returns an error if the var set does not match the manifest variable schemas.
func (m *Manifest) ValidateDefaults() error {
	for _, schema := range m.VariableSchemas {
		if schema.Default != nil {
			err := schema.Validate(schema.Default)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ValidateVarSet returns an error if the var set does not match the manifest variable schemas.
func (m *Manifest) ValidateVarSet(vs vars.VarSet, write bool) error {
	for k, schema := range m.VariableSchemas {

		if _, ok := vs[k]; schema.ReadOnly && write && ok {
			return fmt.Errorf("[%s] may not be set", k)
		}

		if _, ok := vs[k]; !schema.Optional && !schema.ReadOnly && schema.Default == nil && !ok {
			return fmt.Errorf("[%s] is required", k)
		}

		if err := schema.Validate(vs[k]); err != nil && vs[k] != nil {
			return errors.Wrap(err, k+":")
		}
	}

	return nil
}

// FilterVars returns the given VarSet without any variables not defined in the manifest
func (m *Manifest) FilterVars(vs vars.VarSet) vars.VarSet {
	rval := vars.VarSet{}

	for k, v := range vs {
		if _, ok := m.VariableSchemas[k]; !ok {
			continue
		}

		rval[k] = v
	}

	return rval
}

// FilterSensitiveVars returns the given VarSet without any variables marked as sensitive in the manifest
func (m *Manifest) FilterSensitiveVars(vs vars.VarSet) vars.VarSet {
	rval := vars.VarSet{}

	for k, v := range vs {
		schema := m.VariableSchemas[k]

		if schema.Sensitive {
			continue
		}

		rval[k] = v
	}

	return rval
}

func (m *Manifest) GetAnnotation(key string) string {
	if m.Annotations != nil {
		return m.Annotations[key]
	}

	return ""
}

// ValidateManifest returns an error of the manifest violates any rules defined in the package-manifest.schema.json
func ValidateManifest(manifest []byte) error {
	var yml any

	_ = yaml.Unmarshal(manifest, &yml)

	j := toStringKeys(yml)

	return manifestSchema.Validate(j)
}

// toStringKeys converts any map[any]any to map[string]any recursively.
func toStringKeys(val any) any {
	switch val := val.(type) {
	case map[any]any:
		m := make(map[string]any)
		for k, v := range val {
			k := k.(string)
			m[k] = toStringKeys(v)
		}
		return m
	case []any:
		var l = make([]any, len(val))
		for i, v := range val {
			l[i] = toStringKeys(v)
		}
		return l
	default:
		return val
	}
}

func (s *VariableSchemas) UnmarshalYAML(unmarshal func(any) error) error {
	*s = VariableSchemas{}
	_vars := struct {
		VariableSchemas map[string]any `yaml:"variables,inline,omitempty"`
	}{
		VariableSchemas: map[string]any{},
	}
	err := unmarshal(&_vars)
	if err != nil {
		return err
	}

	for k, v := range _vars.VariableSchemas {
		var buf bytes.Buffer
		var schema Schema

		_ = json.NewEncoder(&buf).Encode(toStringKeys(v))
		_ = schemaCompiler.AddResource(k, &buf)
		schema.Schema = schemaCompiler.MustCompile(k)
		schema.Location = k

		if vv, ok := v.(map[string]any); ok {
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

			if dflt, okk := vv["default"]; okk {
				schema.Default = dflt
			}
		}

		(*s)[k] = schema
	}

	return nil
}
