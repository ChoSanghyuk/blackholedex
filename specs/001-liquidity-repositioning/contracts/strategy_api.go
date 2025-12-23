// Package contracts defines the API contract for the automated liquidity repositioning strategy
// This file serves as the interface definition and type specifications for RunStrategy1
package contracts

import (
	"context"
	"math/big"
	"time"
)

// StrategyConfig defines configuration parameters for RunStrategy1 execution
// All fields are required except where noted as optional
type StrategyConfig struct {
	// MonitoringInterval specifies time between pool price checks
	// Default: 60 seconds (constitutional minimum)
	// Must be >= 1 minute
	MonitoringInterval time.Duration

	// StabilityThreshold defines maximum acceptable price change percentage
	// to consider price "stable" (e.g., 0.005 = 0.5%)
	// Default: 0.005 (0.5%)
	// Must be > 0 and < 0.1
	StabilityThreshold float64

	// StabilityIntervals specifies number of consecutive stable intervals
	// required before re-entering position after withdrawal
	// Default: 5
	// Must be >= 3
	StabilityIntervals int

	// RangeWidth defines position tick width (e.g., 10 = ±5 ticks from center)
	// Default: 10
	// Must be even and > 0
	RangeWidth int

	// SlippagePct defines slippage tolerance percentage for swaps and mints
	// Default: 1 (1%)
	// Must be > 0 and <= 5
	SlippagePct int

	// MaxWAVAX specifies maximum WAVAX amount (in wei) to use for liquidity
	// Required: Yes
	// Must be > 0 and <= wallet WAVAX balance
	MaxWAVAX *big.Int

	// MaxUSDC specifies maximum USDC amount (in smallest unit) to use for liquidity
	// Required: Yes
	// Must be > 0 and <= wallet USDC balance
	MaxUSDC *big.Int

	// CircuitBreakerWindow defines time window for error accumulation
	// Default: 5 minutes
	// Must be > 0
	CircuitBreakerWindow time.Duration

	// CircuitBreakerThreshold defines maximum errors allowed in window before halting
	// Default: 5
	// Must be >= 3
	CircuitBreakerThreshold int
}

// DefaultStrategyConfig returns a StrategyConfig with constitutional defaults
// User must still provide MaxWAVAX and MaxUSDC
func DefaultStrategyConfig() *StrategyConfig {
	return &StrategyConfig{
		MonitoringInterval:      60 * time.Second,
		StabilityThreshold:      0.005, // 0.5%
		StabilityIntervals:      5,
		RangeWidth:              10,
		SlippagePct:             1,
		CircuitBreakerWindow:    5 * time.Minute,
		CircuitBreakerThreshold: 5,
		// MaxWAVAX and MaxUSDC must be set by user
	}
}

// Validate checks StrategyConfig for validity
// Returns error if any field violates constraints
func (sc *StrategyConfig) Validate() error {
	// Implementation to be added in tasks phase
	return nil
}

// StrategyPhase represents the current execution phase of RunStrategy1
type StrategyPhase int

const (
	// Initializing: Initial setup, validating balances, creating first position
	Initializing StrategyPhase = iota

	// ActiveMonitoring: Monitoring pool price, position is staked and active
	ActiveMonitoring

	// RebalancingRequired: Out-of-range condition detected, preparing to rebalance
	RebalancingRequired

	// WaitingForStability: Position withdrawn, waiting for price stabilization
	WaitingForStability

	// ExecutingRebalancing: Performing token rebalancing and creating new position
	ExecutingRebalancing

	// Halted: Strategy stopped due to error or shutdown signal
	Halted
)

// String returns human-readable phase name
func (sp StrategyPhase) String() string {
	return [...]string{
		"Initializing",
		"ActiveMonitoring",
		"RebalancingRequired",
		"WaitingForStability",
		"ExecutingRebalancing",
		"Halted",
	}[sp]
}

