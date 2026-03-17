package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Position represents a liquidity position returned by positions() function
// Matches the return values from NonfungiblePositionManager.positions(tokenId)
type Position struct {
	Nonce                    *big.Int       `json:"nonce"`                    // uint88
	Operator                 common.Address `json:"operator"`                 // address
	Token0                   common.Address `json:"token0"`                   // address
	Token1                   common.Address `json:"token1"`                   // address
	Deployer                 common.Address `json:"deployer"`                 // address
	TickLower                int32          `json:"tickLower"`                // int24
	TickUpper                int32          `json:"tickUpper"`                // int24
	Liquidity                *big.Int       `json:"liquidity"`                // uint128
	FeeGrowthInside0LastX128 *big.Int       `json:"feeGrowthInside0LastX128"` // uint256
	FeeGrowthInside1LastX128 *big.Int       `json:"feeGrowthInside1LastX128"` // uint256
	TokensOwed0              *big.Int       `json:"tokensOwed0"`              // uint128
	TokensOwed1              *big.Int       `json:"tokensOwed1"`              // uint128
}

type PoolType int

const (
	CL1 PoolType = iota
	CL200
)

func (p PoolType) PoolNonce() *big.Int {
	switch p {
	case CL1:
		return big.NewInt(1)
	case CL200:
		return big.NewInt(3)
	default:
		return big.NewInt(3)
	}
}

func (p PoolType) TickSpacing() int {
	switch p {
	case CL1:
		return 1 // memo. CL1에 대해선 200만큼 조정해서 진입. 조정 없을 시, 바로 아웃오브레인지 되는 경우가 많음
	case CL200:
		return 200
	default:
		return 200
	}
}
