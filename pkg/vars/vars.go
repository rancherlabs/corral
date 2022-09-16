package vars

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/pkg/errors"
)

const quote = byte('"')

// if it doesn't start with a { or [, or if it is entirely numbers
var regex = regexp.MustCompile(`(^([\[{])+)|(^\d+$)`)

// ToVar parses a var string and returns the key and value.
func ToVar(in string) (key string, value any, err error) {
	parts := strings.SplitN(in, "=", 2)
	if len(parts) != 2 || len(parts[1]) == 0 {
		return parts[0], nil, nil
	}

	key = parts[0]

	value, err = FromJson(parts[1])
	if err != nil {
		return "", nil, err
	}

	return
}

func FromJson(in string) (value any, err error) {
	// raw string values need to be quoted, so any value that doesn't start with a { or [, or is entirely numbers is
	// assumed a string
	// todo(jhyde): allow specifying types when using corral_set if necessary (i.e. variable is known, or e.g. "1")
	if !regex.Match([]byte(in)) && !strings.HasPrefix(in, `"`) && !strings.HasSuffix(in, `"`) {
		in = fmt.Sprintf(`"%s"`, in)
	}

	err = json.Unmarshal([]byte(in), &value)
	if err != nil {
		return nil, errors.Wrapf(err, `unmarshaling "%s"`, in)
	}

	return
}

// FromTerraformOutputMeta returns strings as they are and properly escapes json objects
func FromTerraformOutputMeta(in tfexec.OutputMeta) (any, error) {
	raw, _ := in.Value.MarshalJSON()
	var m any

	err := json.Unmarshal(raw, &m)
	if err != nil {
		return nil, err
	}

	return m, err
}

func Escape(buf *bytes.Buffer) (out string) {
	c, err := buf.ReadByte()
	if err != nil {
		return
	}

	for {
		if c == quote {
			out += `\`
		}

		out += string(c)

		c, err = buf.ReadByte()
		if err != nil {
			return
		}
	}
}
