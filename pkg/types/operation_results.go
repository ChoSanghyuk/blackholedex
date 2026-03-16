package types

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

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

// UnstakeResult represents the complete output of unstake operation
type UnstakeResult struct {
	NFTTokenID   *big.Int            // Unstaked NFT token ID
	Rewards      *RewardAmounts      // Rewards collected (nil if not collected)
	Transactions []TransactionRecord // All transactions executed
	TotalGasCost *big.Int            // Sum of all gas costs (wei)
	Success      bool                // Whether operation succeeded
	ErrorMessage string              // Error message if failed (empty if success)
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

// RewardAmounts tracks rewards collected during unstake operation
type RewardAmounts struct {
	Reward           *big.Int       `json:"reward"`           // Primary reward amount
	BonusReward      *big.Int       `json:"bonusReward"`      // Bonus reward amount
	RewardToken      common.Address `json:"rewardToken"`      // Primary reward token address
	BonusRewardToken common.Address `json:"bonusRewardToken"` // Bonus reward token address
}
