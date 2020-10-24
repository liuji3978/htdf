package testutil

import (
	"fmt"
	"strings"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/orientwalt/htdf/client"
	"github.com/orientwalt/htdf/client/flags"
	"github.com/orientwalt/htdf/crypto/keyring"
	"github.com/orientwalt/htdf/testutil"
	clitestutil "github.com/orientwalt/htdf/testutil/cli"
	"github.com/orientwalt/htdf/x/auth/client/cli"
)

func TxSignExec(clientCtx client.Context, from fmt.Stringer, filename string, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--from=%s", from.String()),
		fmt.Sprintf("--%s=%s", flags.FlagHome, strings.Replace(clientCtx.HomeDir, "simd", "simcli", 1)),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID),
		filename,
	}

	args = append(args, extraArgs...)

	cmd := cli.GetSignCommand()
	tmcli.PrepareBaseCmd(cmd, "", "")

	return clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
}

func TxBroadcastExec(clientCtx client.Context, filename string, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{
		filename,
	}

	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.GetBroadcastCommand(), args)
}

func TxEncodeExec(clientCtx client.Context, filename string, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		filename,
	}

	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.GetEncodeCommand(), args)
}

func TxValidateSignaturesExec(clientCtx client.Context, filename string) (testutil.BufferWriter, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID),
		filename,
	}

	return clitestutil.ExecTestCLICmd(clientCtx, cli.GetValidateSignaturesCommand(), args)
}

func TxMultiSignExec(clientCtx client.Context, from string, filename string, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID),
		filename,
		from,
	}

	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.GetMultiSignCommand(), args)
}

func TxSignBatchExec(clientCtx client.Context, from fmt.Stringer, filename string, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--from=%s", from.String()),
		filename,
	}

	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.GetSignBatchCommand(), args)
}

func TxDecodeExec(clientCtx client.Context, encodedTx string, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		encodedTx,
	}

	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.GetDecodeCommand(), args)
}

// DONTCOVER