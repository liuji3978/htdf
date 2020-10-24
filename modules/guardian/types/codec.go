package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers the necessary module/guardian interfaces and concrete types
// on the provided Amino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*GuardianI)(nil), nil)
	cdc.RegisterConcrete(&MsgAddProfiler{}, "irishub/guardian/MsgAddProfiler", nil)
	cdc.RegisterConcrete(&MsgAddTrustee{}, "irishub/guardian/MsgAddTrustee", nil)
	cdc.RegisterConcrete(&MsgDeleteProfiler{}, "irishub/guardian/MsgDeleteProfiler", nil)
	cdc.RegisterConcrete(&MsgDeleteTrustee{}, "irishub/guardian/MsgDeleteTrustee", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddProfiler{},
		&MsgAddTrustee{},
		&MsgDeleteProfiler{},
		&MsgDeleteTrustee{},
	)

	registry.RegisterInterface(
		"irishub.guardian.GuardianI",
		(*GuardianI)(nil),
		&Guardian{},
	)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
