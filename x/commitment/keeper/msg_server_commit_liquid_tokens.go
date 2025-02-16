package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	aptypes "github.com/elys-network/elys/x/assetprofile/types"
	"github.com/elys-network/elys/x/commitment/types"
)

// CommitLiquidTokens commit the tokens from user's balance
func (k msgServer) CommitLiquidTokens(goCtx context.Context, msg *types.MsgCommitLiquidTokens) (*types.MsgCommitLiquidTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	assetProfile, found := k.assetProfileKeeper.GetEntry(ctx, msg.Denom)
	if !found {
		return nil, errorsmod.Wrapf(aptypes.ErrAssetProfileNotFound, "denom: %s", msg.Denom)
	}

	if !assetProfile.CommitEnabled {
		return nil, errorsmod.Wrapf(types.ErrCommitDisabled, "denom: %s", msg.Denom)
	}

	depositCoins := sdk.NewCoins(sdk.NewCoin(msg.Denom, msg.Amount))

	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "unable to convert address from bech32")
	}

	// send the deposited coins to the module
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, depositCoins)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, fmt.Sprintf("unable to send deposit tokens: %v", depositCoins))
	}

	// Update total commitment
	params := k.GetParams(ctx)
	params.TotalCommitted = params.TotalCommitted.Add(depositCoins...)
	k.SetParams(ctx, params)

	// Get the Commitments for the creator
	commitments := k.GetCommitments(ctx, msg.Creator)

	// Update the commitments
	commitments.AddCommittedTokens(msg.Denom, msg.Amount, msg.LockUntil)
	k.SetCommitments(ctx, commitments)

	// Emit Hook commitment changed
	err = k.CommitmentChanged(ctx, msg.Creator, sdk.Coins{sdk.NewCoin(msg.Denom, msg.Amount)})
	if err != nil {
		return nil, err
	}

	// Emit blockchain event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommitmentChanged,
			sdk.NewAttribute(types.AttributeCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeDenom, msg.Denom),
		),
	)

	return &types.MsgCommitLiquidTokensResponse{}, nil
}
