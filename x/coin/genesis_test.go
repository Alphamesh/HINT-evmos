package coin_test

import (
	"testing"

	keepertest "github.com/evmos/evmos/v10/testutil/keeper"
	"github.com/evmos/evmos/v10/testutil/nullify"
	"github.com/evmos/evmos/v10/x/coin"
	"github.com/evmos/evmos/v10/x/coin/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:	types.DefaultParams(),
		
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.CoinKeeper(t)
	coin.InitGenesis(ctx, *k, genesisState)
	got := coin.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	

	// this line is used by starport scaffolding # genesis/test/assert
}