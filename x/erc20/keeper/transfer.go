package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v10/contracts"
	"math/big"

	"github.com/evmos/evmos/v10/x/erc20/types"
)

// TransferEnabled checks that:
//   - the global parameter for erc20 conversion is enabled
//   - recipient address is not on the blocked list
//   - bank module transfers are enabled for the Cosmos coin
func (k Keeper) TransferEnabled(
	ctx sdk.Context,
	sender, receiver sdk.AccAddress,
	token string,
) (types.TokenPair, error) {
	if !k.IsERC20Enabled(ctx) {
		return types.TokenPair{}, errorsmod.Wrap(
			types.ErrERC20Disabled, "module is currently disabled by governance",
		)
	}

	id := k.GetTokenPairID(ctx, token)
	if len(id) == 0 {
		return types.TokenPair{}, errorsmod.Wrapf(
			types.ErrTokenPairNotFound, "token '%s' not registered by id", token,
		)
	}

	pair, found := k.GetTokenPair(ctx, id)
	if !found {
		return types.TokenPair{}, errorsmod.Wrapf(
			types.ErrTokenPairNotFound, "token '%s' not registered", token,
		)
	}

	if k.bankKeeper.BlockedAddr(receiver.Bytes()) {
		return types.TokenPair{}, errorsmod.Wrapf(
			errortypes.ErrUnauthorized, "%s is not allowed to receive transactions", receiver,
		)
	}

	// NOTE: ignore amount as only denom is checked on IsSendEnabledCoin
	coin := sdk.Coin{Denom: pair.Denom}

	// check if sending a coin is enabled
	// for the given coin denom
	if !sender.Equals(receiver) && !k.bankKeeper.IsSendEnabledCoin(ctx, coin) {
		return types.TokenPair{}, errorsmod.Wrapf(
			banktypes.ErrSendDisabled, "transferring '%s' coins to an external address is currently disabled", token,
		)
	}

	return pair, nil
}

// Transfer erc20 tokens:
// transfer from account to address
func (k Keeper) Transfer(ctx sdk.Context, sender common.Address, receiver common.Address, coin sdk.Coin) error {

	sdkSender, _ := sdk.AccAddressFromHexUnsafe(sender.Hex())
	sdkReceiver, _ := sdk.AccAddressFromHexUnsafe(receiver.Hex())

	pair, err := k.TransferEnabled(ctx, sdkSender, sdkReceiver, coin.Denom)
	if err != nil {
		return err
	}

	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	contract := pair.GetERC20Contract()

	senderBalance := k.BalanceOf(ctx, erc20, contract, sender)
	if senderBalance == nil {
		return errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve sender balance")
	}

	receiverBalance := k.BalanceOf(ctx, erc20, contract, receiver)
	if receiverBalance == nil {
		return errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve receiver balance")
	}

	// Transfer tokens from sender to receiver
	_, err = k.CallEVM(ctx, erc20, sender, contract, true, "transfer", receiver, coin.Amount.BigInt())
	if err != nil {
		return err
	}

	// Check expected sender balance after transfer
	tokens := coin.Amount.BigInt()
	senderBalanceAfter := k.BalanceOf(ctx, erc20, contract, sender)
	if senderBalanceAfter == nil {
		return errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve sender balance")
	}
	expSenderToken := big.NewInt(0).Sub(senderBalance, tokens)

	if r := senderBalanceAfter.Cmp(expSenderToken); r != 0 {
		return errorsmod.Wrapf(
			types.ErrBalanceInvariance,
			"invalid sender token balance - expected: %v, actual: %v", expSenderToken, senderBalanceAfter,
		)
	}

	// Check expected receiver balance after transfer
	receiverBalanceAfter := k.BalanceOf(ctx, erc20, contract, receiver)
	if receiverBalanceAfter == nil {
		return errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve receiver balance")
	}
	expReceiverToken := big.NewInt(0).Add(receiverBalance, tokens)

	if r := receiverBalanceAfter.Cmp(expReceiverToken); r != 0 {
		return errorsmod.Wrapf(
			types.ErrBalanceInvariance,
			"invalid receiver token balance - expected: %v, actual: %v", expReceiverToken, receiverBalanceAfter,
		)
	}

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeTransfer,
				sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
				sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, coin.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyCosmosCoin, coin.Denom),
				sdk.NewAttribute(types.AttributeKeyERC20Token, pair.Erc20Address),
			),
		},
	)

	return nil
}
