// +build norace

package rest_test

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/orientwalt/htdf/testutil"
	sdk "github.com/orientwalt/htdf/types"
	grpctypes "github.com/orientwalt/htdf/types/grpc"
	"github.com/orientwalt/htdf/types/query"
	"github.com/orientwalt/htdf/types/rest"
	"github.com/orientwalt/htdf/x/bank/types"
)

func (s *IntegrationTestSuite) TestTotalSupplyGRPCHandler() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		respType proto.Message
		expected proto.Message
	}{
		{
			"test GRPC total supply",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			&types.QueryTotalSupplyResponse{},
			&types.QueryTotalSupplyResponse{
				Supply: sdk.NewCoins(
					sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
					sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(sdk.NewInt(10))),
				),
			},
		},
		{
			"GRPC total supply of a specific denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/%s", baseURL, s.cfg.BondDenom),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			&types.QuerySupplyOfResponse{},
			&types.QuerySupplyOfResponse{
				Amount: sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(sdk.NewInt(10))),
			},
		},
		{
			"Query for `height` > 1",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/%s", baseURL, s.cfg.BondDenom),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			&types.QuerySupplyOfResponse{},
			&types.QuerySupplyOfResponse{
				Amount: sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(sdk.NewInt(20))),
			},
		},
		{
			"Query params shouldn't be considered as height",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/%s?height=2", baseURL, s.cfg.BondDenom),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			&types.QuerySupplyOfResponse{},
			&types.QuerySupplyOfResponse{
				Amount: sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(sdk.NewInt(10))),
			},
		},
		{
			"GRPC total supply of a bogus denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/foobar", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			&types.QuerySupplyOfResponse{},
			&types.QuerySupplyOfResponse{
				Amount: sdk.NewCoin("foobar", sdk.ZeroInt()),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected.String(), tc.respType.String())
		})
	}
}

func (s *IntegrationTestSuite) TestBalancesGRPCHandler() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		respType proto.Message
		expected proto.Message
	}{
		{
			"gRPC total account balance",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", baseURL, val.Address.String()),
			&types.QueryAllBalancesResponse{},
			&types.QueryAllBalancesResponse{
				Balances: sdk.NewCoins(
					sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
					sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
				),
				Pagination: &query.PageResponse{
					Total: 2,
				},
			},
		},
		{
			"gPRC account balance of a denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s/%s", baseURL, val.Address.String(), s.cfg.BondDenom),
			&types.QueryBalanceResponse{},
			&types.QueryBalanceResponse{
				Balance: &sdk.Coin{
					Denom:  s.cfg.BondDenom,
					Amount: s.cfg.StakingTokens.Sub(s.cfg.BondedTokens),
				},
			},
		},
		{
			"gPRC account balance of a bogus denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s/foobar", baseURL, val.Address.String()),
			&types.QueryBalanceResponse{},
			&types.QueryBalanceResponse{
				Balance: &sdk.Coin{
					Denom:  "foobar",
					Amount: sdk.NewInt(0),
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected.String(), tc.respType.String())
		})
	}
}