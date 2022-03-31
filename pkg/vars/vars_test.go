package vars

import (
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
)

func TestToVar(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name      string
		args      args
		wantKey   string
		wantValue string
	}{
		{
			name:      "valid",
			args:      struct{ in string }{in: "foo=bar"},
			wantKey:   "foo",
			wantValue: "bar",
		},
		{
			name:      "empty value",
			args:      struct{ in string }{in: "foo="},
			wantKey:   "foo",
			wantValue: "",
		},
		{
			name:      "missing value",
			args:      struct{ in string }{in: "foo"},
			wantKey:   "",
			wantValue: "",
		},
		{
			name:      "empty",
			args:      struct{ in string }{in: ""},
			wantKey:   "",
			wantValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotValue := ToVar(tt.args.in)
			if gotKey != tt.wantKey {
				t.Errorf("ToVar() gotKey = %v, want %v", gotKey, tt.wantKey)
			}
			if gotValue != tt.wantValue {
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
		want string
	}{
		{
			name: "int",
			args: struct{ in tfexec.OutputMeta }{in: tfexec.OutputMeta{
				Value: []byte(`1`),
			}},
			want: `1`,
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
				Value: []byte(`["a", "b", "c"]`),
			}},
			want: `[\"a\",\"b\",\"c\"]`,
		},
		{
			name: "object",
			args: struct{ in tfexec.OutputMeta }{in: tfexec.OutputMeta{
				Value: []byte(`{"a": "b"}`),
			}},
			want: `{\"a\":\"b\"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromTerraformOutputMeta(tt.args.in); got != tt.want {
				t.Errorf("FromTerraformOutputMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}
