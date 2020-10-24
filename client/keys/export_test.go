package keys

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/orientwalt/htdf/client"
	"github.com/orientwalt/htdf/crypto/hd"
	"github.com/orientwalt/htdf/testutil"

	"github.com/orientwalt/htdf/client/flags"
	"github.com/orientwalt/htdf/crypto/keyring"
	sdk "github.com/orientwalt/htdf/types"
)

func Test_runExportCmd(t *testing.T) {
	cmd := ExportKeyCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

	// Now add a temporary keybase
	kbHome := t.TempDir()

	// create a key
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn)
	require.NoError(t, err)
	t.Cleanup(func() {
		kb.Delete("keyname1") // nolint:errcheck
	})

	path := sdk.GetConfig().GetFullFundraiserPath()
	_, err = kb.NewAccount("keyname1", testutil.TestMnemonic, "", path, hd.Secp256k1)
	require.NoError(t, err)

	// Now enter password
	mockIn.Reset("123456789\n123456789\n")
	cmd.SetArgs([]string{
		"keyname1",
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})

	clientCtx := client.Context{}.WithKeyring(kb)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	require.NoError(t, cmd.ExecuteContext(ctx))
}