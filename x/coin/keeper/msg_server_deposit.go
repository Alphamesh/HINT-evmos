package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v10/x/coin/types"
)

func (k msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Error checked during msg validation
	receiver := common.HexToAddress(msg.Account)
	sender := sdk.MustAccAddressFromBech32(msg.Creator)

	amount, ok := sdk.NewIntFromString(msg.Amount)
	if !ok {
		return nil, errorsmod.Wrapf(errors.New("invalid deposit amount"), "received: %v", msg.Amount)
	}
	coin := sdk.NewCoin(msg.Denom, amount)

	err := k.erc20Keeper.Mint(ctx, sender, receiver, coin)
	if err != nil {
		return nil, err
	}

	return &types.MsgDepositResponse{}, nil
}