// StrategyReport represents a structured message sent via the reporting channel
// All reports are JSON-serializable for flexible consumption
type StrategyReport struct {
	// Timestamp when this report was generated
	Timestamp time.Time `json:"timestamp"`

	// EventType categorizes the report (see Event Types documentation)
	// Examples: "strategy_start", "out_of_range", "gas_cost", "profit", "error", "halt"
	EventType string `json:"event_type"`

	// Message provides human-readable description of the event
	Message string `json:"message"`

	// Phase indicates current strategy phase (optional, included in most reports)
	Phase *StrategyPhase `json:"phase,omitempty"`

	// GasCost is the gas cost for this specific event in wei (optional)
	GasCost *big.Int `json:"gas_cost,omitempty"`

	// CumulativeGas is the total gas spent since strategy start in wei (optional)
	CumulativeGas *big.Int `json:"cumulative_gas,omitempty"`

	// Profit is the profit/reward amount for this event (optional, in BLACK tokens)
	Profit *big.Int `json:"profit,omitempty"`

	// NetPnL is the net profit/loss: (rewards - gas - fees) in wei (optional)
	NetPnL *big.Int `json:"net_pnl,omitempty"`

	// Error contains error message if EventType is "error" (optional)
	Error string `json:"error,omitempty"`

	// NFTTokenID is the current or affected position NFT ID (optional)
	NFTTokenID *big.Int `json:"nft_token_id,omitempty"`

	// PositionDetails provides snapshot of position state (optional)
	PositionDetails *PositionSnapshot `json:"position_details,omitempty"`
}

// ToJSON serializes StrategyReport to JSON string
func (sr *StrategyReport) ToJSON() (string, error) {
	// Implementation to be added in tasks phase
	return "", nil
}

// PositionSnapshot captures position details at a point in time
type PositionSnapshot struct {
	// NFTTokenID is the position NFT ID
	NFTTokenID *big.Int `json:"nft_token_id"`

	// TickLower is the lower tick bound (inclusive)
	TickLower int32 `json:"tick_lower"`

	// TickUpper is the upper tick bound (inclusive)
	TickUpper int32 `json:"tick_upper"`

	// Liquidity is the liquidity amount (uint128)
	Liquidity *big.Int `json:"liquidity"`

	// Amount0 is the WAVAX amount in position (wei)
	Amount0 *big.Int `json:"amount0"`

	// Amount1 is the USDC amount in position (smallest unit)
	Amount1 *big.Int `json:"amount1"`

	// FeeGrowth0 is accumulated fee growth for token0
	FeeGrowth0 *big.Int `json:"fee_growth0"`

	// FeeGrowth1 is accumulated fee growth for token1
	FeeGrowth1 *big.Int `json:"fee_growth1"`

	// Timestamp when this snapshot was taken
	Timestamp time.Time `json:"timestamp"`
}

// PositionRange encapsulates concentrated liquidity position tick bounds
type PositionRange struct {
	// TickLower is the lower tick bound (inclusive)
	// Must be less than TickUpper and divisible by tickSpacing (200)
	TickLower int32

	// TickUpper is the upper tick bound (inclusive)
	// Must be greater than TickLower and divisible by tickSpacing (200)
	TickUpper int32
}

// IsOutOfRange checks if current pool tick is outside this position's active range
// Returns true if currentTick < TickLower OR currentTick > TickUpper
func (pr *PositionRange) IsOutOfRange(currentTick int32) bool {
	return currentTick < pr.TickLower || currentTick > pr.TickUpper
}

// Width returns the tick width of this range
func (pr *PositionRange) Width() int32 {
	return pr.TickUpper - pr.TickLower
}

// Center returns the center tick of this range
func (pr *PositionRange) Center() int32 {
	return (pr.TickLower + pr.TickUpper) / 2
}

// StabilityWindow implements the price stability detection algorithm
// Used during WaitingForStability phase to determine when to re-enter position
type StabilityWindow struct {
	// Threshold is the maximum acceptable price change (0.005 = 0.5%)
	Threshold float64

	// RequiredIntervals is the number of consecutive stable intervals needed
	RequiredIntervals int

	// LastPrice is the previous interval's price (sqrtPrice from AMMState)
	LastPrice *big.Int

	// StableCount is the current count of consecutive stable intervals
	StableCount int
}

