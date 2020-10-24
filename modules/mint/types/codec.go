package types

import (
	"github.com/orientwalt/htdf/codec"
	cryptocodec "github.com/orientwalt/htdf/crypto/codec"
)

var (
	amino = codec.NewLegacyAmino()

	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
