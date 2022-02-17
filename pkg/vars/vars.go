package vars

import (
	"bytes"
	"encoding/json"
	"github.com/hashicorp/terraform-exec/tfexec"
	"strings"
)

const quote = byte('"')

// ToVar parses a var string and returns the key and value.
func ToVar(in string) (key, value string) {
	parts := strings.SplitN(in, "=", 2)
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}

// FromTerraformOutputMeta returns strings as they are and properly escapes json objects
func FromTerraformOutputMeta(in tfexec.OutputMeta) string {
	var buf bytes.Buffer
	var rval string
	raw, _ := in.Value.MarshalJSON()

	// if this is a json string get the raw value
	if bytes.HasPrefix(raw, []byte{quote}) {
		raw = bytes.TrimPrefix(raw, []byte{quote})
		raw = bytes.TrimSuffix(raw, []byte{quote})

		// attempt to compact the json and escape it
	} else if json.Compact(&buf, raw) == nil {
		for {
			c, err := buf.ReadByte()
			if err != nil {
				return rval
			}

			if c == quote {
				rval += `\`
			}

			rval += string(c)
		}
	}

	// default to returning the raw value
	return string(raw)
}
