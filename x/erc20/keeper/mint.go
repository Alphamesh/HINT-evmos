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

// MintingEnabled checks that:
//   - the global parameter for erc20 conversion is enabled
//   - minting is enabled for the given (erc20,coin) token pair
//   - recipient address is not on the blocked list
//   - bank module transfers are enabled for the Cosmos coin
func (k Keeper) MintingEnabled(
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

	if !pair.Enabled {
		return types.TokenPair{}, errorsmod.Wrapf(
			types.ErrERC20TokenPairDisabled, "minting token '%s' is not enabled by governance", token,
		)
	}

	if k.bankKeeper.BlockedAddr(receiver.Bytes()) {
		return types.TokenPair{}, errorsmod.Wrapf(
			errortypes.ErrUnauthorized, "%s is not allowed to receive transactions", receiver,
		)
	}

	// NOTE: ignore amount as only denom is checked on IsSendEnabledCoin
	coin := sdk.Coin{Denom: pair.Denom}

	// check if minting to a recipient address other than the sender is enabled
	// for the given coin denom
	if !sender.Equals(receiver) && !k.bankKeeper.IsSendEnabledCoin(ctx, coin) {
		return types.TokenPair{}, errorsmod.Wrapf(
			banktypes.ErrSendDisabled, "minting '%s' coins to an external address is currently disabled", token,
		)
	}

	return pair, nil
}

// Mint erc20 tokens:
//   - only allow minting from authorized modules
func (k Keeper) Mint(ctx sdk.Context, sender sdk.AccAddress, receiver common.Address, coin sdk.Coin) error {

	sdkReceiver, _ := sdk.AccAddressFromHexUnsafe(receiver.Hex())

	pair, err := k.MintingEnabled(ctx, sender, sdkReceiver, coin.Denom)
	if err != nil {
		return err
	}

	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	contract := pair.GetERC20Contract()

	balanceToken := k.BalanceOf(ctx, erc20, contract, receiver)
	if balanceToken == nil {
		return errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve balance")
	}

	// Mint tokens and send to receiver
	_, err = k.CallEVM(ctx, erc20, types.ModuleAddress, contract, true, "mint", receiver, coin.Amount.BigInt())
	if err != nil {
		return err
	}

	// Check expected receiver balance after transfer
	tokens := coin.Amount.BigInt()
	balanceTokenAfter := k.BalanceOf(ctx, erc20, contract, receiver)
	if balanceTokenAfter == nil {
		return errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve balance")
	}
	expToken := big.NewInt(0).Add(balanceToken, tokens)

	if r := balanceTokenAfter.Cmp(expToken); r != 0 {
		return errorsmod.Wrapf(
			types.ErrBalanceInvariance,
			"invalid token balance - expected: %v, actual: %v", expToken, balanceTokenAfter,
		)
	}

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeMint,
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

// Burn erc20 tokens:
//   - only allow burning from authorized modules
func (k Keeper) Burn(ctx sdk.Context, burner sdk.AccAddress, burnee common.Address, coin sdk.Coin) error {

	sdkReceiver, _ := sdk.AccAddressFromHexUnsafe(burnee.Hex())

	pair, err := k.MintingEnabled(ctx, burner, sdkReceiver, coin.Denom)
	if err != nil {
		return err
	}

	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	contract := pair.GetERC20Contract()

	balanceToken := k.BalanceOf(ctx, erc20, contract, burnee)
	if balanceToken == nil {
		return errorsmod.Wrap(types.ErrEVMCall, "failed to burnee balance")
	}

	// Mint tokens and send to receiver
	_, err = k.CallEVM(ctx, erc20, types.ModuleAddress, contract, true, "burnCoins", burnee, coin.Amount.BigInt())
	if err != nil {
		return err
	}

	// Check expected receiver balance after transfer
	tokens := coin.Amount.BigInt()
	balanceTokenAfter := k.BalanceOf(ctx, erc20, contract, burnee)
	if balanceTokenAfter == nil {
		return errorsmod.Wrap(types.ErrEVMCall, "failed to burnee balance")
	}
	expToken := big.NewInt(0).Sub(balanceToken, tokens)

	if r := balanceTokenAfter.Cmp(expToken); r != 0 {
		return errorsmod.Wrapf(
			types.ErrBalanceInvariance,
			"invalid token balance - expected: %v, actual: %v", expToken, balanceTokenAfter,
		)
	}

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeBurn,
				sdk.NewAttribute(sdk.AttributeKeySender, burner.String()),
				sdk.NewAttribute(types.AttributeKeyReceiver, burnee.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, coin.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyCosmosCoin, coin.Denom),
				sdk.NewAttribute(types.AttributeKeyERC20Token, pair.Erc20Address),
			),
		},
	)

	return nil
}
