package types_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/orientwalt/htdf/codec"
	cryptotypes "github.com/orientwalt/htdf/crypto/types"
	"github.com/orientwalt/htdf/simapp"
	sdk "github.com/orientwalt/htdf/types"
	clienttypes "github.com/orientwalt/htdf/x/ibc/core/02-client/types"
	ibctmtypes "github.com/orientwalt/htdf/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/orientwalt/htdf/x/ibc/testing"
	ibctestingmock "github.com/orientwalt/htdf/x/ibc/testing/mock"
)

const (
	chainID                       = "gaia"
	chainIDVersion0               = "gaia-version-0"
	chainIDVersion1               = "gaia-version-1"
	clientID                      = "gaiamainnet"
	trustingPeriod  time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod       time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift   time.Duration = time.Second * 10
)

var (
	height          = clienttypes.NewHeight(0, 4)
	newClientHeight = clienttypes.NewHeight(1, 1)
	upgradePath     = fmt.Sprintf("%s/%s", "upgrade", "upgradedClient")
)

type TendermintTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	// TODO: deprecate usage in favor of testing package
	ctx        sdk.Context
	cdc        codec.Marshaler
	privVal    tmtypes.PrivValidator
	valSet     *tmtypes.ValidatorSet
	valsHash   tmbytes.HexBytes
	header     *ibctmtypes.Header
	now        time.Time
	headerTime time.Time
	clientTime time.Time
}

func (suite *TendermintTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))

	// TODO: deprecate usage in favor of testing package
	checkTx := false
	app := simapp.Setup(checkTx)

	suite.cdc = app.AppCodec()

	// now is the time of the current chain, must be after the updating header
	// mocks ctx.BlockTime()
	suite.now = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	suite.clientTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	// Header time is intended to be time for any new header used for updates
	suite.headerTime = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	suite.privVal = ibctestingmock.NewPV()

	pubKey, err := suite.privVal.GetPubKey()
	suite.Require().NoError(err)

	heightMinus1 := clienttypes.NewHeight(0, height.VersionHeight-1)

	val := tmtypes.NewValidator(pubKey.(cryptotypes.IntoTmPubKey).AsTmPubKey(), 10)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{val})
	suite.valsHash = suite.valSet.Hash()
	suite.header = ibctmtypes.CreateTestHeader(chainID, height, heightMinus1, suite.now, suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1, Time: suite.now})
}

func TestTendermintTestSuite(t *testing.T) {
	suite.Run(t, new(TendermintTestSuite))
}
