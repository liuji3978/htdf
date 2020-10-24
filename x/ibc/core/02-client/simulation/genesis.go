package simulation

import (
	"math/rand"

	simtypes "github.com/orientwalt/htdf/types/simulation"
	"github.com/orientwalt/htdf/x/ibc/core/02-client/types"
)

// GenClientGenesis returns the default client genesis state.
func GenClientGenesis(_ *rand.Rand, _ []simtypes.Account) types.GenesisState {
	return types.DefaultGenesisState()
}
