package types

import (
	"github.com/orientwalt/htdf/codec"
	codectypes "github.com/orientwalt/htdf/codec/types"
	"github.com/orientwalt/htdf/x/ibc/core/exported"
)

// RegisterInterfaces register the ibc interfaces submodule implementations to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*exported.ClientState)(nil),
		&ClientState{},
	)
}

var (
	// SubModuleCdc references the global x/ibc/light-clients/09-localhost module codec.
	// The actual codec used for serialization should be provided to x/ibc/light-clients/09-localhost and
	// defined at the application level.
	SubModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)