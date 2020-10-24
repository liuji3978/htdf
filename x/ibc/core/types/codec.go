package types

import (
	codectypes "github.com/orientwalt/htdf/codec/types"
	clienttypes "github.com/orientwalt/htdf/x/ibc/core/02-client/types"
	connectiontypes "github.com/orientwalt/htdf/x/ibc/core/03-connection/types"
	channeltypes "github.com/orientwalt/htdf/x/ibc/core/04-channel/types"
	commitmenttypes "github.com/orientwalt/htdf/x/ibc/core/23-commitment/types"
	solomachinetypes "github.com/orientwalt/htdf/x/ibc/light-clients/06-solomachine/types"
	ibctmtypes "github.com/orientwalt/htdf/x/ibc/light-clients/07-tendermint/types"
	localhosttypes "github.com/orientwalt/htdf/x/ibc/light-clients/09-localhost/types"
)

// RegisterInterfaces registers x/ibc interfaces into protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	clienttypes.RegisterInterfaces(registry)
	connectiontypes.RegisterInterfaces(registry)
	channeltypes.RegisterInterfaces(registry)
	solomachinetypes.RegisterInterfaces(registry)
	ibctmtypes.RegisterInterfaces(registry)
	localhosttypes.RegisterInterfaces(registry)
	commitmenttypes.RegisterInterfaces(registry)
}
