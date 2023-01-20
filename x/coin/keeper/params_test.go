package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "github.com/evmos/evmos/v10/testutil/keeper"
	"github.com/evmos/evmos/v10/x/coin/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.CoinKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
