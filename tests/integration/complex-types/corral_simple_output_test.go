package complex_types

import (
	"crypto/rand"
	"crypto/rsa"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/rancherlabs/corral/cmd"
	cmdconfig "github.com/rancherlabs/corral/cmd/config"
	cmdpackage "github.com/rancherlabs/corral/cmd/package"
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/rancherlabs/corral/pkg/corral"
	"github.com/rancherlabs/corral/pkg/vars"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"gotest.tools/v3/assert"
)

func TestSimpleOutput(t *testing.T) {
	config.InitializeRootPath(t.TempDir())
	t.Run("validate", func(t *testing.T) {
		validateCmd := cmdpackage.NewCommandValidate()
		validateCmd.SetArgs([]string{"testdata"})
		require.NoError(t, validateCmd.Execute())
	})
	configCmd := cmdconfig.NewCommandConfig()
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)
	require.NoError(t, priv.Validate())

	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	require.NoError(t, err)

	pubBytes := ssh.MarshalAuthorizedKey(pub)
	pubPath := filepath.Join(t.TempDir(), "id_rsa.pub")
	require.NoError(t, ioutil.WriteFile(pubPath, pubBytes, 0o600))

	configCmd.SetArgs([]string{"--user_id", "testuser", "--public_key", pubPath})
	require.NoError(t, configCmd.Execute())
	t.Cleanup(func() {
		deleteCmd := cmd.NewCommandDelete()
		deleteCmd.SetArgs([]string{"test-corral"})
		require.NoError(t, deleteCmd.Execute())
	})
	t.Run("create", func(t *testing.T) {
		createCmd := cmd.NewCommandCreate()
		createCmd.SetArgs([]string{"test-corral", "testdata"})
		require.NoError(t, createCmd.Execute())
		t.Run("variables", func(t *testing.T) {
			c, err := corral.Load(config.CorralPath("test-corral"))
			require.NoError(t, err)

			tests := []struct {
				name     string
				expected any
			}{
				{
					name:     "number",
					expected: 1,
				},
				{
					name:     "singlequotednumber",
					expected: 1,
				},
				{
					name:     "doublequotednumber",
					expected: 1,
				},
				{
					name:     "string",
					expected: "abc",
				},
				{
					name:     "singlequotedstring",
					expected: "abc",
				},
				{
					name:     "doublequotedstring",
					expected: "abc",
				},
				{
					name:     "array",
					expected: []any{1, 2, 3},
				},
				{
					name:     "singlequotedarray",
					expected: []any{1, 2, 3},
				},
				{
					name:     "doublequotedarray",
					expected: []any{1, 2, 3},
				},
				{
					name:     "object",
					expected: vars.VarSet{"a": 1, "b": 2, "c": "3", "d": []any{4, "5"}},
				},
				{
					name:     "singlequotedobject",
					expected: vars.VarSet{"a": 1, "b": 2, "c": "3", "d": []any{4, "5"}},
				},
				{
					name:     "doublequotedobject",
					expected: vars.VarSet{"a": 1, "b": 2, "c": "3", "d": []any{4, "5"}},
				},
				{
					name:     "string_output",
					expected: "a",
				},
				{
					name:     "number_output",
					expected: 1,
				},
				{
					name:     "array_output",
					expected: []any{1, 2, 3},
				},
				{
					name:     "object_output",
					expected: vars.VarSet{"a": 1, "b": 2, "c": "3", "d": []any{4, "5"}},
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					assert.DeepEqual(t, c.Vars[tt.name], tt.expected)
				})
			}
		})
	})
}
