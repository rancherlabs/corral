package shell

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/rancherlabs/corral/pkg/vars"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
)

func TestVarsToEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    vars.VarSet
		expected []string
	}{
		{
			name: "int",
			input: map[string]any{
				"int": 1,
			},
			expected: []string{
				"export CORRAL_int='1'",
			},
		},
		{
			name: "string",
			input: map[string]any{
				"string": "test",
			},
			expected: []string{
				"export CORRAL_string='test'",
			},
		},
		{
			name: "empty array",
			input: map[string]any{
				"array": []string{},
			},
			expected: []string{
				`export CORRAL_array='[]'`,
			},
		},
		{
			name: "array of strings",
			input: map[string]any{
				"array": []string{"a", "b", "c"},
			},
			expected: []string{
				`export CORRAL_array='["a","b","c"]'`,
			},
		},
		{
			name: "array of numbers",
			input: map[string]any{
				"array": []int{1, 2, 3},
			},
			expected: []string{
				`export CORRAL_array='[1,2,3]'`,
			},
		},
		{
			name: "empty object",
			input: map[string]any{
				"object": map[string]any{},
			},
			expected: []string{
				`export CORRAL_object='{}'`,
			},
		},
		{
			name: "object",
			input: map[string]any{
				"object": map[string]any{
					"1": "a",
					"2": 2,
					"3": []any{4.1, "5"},
				},
			},
			expected: []string{
				`export CORRAL_object='{"1":"a","2":2,"3":[4.1,"5"]}'`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := varsToEnvVars(tt.input)
			require.NoError(t, err)

			assert.DeepEqual(t, actual, tt.expected)
		})
	}
}

func TestConsumeStdout(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name:     "unqouted string",
			input:    "a",
			expected: "a",
		},
		{
			name:     "single qouted string",
			input:    "'a'",
			expected: "'a'",
		},
		{
			name:     "double qouted string",
			input:    `"a"`,
			expected: "a",
		},
		{
			name:     "unqouted number",
			input:    "1",
			expected: 1.,
		},
		{
			name:     "single qouted number",
			input:    "'1'",
			expected: "'1'",
		},
		{
			name:     "double qouted number",
			input:    `"1"`,
			expected: "1",
		},
		{
			name:     "homogeneously typed array",
			input:    `[1,2,3]`,
			expected: []any{1., 2., 3.},
		},
		{
			name:     "variously typed array",
			input:    `["1",2,3.1]`,
			expected: []any{"1", 2., 3.1},
		},
		{
			name:     "object",
			input:    `{"a":1,"b":"2","c":["1",2.0]}`,
			expected: map[string]any{"a": 1., "b": "2", "c": []any{"1", 2.}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			s := Shell{
				Vars: map[string]any{},
			}
			b.WriteString(fmt.Sprintf("corral_set test=%s\n", tt.input))
			s.consumeStdout(&b)
			assert.DeepEqual(t, s.Vars["test"], tt.expected)
		})
	}
}
