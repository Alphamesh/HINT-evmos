package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	"errors"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/coin/types"
)

func (k msgServer) Transfer(goCtx context.Context, msg *types.MsgTransfer) (*types.MsgTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Error checked during msg validation
	sender := common.HexToAddress(msg.From)
	receiver := common.HexToAddress(msg.To)

	// TODO: validate creator
	_ = sdk.MustAccAddressFromBech32(msg.Creator)

	amount, ok := sdk.NewIntFromString(msg.Amount)
	if !ok {
		return nil, errorsmod.Wrapf(errors.New("invalid withdraw amount"), "received: %v", msg.Amount)
	}
	coin := sdk.NewCoin(msg.Denom, amount)

	err := k.erc20Keeper.Transfer(ctx, sender, receiver, coin)
	if err != nil {
		return nil, err
	}

	return &types.MsgTransferResponse{}, nil
}
