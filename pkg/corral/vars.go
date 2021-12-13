package corral

import (
	"strings"
)

func ToVar(in string) (key, value string) {
	parts := strings.SplitN(in, "=", 2)
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}
