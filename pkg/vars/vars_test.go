package vars

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/stretchr/testify/assert"
)

func TestToVar(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name      string
		args      args
		wantKey   string
		wantValue any
	}{
		{
			name:      "valid",
			args:      struct{ in string }{in: "foo=bar"},
			wantKey:   "foo",
			wantValue: "bar",
		},
		{
			name:      "valid ip",
			args:      struct{ in string }{in: `foo=127.0.0.1`},
			wantKey:   "foo",
			wantValue: "127.0.0.1",
		},
		{
			name:      "valid int",
			args:      struct{ in string }{in: "foo=1"},
			wantKey:   "foo",
			wantValue: 1.,
		},
		{
			name:      "json map",
			args:      struct{ in string }{in: `foo={"test":1}`},
			wantKey:   "foo",
			wantValue: map[string]any{"test": 1.},
		},
		{
			name:      "json slice",
			args:      struct{ in string }{in: `foo=[1, 2, 3]`},
			wantKey:   "foo",
			wantValue: []any{1., 2., 3.},
		},
		{
			name:      "json object",
			args:      struct{ in string }{in: `foo={"test":{"a":1,"b":"2"}}`},
			wantKey:   "foo",
			wantValue: map[string]any{"test": map[string]any{"a": 1., "b": "2"}},
		},
		{
			name:      "empty value",
			args:      struct{ in string }{in: "foo="},
			wantKey:   "foo",
			wantValue: nil,
		},
		{
			name:      "missing value",
			args:      struct{ in string }{in: "foo"},
			wantKey:   "foo",
			wantValue: nil,
		},
		{
			name:      "empty",
			args:      struct{ in string }{in: ""},
			wantKey:   "",
			wantValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotValue, err := ToVar(tt.args.in)
			assert.NoError(t, err)
			if gotKey != tt.wantKey {
				t.Errorf("ToVar() gotKey = %v, want %v", gotKey, tt.wantKey)
			}
			if !Equal(gotValue, tt.wantValue) {
				t.Errorf("ToVar() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestFromTerraformOutputMeta(t *testing.T) {
	type args struct {
		in tfexec.OutputMeta
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "int",
			args: struct{ in tfexec.OutputMeta }{in: tfexec.OutputMeta{
				Value: []byte(`1`),
			}},
			want: 1.,
		},
		{
			name: "string",
			args: struct{ in tfexec.OutputMeta }{in: tfexec.OutputMeta{
				Value: []byte(`"foo"`),
			}},
			want: `foo`,
		},
		{
			name: "array",
			args: struct{ in tfexec.OutputMeta }{in: tfexec.OutputMeta{
				Value: []byte(`["a","b","c"]`),
			}},
			want: []string{"a", "b", "c"},
		},
		{
			name: "object",
			args: struct{ in tfexec.OutputMeta }{in: tfexec.OutputMeta{
				Value: []byte(`{"a":"b"}`),
			}},
			want: map[string]any{"a": "b"},
		},
		{
			name: "nested object",
			args: struct{ in tfexec.OutputMeta }{in: tfexec.OutputMeta{
				Value: []byte(`{"a":{"b":2}}`),
			}},
			want: map[string]any{
				"a": map[string]any{
					"b": 2.,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := FromTerraformOutputMeta(tt.args.in); err != nil {
				t.Error(err)
			} else {
				if !Equal(got, tt.want) {
					t.Errorf("FromTerraformOutputMeta() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Equal(got, wanted any) bool {
	if reflect.TypeOf(got) == reflect.TypeOf(wanted) {
		return reflect.DeepEqual(got, wanted)
	}

	if got == nil || wanted == nil {
		return got == wanted
	}

	kind := reflect.TypeOf(got).Kind()

	if kind != reflect.TypeOf(wanted).Kind() {
		return false
	}

	gotValue := reflect.ValueOf(got)
	wantedValue := reflect.ValueOf(wanted)

	switch kind {
	case reflect.Slice:
		if gotValue.Len() != wantedValue.Len() {
			return false
		}
		for i := 0; i < gotValue.Len(); i++ {
			if !Equal(gotValue.Index(i).Interface(), wantedValue.Index(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.Map:
		if gotValue.Len() != wantedValue.Len() {
			return false
		}
		for _, k := range gotValue.MapKeys() {
			val1 := gotValue.MapIndex(k)
			val2 := wantedValue.MapIndex(k)
			if !val1.IsValid() || !val2.IsValid() || !Equal(val1.Interface(), val2.Interface()) {
				return false
			}
		}
		return true
	}

	return false
}
