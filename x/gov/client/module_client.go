package client

import (
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/orientwalt/htdf/client"
	"github.com/orientwalt/htdf/x/gov"
	govcli "github.com/orientwalt/htdf/x/gov/client/cli"
	hsgovcli "github.com/orientwalt/htdf/x/gov/client/cli"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	// Group gov queries under a subcommand
	govQueryCmd := &cobra.Command{
		Use:   gov.ModuleName,
		Short: "Querying commands for the governance module",
	}

	govQueryCmd.AddCommand(client.GetCommands(
		govcli.GetCmdQueryProposal(mc.storeKey, mc.cdc),
		govcli.GetCmdQueryProposals(mc.storeKey, mc.cdc),
		govcli.GetCmdQueryVote(mc.storeKey, mc.cdc),
		govcli.GetCmdQueryVotes(mc.storeKey, mc.cdc),
		govcli.GetCmdQueryParam(mc.storeKey, mc.cdc),
		govcli.GetCmdQueryParams(mc.storeKey, mc.cdc),
		govcli.GetCmdQueryProposer(mc.storeKey, mc.cdc),
		govcli.GetCmdQueryDeposit(mc.storeKey, mc.cdc),
		govcli.GetCmdQueryDeposits(mc.storeKey, mc.cdc),
		govcli.GetCmdQueryTally(mc.storeKey, mc.cdc))...)

	return govQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	govTxCmd := &cobra.Command{
		Use:   gov.ModuleName,
		Short: "Governance transactions subcommands",
	}

	govTxCmd.AddCommand(client.PostCommands(
		hsgovcli.GetCmdDeposit(mc.storeKey, mc.cdc),
		hsgovcli.GetCmdVote(mc.storeKey, mc.cdc),
		hsgovcli.GetCmdSubmitProposal(mc.cdc),
	)...)

	return govTxCmd
}
