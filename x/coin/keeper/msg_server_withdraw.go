package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	"errors"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/coin/types"
)

func (k msgServer) Withdraw(goCtx context.Context, msg *types.MsgWithdraw) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Error checked during msg validation
	receiver := common.HexToAddress(msg.Account)
	sender := sdk.MustAccAddressFromBech32(msg.Creator)

	amount, ok := sdk.NewIntFromString(msg.Amount)
	if !ok {
		return nil, errorsmod.Wrapf(errors.New("invalid withdraw amount"), "received: %v", msg.Amount)
	}
	coin := sdk.NewCoin(msg.Denom, amount)

	err := k.erc20Keeper.Burn(ctx, sender, receiver, coin)
	if err != nil {
		return nil, err
	}

	// TODO:
	// Send a event to notify other chains

	return &types.MsgWithdrawResponse{}, nil
}
