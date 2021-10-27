package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

const (
	erc20Denom       = "coin"
	erc20Symbol      = "token"
	cosmosTokenDenom = "coinevm"
	defaultExponent  = uint32(18)
	zeroExponent     = uint32(0)
)

func (suite *KeeperTestSuite) setupNewTokenPair() common.Address {
	suite.SetupTest()
	contractAddr := suite.DeployContract(erc20Denom, erc20Symbol)
	suite.Commit()
	pair := types.NewTokenPair(contractAddr, cosmosTokenDenom, true)
	err := suite.app.IntrarelayerKeeper.RegisterTokenPair(suite.ctx, pair)
	suite.Require().NoError(err)
	return contractAddr
}

func (suite *KeeperTestSuite) TestConvertCoin() {
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{"coin not registered", func() {}, false},
		{
			"coin registered - insufficient funds",
			func() {
				pair := types.NewTokenPair(tests.GenerateAddress(), erc20Denom, true)
				id := pair.GetID()
				suite.app.IntrarelayerKeeper.SetTokenPair(suite.ctx, pair)
				suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, pair.Denom, id)
				suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, pair.GetERC20Contract(), id)
			},
			false,
		},
		// TODO use mint contract with ABI
		{
			"coin registered - sufficient funds - callEVM",
			func() {
				contractAddr := suite.setupNewTokenPair()
				id := suite.app.IntrarelayerKeeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, found := suite.app.IntrarelayerKeeper.GetTokenPair(suite.ctx, id)

				suite.Require().NotNil(pair)
				suite.Require().True(found)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			// coins := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100)))
			// err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			// suite.Require().NoError(err)
			// err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
			// suite.Require().NoError(err)

			ctx := sdk.WrapSDKContext(suite.ctx)
			sender := sdk.AccAddress(tests.GenerateAddress().Bytes())
			receiver := tests.GenerateAddress()
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(erc20Denom, sdk.NewInt(100)),
				receiver,
				sender,
			)
			res, err := suite.app.IntrarelayerKeeper.ConvertCoin(ctx, msg)
			expRes := &types.MsgConvertCoinResponse{}

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestConvertECR20() {
	erc20 := tests.GenerateAddress()
	// denom := "coin"
	// pair := types.NewTokenPair(erc20, denom, true)
	// id := pair.GetID()

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{"coin not registered", func() {}, false},
		// TODO: use burn contract with ABI
		// {
		// 	"erc20 has no burn method",
		// 	func() {
		// 		suite.app.IntrarelayerKeeper.SetTokenPair(suite.ctx, pair)
		// 		suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, pair.Denom, id)
		// 		suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, pair.GetERC20Contract(), id)
		// 	},
		// 	true,
		// },
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			msg := types.NewMsgConvertERC20(
				sdk.NewInt(100),
				sdk.AccAddress{},
				erc20,
				tests.GenerateAddress(),
			)
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.app.IntrarelayerKeeper.ConvertERC20(ctx, msg)
			expRes := &types.MsgConvertERC20Response{}

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}