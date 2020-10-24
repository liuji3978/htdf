package types_test

import (
	"github.com/orientwalt/htdf/x/ibc/core/exported"
	"github.com/orientwalt/htdf/x/ibc/light-clients/06-solomachine/types"
	ibctmtypes "github.com/orientwalt/htdf/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/orientwalt/htdf/x/ibc/testing"
)

func (suite *SoloMachineTestSuite) TestCheckMisbehaviourAndUpdateState() {
	var (
		clientState  exported.ClientState
		misbehaviour exported.Misbehaviour
	)

	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		testCases := []struct {
			name    string
			setup   func()
			expPass bool
		}{
			{
				"valid misbehaviour",
				func() {
					clientState = solomachine.ClientState()
					misbehaviour = solomachine.CreateMisbehaviour()
				},
				true,
			},
			{
				"client is frozen",
				func() {
					cs := solomachine.ClientState()
					cs.FrozenSequence = 1
					clientState = cs
					misbehaviour = solomachine.CreateMisbehaviour()
				},
				false,
			},
			{
				"wrong client state type",
				func() {
					clientState = &ibctmtypes.ClientState{}
					misbehaviour = solomachine.CreateMisbehaviour()
				},
				false,
			},
			{
				"invalid misbehaviour type",
				func() {
					clientState = solomachine.ClientState()
					misbehaviour = ibctmtypes.Misbehaviour{}
				},
				false,
			},
			{
				"invalid SignatureOne SignatureData",
				func() {
					clientState = solomachine.ClientState()
					m := solomachine.CreateMisbehaviour()

					m.SignatureOne.Signature = suite.GetInvalidProof()
					misbehaviour = m
				}, false,
			},
			{
				"invalid SignatureTwo SignatureData",
				func() {
					clientState = solomachine.ClientState()
					m := solomachine.CreateMisbehaviour()

					m.SignatureTwo.Signature = suite.GetInvalidProof()
					misbehaviour = m
				}, false,
			},
			{
				"invalid first signature data",
				func() {
					clientState = solomachine.ClientState()

					// store in temp before assigning to interface type
					m := solomachine.CreateMisbehaviour()

					msg := []byte("DATA ONE")
					signBytes := &types.SignBytes{
						Sequence:    solomachine.Sequence + 1,
						Timestamp:   solomachine.Time,
						Diversifier: solomachine.Diversifier,
						DataType:    types.CLIENT,
						Data:        msg,
					}

					data, err := suite.chainA.Codec.MarshalBinaryBare(signBytes)
					suite.Require().NoError(err)

					sig := solomachine.GenerateSignature(data)

					m.SignatureOne.Signature = sig
					m.SignatureOne.Data = msg
					misbehaviour = m
				},
				false,
			},
			{
				"invalid second signature data",
				func() {
					clientState = solomachine.ClientState()

					// store in temp before assigning to interface type
					m := solomachine.CreateMisbehaviour()

					msg := []byte("DATA TWO")
					signBytes := &types.SignBytes{
						Sequence:    solomachine.Sequence + 1,
						Timestamp:   solomachine.Time,
						Diversifier: solomachine.Diversifier,
						DataType:    types.CLIENT,
						Data:        msg,
					}

					data, err := suite.chainA.Codec.MarshalBinaryBare(signBytes)
					suite.Require().NoError(err)

					sig := solomachine.GenerateSignature(data)

					m.SignatureTwo.Signature = sig
					m.SignatureTwo.Data = msg
					misbehaviour = m
				},
				false,
			},
			{
				"wrong pubkey generates first signature",
				func() {
					clientState = solomachine.ClientState()
					badMisbehaviour := solomachine.CreateMisbehaviour()

					// update public key to a new one
					solomachine.CreateHeader()
					m := solomachine.CreateMisbehaviour()

					// set SignatureOne to use the wrong signature
					m.SignatureOne = badMisbehaviour.SignatureOne
					misbehaviour = m
				}, false,
			},
			{
				"wrong pubkey generates second signature",
				func() {
					clientState = solomachine.ClientState()
					badMisbehaviour := solomachine.CreateMisbehaviour()

					// update public key to a new one
					solomachine.CreateHeader()
					m := solomachine.CreateMisbehaviour()

					// set SignatureTwo to use the wrong signature
					m.SignatureTwo = badMisbehaviour.SignatureTwo
					misbehaviour = m
				}, false,
			},

			{
				"signatures sign over different sequence",
				func() {
					clientState = solomachine.ClientState()

					// store in temp before assigning to interface type
					m := solomachine.CreateMisbehaviour()

					// Signature One
					msg := []byte("DATA ONE")
					// sequence used is plus 1
					signBytes := &types.SignBytes{
						Sequence:    solomachine.Sequence + 1,
						Timestamp:   solomachine.Time,
						Diversifier: solomachine.Diversifier,
						DataType:    types.CLIENT,
						Data:        msg,
					}

					data, err := suite.chainA.Codec.MarshalBinaryBare(signBytes)
					suite.Require().NoError(err)

					sig := solomachine.GenerateSignature(data)

					m.SignatureOne.Signature = sig
					m.SignatureOne.Data = msg

					// Signature Two
					msg = []byte("DATA TWO")
					// sequence used is minus 1

					signBytes = &types.SignBytes{
						Sequence:    solomachine.Sequence - 1,
						Timestamp:   solomachine.Time,
						Diversifier: solomachine.Diversifier,
						DataType:    types.CLIENT,
						Data:        msg,
					}
					data, err = suite.chainA.Codec.MarshalBinaryBare(signBytes)
					suite.Require().NoError(err)

					sig = solomachine.GenerateSignature(data)

					m.SignatureTwo.Signature = sig
					m.SignatureTwo.Data = msg

					misbehaviour = m

				},
				false,
			},
		}

		for _, tc := range testCases {
			tc := tc

			suite.Run(tc.name, func() {
				// setup test
				tc.setup()

				clientState, err := clientState.CheckMisbehaviourAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), suite.store, misbehaviour)

				if tc.expPass {
					suite.Require().NoError(err)
					suite.Require().True(clientState.IsFrozen(), "client not frozen")
				} else {
					suite.Require().Error(err)
					suite.Require().Nil(clientState)
				}
			})
		}
	}
}