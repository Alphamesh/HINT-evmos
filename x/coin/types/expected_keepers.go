package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
)

type Erc20Keeper interface {
	// Methods imported from erc20 should be defined here
	Mint(ctx sdk.Context, sender sdk.AccAddress, receiver common.Address, coin sdk.Coin) error
	Burn(ctx sdk.Context, sender sdk.AccAddress, receiver common.Address, coin sdk.Coin) error
	Transfer(ctx sdk.Context, sender common.Address, receiver common.Address, coin sdk.Coin) error
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
	HasSupply(ctx sdk.Context, denom string) bool
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool)
	HasDenomMetaData(ctx sdk.Context, denom string) bool
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}
