package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/orientwalt/htdf/codec"
	codectypes "github.com/orientwalt/htdf/codec/types"
	cryptocodec "github.com/orientwalt/htdf/crypto/codec"
	"github.com/orientwalt/htdf/crypto/keys/secp256k1"
	cryptotypes "github.com/orientwalt/htdf/crypto/types"
	"github.com/orientwalt/htdf/testutil/testdata"
	sdk "github.com/orientwalt/htdf/types"
	host "github.com/orientwalt/htdf/x/ibc/core/24-host"
	"github.com/orientwalt/htdf/x/ibc/core/exported"
	"github.com/orientwalt/htdf/x/ibc/light-clients/06-solomachine/types"
	ibctesting "github.com/orientwalt/htdf/x/ibc/testing"
)

type SoloMachineTestSuite struct {
	suite.Suite

	solomachine      *ibctesting.Solomachine // singlesig public key
	solomachineMulti *ibctesting.Solomachine // multisig public key
	coordinator      *ibctesting.Coordinator

	// testing chain used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	store sdk.KVStore
}

func (suite *SoloMachineTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))

	suite.solomachine = ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachinesingle", "testing", 1)
	suite.solomachineMulti = ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachinemulti", "testing", 4)

	suite.store = suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), types.SoloMachine)
}

func TestSoloMachineTestSuite(t *testing.T) {
	suite.Run(t, new(SoloMachineTestSuite))
}

func (suite *SoloMachineTestSuite) GetSequenceFromStore() uint64 {
	bz := suite.store.Get(host.KeyClientState())
	suite.Require().NotNil(bz)

	var clientState exported.ClientState
	err := codec.UnmarshalAny(suite.chainA.Codec, &clientState, bz)
	suite.Require().NoError(err)
	return clientState.GetLatestHeight().GetVersionHeight()
}

func (suite *SoloMachineTestSuite) GetInvalidProof() []byte {
	invalidProof, err := suite.chainA.Codec.MarshalBinaryBare(&types.TimestampedSignatureData{Timestamp: suite.solomachine.Time})
	suite.Require().NoError(err)

	return invalidProof
}

func TestUnpackInterfaces_Header(t *testing.T) {
	registry := testdata.NewTestInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)

	pk := secp256k1.GenPrivKey().PubKey().(cryptotypes.PubKey)
	any, err := codectypes.NewAnyWithValue(pk)
	require.NoError(t, err)

	header := types.Header{
		NewPublicKey: any,
	}
	bz, err := header.Marshal()
	require.NoError(t, err)

	var header2 types.Header
	err = header2.Unmarshal(bz)
	require.NoError(t, err)

	err = codectypes.UnpackInterfaces(header2, registry)
	require.NoError(t, err)

	require.Equal(t, pk, header2.NewPublicKey.GetCachedValue())
}

func TestUnpackInterfaces_HeaderData(t *testing.T) {
	registry := testdata.NewTestInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)

	pk := secp256k1.GenPrivKey().PubKey().(cryptotypes.PubKey)
	any, err := codectypes.NewAnyWithValue(pk)
	require.NoError(t, err)

	hd := types.HeaderData{
		NewPubKey: any,
	}
	bz, err := hd.Marshal()
	require.NoError(t, err)

	var hd2 types.HeaderData
	err = hd2.Unmarshal(bz)
	require.NoError(t, err)

	err = codectypes.UnpackInterfaces(hd2, registry)
	require.NoError(t, err)

	require.Equal(t, pk, hd2.NewPubKey.GetCachedValue())
}