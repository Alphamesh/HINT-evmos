package keeper

import (
	"github.com/evmos/evmos/v10/x/coin/types"
)

var _ types.QueryServer = Keeper{}
