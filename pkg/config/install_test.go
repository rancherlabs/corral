package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstall(t *testing.T) {
	rootPath = t.TempDir()

	require.NoError(t, Install())
}
