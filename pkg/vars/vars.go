package vars

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
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
	raw, _ := in.Value.MarshalJSON()

	// if this is a json string get the raw value
	if bytes.HasPrefix(raw, []byte{quote}) {
		raw = bytes.TrimPrefix(raw, []byte{quote})
		raw = bytes.TrimSuffix(raw, []byte{quote})

		// attempt to compact the json and escape it
	} else if json.Compact(&buf, raw) == nil {
		return Escape(&buf)
	}

	// default to returning the raw value
	return string(raw)
}

func Escape(in io.ByteReader) (out string) {
	for {
		c, err := in.ReadByte()
		if err != nil {
			return
		}

		if c == quote {
			out += `\`
		}

		out += string(c)
	}
}
