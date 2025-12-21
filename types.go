package blackholedex

import (
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Route represents a single swap route in the BlackholeDEX router
// Matches the Solidity struct: IRouter.route
type Route struct {
	Pair         common.Address `json:"pair"`
	From         common.Address `json:"from"`
	To           common.Address `json:"to"`
	Stable       bool           `json:"stable"`
	Concentrated bool           `json:"concentrated"`
	Receiver     common.Address `json:"receiver"`
}

// SWAPExactETHForTokensParams represents parameters for swapExactETHForTokens function
type SWAPExactETHForTokensParams struct {
	AmountOutMin *big.Int       `json:"amountOutMin"`
	Routes       []Route        `json:"routes"`
	To           common.Address `json:"to"`
	Deadline     *big.Int       `json:"deadline"`
}

// SWAPExactETHForTokensParams represents parameters for swapExactTokensForTokens function
type SWAPExactTokensForTokensParams struct {
	AmountIn     *big.Int       `json:"amountIn"`
	AmountOutMin *big.Int       `json:"amountOutMin"`
	Routes       []Route        `json:"routes"`
	To           common.Address `json:"to"`
	Deadline     *big.Int       `json:"deadline"`
}

// AddLiquidityParams represents parameters for addLiquidity function
// 미확인. 실제 유동성 공급 시에는 MintParams 사용.
type AddLiquidityParams struct {
	TokenA         common.Address `json:"tokenA"`
	TokenB         common.Address `json:"tokenB"`
	Stable         bool           `json:"stable"`
	AmountADesired *big.Int       `json:"amountADesired"`
	AmountBDesired *big.Int       `json:"amountBDesired"`
	AmountAMin     *big.Int       `json:"amountAMin"`
	AmountBMin     *big.Int       `json:"amountBMin"`
	To             common.Address `json:"to"`
	Deadline       *big.Int       `json:"deadline"`
}

// RemoveLiquidityParams represents parameters for removeLiquidity function
type RemoveLiquidityParams struct {
	TokenA     common.Address `json:"tokenA"`
	TokenB     common.Address `json:"tokenB"`
	Stable     bool           `json:"stable"`
	Liquidity  *big.Int       `json:"liquidity"`
	AmountAMin *big.Int       `json:"amountAMin"`
	AmountBMin *big.Int       `json:"amountBMin"`
	To         common.Address `json:"to"`
	Deadline   *big.Int       `json:"deadline"`
}

// VotingEscrow types

// CreateLockParams represents parameters for create_lock function
type CreateLockParams struct {
	Value        *big.Int `json:"value"`
	LockDuration *big.Int `json:"lockDuration"` // in seconds
}

// IncreaseAmountParams represents parameters for increase_amount function
type IncreaseAmountParams struct {
	TokenID *big.Int `json:"tokenId"`
	Value   *big.Int `json:"value"`
}

// IncreaseUnlockTimeParams represents parameters for increase_unlock_time function
type IncreaseUnlockTimeParams struct {
	TokenID      *big.Int `json:"tokenId"`
	LockDuration *big.Int `json:"lockDuration"`
}

// WithdrawParams represents parameters for withdraw function
type WithdrawParams struct {
	TokenID *big.Int `json:"tokenId"`
}

// LockedBalance represents the locked balance information
type LockedBalance struct {
	Amount *big.Int `json:"amount"`
	End    *big.Int `json:"end"`
}

// Voter types

// VoteParams represents parameters for vote function
type VoteParams struct {
	TokenID *big.Int         `json:"tokenId"`
	Pools   []common.Address `json:"pools"`
	Weights []*big.Int       `json:"weights"`
}

// Gauge types

// GaugeDepositParams represents parameters for gauge deposit function
type GaugeDepositParams struct {
	Amount  *big.Int `json:"amount"`
	TokenID *big.Int `json:"tokenId"`
}

// GaugeWithdrawParams represents parameters for gauge withdraw function
type GaugeWithdrawParams struct {
	Amount *big.Int `json:"amount"`
}

// GetRewardParams represents parameters for getReward function
type GetRewardParams struct {
	Account common.Address   `json:"account"`
	Tokens  []common.Address `json:"tokens"`
}

// NonfungiblePositionManager types

// MintParams represents parameters for mint function in NonfungiblePositionManager
// Matches the Solidity struct: INonfungiblePositionManager.MintParams
type MintParams struct {
	Token0         common.Address `json:"token0"`
	Token1         common.Address `json:"token1"`
	Deployer       common.Address `json:"deployer"`
	TickLower      *big.Int       `json:"tickLower"`
	TickUpper      *big.Int       `json:"tickUpper"`
	Amount0Desired *big.Int       `json:"amount0Desired"`
	Amount1Desired *big.Int       `json:"amount1Desired"`
	Amount0Min     *big.Int       `json:"amount0Min"`
	Amount1Min     *big.Int       `json:"amount1Min"`
	Recipient      common.Address `json:"recipient"`
	Deadline       *big.Int       `json:"deadline"`
}

// ERC20 types

// ApproveParams represents parameters for ERC20 approve function
type ApproveParams struct {
	Spender common.Address `json:"spender"`
	Amount  *big.Int       `json:"amount"`
}

// Pool State types

// AMMState represents the state of an AMM pool
// Returned by safelyGetStateOfAMM function in IAlgebraPoolState
type AMMState struct {
	SqrtPrice       *big.Int `json:"sqrtPrice"`       // uint160 - Current sqrt price
	Tick            int32    `json:"tick"`            // int24 - Current tick
	LastFee         uint16   `json:"lastFee"`         // uint16 - Last swap fee
	PluginConfig    uint8    `json:"pluginConfig"`    // uint8 - Plugin configuration
	ActiveLiquidity *big.Int `json:"activeLiquidity"` // uint128 - Active liquidity
	NextTick        int32    `json:"nextTick"`        // int24 - Next initialized tick
	PreviousTick    int32    `json:"previousTick"`    // int24 - Previous initialized tick
}

// Liquidity Staking types

// TransactionRecord tracks individual transaction details for financial transparency
type TransactionRecord struct {
	TxHash    common.Hash // Transaction hash
	GasUsed   uint64      // Gas consumed
	GasPrice  *big.Int    // Effective gas price (wei)
	GasCost   *big.Int    // Total gas cost (wei) = GasUsed * GasPrice
	Timestamp time.Time   // Transaction timestamp
	Operation string      // Operation type ("ApproveWAVAX", "ApproveUSDC", "Mint")
}

// StakingResult represents the complete output of staking operation
type StakingResult struct {
	NFTTokenID     *big.Int            // Liquidity position NFT token ID
	ActualAmount0  *big.Int            // Actual WAVAX staked (wei)
	ActualAmount1  *big.Int            // Actual USDC staked (smallest unit)
	FinalTickLower int32               // Final lower tick bound
	FinalTickUpper int32               // Final upper tick bound
	Transactions   []TransactionRecord // All transactions executed
	TotalGasCost   *big.Int            // Sum of all gas costs (wei)
	Success        bool                // Whether operation succeeded
	ErrorMessage   string              // Error message if failed (empty if success)
}

// Unstake types

// IncentiveKey identifies a specific farming incentive program
// Matches the Solidity struct: IncentiveKey from FarmingCenter.sol
type IncentiveKey struct {
	RewardToken      common.Address `json:"rewardToken"`      // Primary reward token address
	BonusRewardToken common.Address `json:"bonusRewardToken"` // Bonus reward token address (can be zero)
	Pool             common.Address `json:"pool"`             // WAVAX/USDC pool address
	Nonce            *big.Int       `json:"nonce"`            // Incentive nonce/version
}

// RewardAmounts tracks rewards collected during unstake operation
type RewardAmounts struct {
	Reward           *big.Int       `json:"reward"`           // Primary reward amount
	BonusReward      *big.Int       `json:"bonusReward"`      // Bonus reward amount
	RewardToken      common.Address `json:"rewardToken"`      // Primary reward token address
	BonusRewardToken common.Address `json:"bonusRewardToken"` // Bonus reward token address
}

// UnstakeResult represents the complete output of unstake operation
type UnstakeResult struct {
	NFTTokenID   *big.Int            // Unstaked NFT token ID
	Rewards      *RewardAmounts      // Rewards collected (nil if not collected)
	Transactions []TransactionRecord // All transactions executed
	TotalGasCost *big.Int            // Sum of all gas costs (wei)
	Success      bool                // Whether operation succeeded
	ErrorMessage string              // Error message if failed (empty if success)
}

// UnstakeParams contains parameters for unstaking an NFT position
type UnstakeParams struct {
	NFTTokenID     *big.Int      `json:"nftTokenId"`     // ERC721 token ID
	IncentiveKey   *IncentiveKey `json:"incentiveKey"`   // Farming incentive to exit
	CollectRewards bool          `json:"collectRewards"` // Whether to collect rewards
}

// Validate checks if UnstakeParams are valid
func (p *UnstakeParams) Validate() error {
	if p.NFTTokenID == nil || p.NFTTokenID.Sign() <= 0 {
		return errors.New("NFT token ID must be positive")
	}
	if p.IncentiveKey == nil {
		return errors.New("IncentiveKey is required")
	}
	if p.IncentiveKey.Pool == (common.Address{}) {
		return errors.New("Pool address cannot be zero")
	}
	if p.IncentiveKey.Nonce == nil {
		return errors.New("Nonce cannot be nil")
	}
	return nil
}

// Withdraw types

// WithdrawResult represents the complete output of withdrawal operation
type WithdrawResult struct {
	NFTTokenID   *big.Int            // Withdrawn NFT token ID
	Amount0      *big.Int            // WAVAX withdrawn (wei)
	Amount1      *big.Int            // USDC withdrawn (smallest unit)
	Transactions []TransactionRecord // All transactions executed
	TotalGasCost *big.Int            // Sum of all gas costs (wei)
	Success      bool                // Whether operation succeeded
	ErrorMessage string              // Error message if failed (empty if success)
}

// DecreaseLiquidityParams for decreaseLiquidity operation
type DecreaseLiquidityParams struct {
	TokenId    *big.Int `json:"tokenId"`
	Liquidity  *big.Int `json:"liquidity"`   // uint128
	Amount0Min *big.Int `json:"amount0Min"`
	Amount1Min *big.Int `json:"amount1Min"`
	Deadline   *big.Int `json:"deadline"`
}

// CollectParams for collect operation
type CollectParams struct {
	TokenId    *big.Int       `json:"tokenId"`
	Recipient  common.Address `json:"recipient"`
	Amount0Max *big.Int       `json:"amount0Max"` // uint128
	Amount1Max *big.Int       `json:"amount1Max"` // uint128
}
