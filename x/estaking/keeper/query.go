package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/elys-network/elys/x/estaking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Rewards(goCtx context.Context, req *types.QueryRewardsRequest) (*types.QueryRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "empty delegator address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	total := sdk.DecCoins{}
	var delRewards []types.DelegationDelegatorReward

	delAdr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	k.IterateDelegations(
		ctx, delAdr,
		func(_ int64, del stakingtypes.DelegationI) (stop bool) {
			valAddr := del.GetValidatorAddr()
			val := k.Keeper.Validator(ctx, valAddr)
			endingPeriod := k.distrKeeper.IncrementValidatorPeriod(ctx, val)
			delReward := k.distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

			delRewards = append(delRewards, types.DelegationDelegatorReward{
				ValidatorAddress: valAddr.String(),
				Reward:           delReward,
			})
			total = total.Add(delReward...)
			return false
		},
	)

	return &types.QueryRewardsResponse{Rewards: delRewards, Total: total}, nil
}
