package cmd

import (
	"encoding/json"
	"errors"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type OutputFormat string

const (
	OutputFormatTable OutputFormat = "table"
	OutputFormatJSON  OutputFormat = "json"
	OutputFormatYAML  OutputFormat = "yaml"
)

func (e *OutputFormat) String() string {
	return string(*e)
}

var ErrUnknownOutputFormat = errors.New(`must be one of "table", "json", or "yaml"`)

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *OutputFormat) Set(v string) error {
	switch v {
	case "table", "json", "yaml":
		*e = OutputFormat(v)
		return nil
	default:
		return ErrUnknownOutputFormat
	}
}

func (e *OutputFormat) Type() string {
	return ""
}

type OutputOptions struct {
	Key   string
	Value string
}

func Output[K comparable, V any](out map[K]V, output OutputFormat, opts OutputOptions) (string, error) {
	switch output {
	case OutputFormatTable:
		tbl := table.NewWriter()
		tbl.AppendHeader(table.Row{opts.Key, opts.Value})
		tbl.AppendSeparator()
		for k, v := range out {
			tbl.AppendRow(table.Row{k, v})
		}
		return tbl.Render(), nil
	case OutputFormatJSON:
		data, err := json.Marshal(&out)
		if err != nil {
			return "", err
		}
		return string(data), nil
	case OutputFormatYAML:
		data, err := yaml.Marshal(&out)
		if err != nil {
			logrus.Error(err)
			return "", err
		}
		return string(data[:len(data)-1]), nil // remove trailing newline
	}
	return "", ErrUnknownOutputFormat
}
