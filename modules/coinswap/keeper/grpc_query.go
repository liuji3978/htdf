package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/orientwalt/htdf/types"

	"github.com/orientwalt/htdf/modules/coinswap/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Liquidity(c context.Context, req *types.QueryLiquidityRequest) (*types.QueryLiquidityResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	if err := types.CheckUniDenom(req.Id); err != nil {
		return nil, err
	}

	uniDenom := req.Id

	tokenDenom, err := types.GetCoinDenomFromUniDenom(uniDenom)
	if err != nil {
		return nil, err
	}

	reservePool := k.GetReservePool(ctx, req.Id)

	standardDenom := k.GetParams(ctx).StandardDenom
	standard := sdk.NewCoin(standardDenom, reservePool.AmountOf(standardDenom))
	token := sdk.NewCoin(tokenDenom, reservePool.AmountOf(tokenDenom))
	liquidity := sdk.NewCoin(uniDenom, k.bk.GetSupply(ctx).GetTotal().AmountOf(uniDenom))

	swapParams := k.GetParams(ctx)
	fee := swapParams.Fee.String()
	res := types.QueryLiquidityResponse{
		Standard:  standard,
		Token:     token,
		Liquidity: liquidity,
		Fee:       fee,
	}
	return &res, nil
}