// CheckStability evaluates whether current price is stable
// Returns true if price has been stable for RequiredIntervals consecutive checks
// Resets counter if price change exceeds Threshold
func (sw *StabilityWindow) CheckStability(currentPrice *big.Int) bool {
	// Implementation to be added in tasks phase
	return false
}

// Reset clears the stability window state
func (sw *StabilityWindow) Reset() {
	sw.LastPrice = nil
	sw.StableCount = 0
}

// Progress returns stability progress as a fraction (0.0 to 1.0)
// Example: 3 stable intervals out of 5 required = 0.6
func (sw *StabilityWindow) Progress() float64 {
	if sw.RequiredIntervals == 0 {
		return 0.0
	}
	progress := float64(sw.StableCount) / float64(sw.RequiredIntervals)
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// CircuitBreaker tracks errors and determines when to halt the strategy
// Implements fail-safe error handling per Constitutional Principle 5
type CircuitBreaker struct {
	// ErrorWindow is the time window for error counting (e.g., 5 minutes)
	ErrorWindow time.Duration

	// ErrorThreshold is the maximum errors allowed in window before halting
	ErrorThreshold int

	// LastErrors stores timestamps of recent errors within the window
	LastErrors []time.Time

	// CriticalErrorOccurred indicates a critical error has happened (immediate halt)
	CriticalErrorOccurred bool
}

// RecordError records an error occurrence and determines if halt is required
// critical=true causes immediate halt, false uses threshold-based logic
// Returns true if strategy should halt, false if it can continue
func (cb *CircuitBreaker) RecordError(err error, critical bool) bool {
	// Implementation to be added in tasks phase
	return false
}

// Reset clears the circuit breaker state
func (cb *CircuitBreaker) Reset() {
	cb.LastErrors = []time.Time{}
	cb.CriticalErrorOccurred = false
}

// ErrorRate returns current error rate (errors per hour)
func (cb *CircuitBreaker) ErrorRate() float64 {
	if len(cb.LastErrors) == 0 {
		return 0.0
	}
	hoursInWindow := cb.ErrorWindow.Hours()
	return float64(len(cb.LastErrors)) / hoursInWindow
}

// StrategyRunner defines the interface for running the strategy
// This interface is implemented by the Blackhole struct
type StrategyRunner interface {
	// RunStrategy1 executes the automated liquidity repositioning strategy
	//
	// Parameters:
	//   ctx: Context for cancellation and timeout control
	//   reportChan: Channel for sending StrategyReport messages (JSON strings)
	//   config: Strategy configuration parameters
	//
	// Returns:
	//   error: nil on graceful shutdown, non-nil on fatal error
	//
	// Behavior:
	//   - Continuously monitors WAVAX/USDC pool price at config.MonitoringInterval
	//   - Automatically rebalances position when price exits active range
	//   - Waits for price stability before re-entering after withdrawal
	//   - Sends JSON reports to reportChan for all significant events
	//   - Halts on circuit breaker trigger or context cancellation
	//   - Respects all 5 constitutional principles
	//
	// Usage Example:
	//   config := DefaultStrategyConfig()
	//   config.MaxWAVAX = big.NewInt(1000000000000000000) // 1 WAVAX
	//   config.MaxUSDC = big.NewInt(10000000)             // 10 USDC
	//
	//   ctx, cancel := context.WithCancel(context.Background())
	//   defer cancel()
	//
	//   reportChan := make(chan string, 100)
	//   go func() {
	//       for report := range reportChan {
	//           fmt.Println(report)
	//       }
	//   }()
	//
	//   err := blackhole.RunStrategy1(ctx, reportChan, config)
	//   if err != nil {
	//       log.Fatalf("Strategy error: %v", err)
	//   }
	RunStrategy1(ctx context.Context, reportChan chan<- string, config *StrategyConfig) error
}

// Event Types Reference
//
// The following event types are sent via StrategyReport:
//
// 1. "strategy_start" - RunStrategy1 begins execution
//    Fields: Timestamp, Message, Phase
//
// 2. "monitoring" - Periodic price check (optional, can be verbose)
//    Fields: Timestamp, Message, Phase, PositionDetails
//
// 3. "out_of_range" - Current price has exited active range
//    Fields: Timestamp, Message, Phase, NFTTokenID
//
// 4. "rebalance_start" - Beginning withdrawal and rebalancing workflow
//    Fields: Timestamp, Message, Phase, NFTTokenID
//
// 5. "gas_cost" - Transaction gas cost incurred
//    Fields: Timestamp, Message, GasCost, CumulativeGas
//
// 6. "swap_complete" - Token rebalancing swap completed
//    Fields: Timestamp, Message, GasCost, CumulativeGas
//
// 7. "position_created" - New liquidity position minted and staked
//    Fields: Timestamp, Message, Phase, NFTTokenID, PositionDetails, GasCost
//
// 8. "profit" - Rewards collected from unstaking
//    Fields: Timestamp, Message, Profit, NetPnL, CumulativeGas
//
// 9. "stability_check" - Price stability check result
//    Fields: Timestamp, Message, Phase
//
// 10. "error" - Error occurred (may or may not halt strategy)
//     Fields: Timestamp, Message, Error, Phase
//
// 11. "halt" - Circuit breaker triggered, strategy halting
//     Fields: Timestamp, Message, Error, NetPnL, CumulativeGas
//
// 12. "shutdown" - Graceful shutdown via context cancellation
//     Fields: Timestamp, Message, NetPnL, CumulativeGas

// Nonce Management
//
// IncentiveKey.Nonce identifies the farming program for unstaking
// For WAVAX/USDC pool on Blackhole DEX:
//   - Query current nonce via FarmingCenter.incentiveKeys() function
//   - Nonce is typically a sequential identifier (e.g., 3, 4, 5...)
//   - Must match the nonce used when staking to successfully unstake
//
// Example:
//   farmingCenterClient.Call(nil, "incentiveKeys", incentiveId)
//   // Returns: rewardToken, bonusRewardToken, pool, nonce

// Gas Cost Tracking
//
// All operations track gas costs via TransactionRecord:
//   - Approve operations (WAVAX, USDC, NFT)
//   - Swap transactions (rebalancing)
//   - Mint transactions (create liquidity position)
//   - Stake transactions (deposit NFT to gauge)
//   - Unstake transactions (exit farming + collect rewards)
//   - Withdraw transactions (decrease liquidity + collect + burn NFT)
//
// CumulativeGas accumulates all gas costs for net P&L calculation:
//   NetPnL = TotalRewards - CumulativeGas - TotalSwapFees

// Price Conversion Utilities
//
// SqrtPrice (from AMMState) to human-readable price:
//   price = (sqrtPrice / 2^96)^2
//
// For WAVAX/USDC pool (assuming token0=WAVAX, token1=USDC):
//   priceUSDCperWAVAX = price * (10^(18-6)) // Adjust for decimals
//
// Tick to price conversion:
//   price = 1.0001^tick
//
// These conversions are used for stability detection and ratio calculations

// Additional Notes
//
// 1. Channel Buffering:
//    Recommend reportChan buffer size >= 100 to prevent blocking
//    during high-frequency events (rebalancing workflow)
//
// 2. Graceful Shutdown:
//    Cancel context to trigger graceful shutdown
//    Strategy checks context before starting new operations
//    Funds are never left in intermediate state
//
// 3. Error Classification:
//    Critical errors (immediate halt):
//      - "insufficient balance"
//      - "NFT not owned"
//      - "transaction reverted"
//      - "invalid position state"
//    Non-critical errors (threshold-based halt):
//      - RPC timeouts
//      - Gas estimation failures
//      - Temporary network issues
//
// 4. Testing Recommendations:
//    - Use mock RPC for integration tests
//    - Simulate price movements via GetAMMState mocks
//    - Test context cancellation at each phase
//    - Verify circuit breaker thresholds
//    - Validate JSON serialization of all reports
