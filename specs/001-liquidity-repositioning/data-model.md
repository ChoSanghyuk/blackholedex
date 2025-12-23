# Data Model: Automated Liquidity Repositioning Strategy

**Feature**: RunStrategy1
**Date**: 2025-12-23
**Source**: Extracted from [spec.md](spec.md) Key Entities section

---

## Overview

This document defines all data structures required for the automated liquidity repositioning strategy. These entities capture strategy state, financial tracking, and operational configuration.

---

## Entity Definitions

### 1. StrategyConfig

**Purpose**: Configuration parameters for RunStrategy1 execution

**Attributes**:
| Field | Type | Description | Default | Validation |
|-------|------|-------------|---------|------------|
| MonitoringInterval | time.Duration | Time between pool price checks | 60s | Must be >= 1 minute (constitutional minimum) |
| StabilityThreshold | float64 | Max price change % to consider stable | 0.005 (0.5%) | Must be > 0 and < 0.1 |
| StabilityIntervals | int | Consecutive stable intervals required | 5 | Must be >= 3 |
| RangeWidth | int | Position tick width (e.g., 10 = ±5 ticks) | 10 | Must be even, > 0 |
| SlippagePct | int | Slippage tolerance percentage | 1 | Must be > 0 and <= 5 |
| MaxWAVAX | *big.Int | Maximum WAVAX to use for liquidity | User-provided | Must be > 0 and <= wallet balance |
| MaxUSDC | *big.Int | Maximum USDC to use for liquidity | User-provided | Must be > 0 and <= wallet balance |
| CircuitBreakerWindow | time.Duration | Time window for error counting | 5 * time.Minute | Must be > 0 |
| CircuitBreakerThreshold | int | Max errors in window before halt | 5 | Must be >= 3 |

**Relationships**:
- Used by: RunStrategy1 method
- Contains: All tunable parameters for strategy behavior

**State Transitions**: Immutable once strategy starts (read-only during execution)

---

### 2. StrategyState

**Purpose**: Tracks the current operational state and position information during strategy execution

**Attributes**:
| Field | Type | Description | Initial Value |
|-------|------|-------------|---------------|
| CurrentState | StrategyPhase | Current phase of execution | Initializing |
| NFTTokenID | *big.Int | Active position NFT ID | nil |
| TickLower | int32 | Active position lower bound | 0 |
| TickUpper | int32 | Active position upper bound | 0 |
| LastPrice | *big.Int | Last observed pool price (sqrtPrice) | nil |
| StableCount | int | Consecutive stable intervals counted | 0 |
| CumulativeGas | *big.Int | Total gas spent (wei) | 0 |
| CumulativeRewards | *big.Int | Total rewards collected (BLACK tokens) | 0 |
| TotalSwapFees | *big.Int | Cumulative swap fees paid | 0 |
| ErrorCount | int | Errors in current circuit breaker window | 0 |
| LastErrorTime | time.Time | Timestamp of most recent error | zero value |
| StartTime | time.Time | Strategy start timestamp | time.Now() |
| PositionCreatedAt | time.Time | When current position was created | zero value |

**Relationships**:
- Maintained by: RunStrategy1 method
- Updated during: All phase transitions and financial events
- Read by: Monitoring loop, reporting logic

**State Transitions**:
```
Initializing → ActiveMonitoring
ActiveMonitoring → RebalancingRequired (when out of range detected)
RebalancingRequired → WaitingForStability (after withdrawal)
WaitingForStability → WaitingForStability (while unstable)
WaitingForStability → ExecutingRebalancing (when stable)
ExecutingRebalancing → ActiveMonitoring (after new position staked)
Any State → Halted (on circuit breaker trigger or shutdown)
```

---

### 3. StrategyPhase (Enum)

**Purpose**: Enumeration of strategy execution phases

**Values**:
| Value | Description |
|-------|-------------|
| Initializing | Initial setup, validating balances |
| ActiveMonitoring | Monitoring pool price, position is staked |
| RebalancingRequired | Out-of-range detected, preparing to rebalance |
| WaitingForStability | Position withdrawn, waiting for price stabilization |
| ExecutingRebalancing | Performing token rebalancing and creating new position |
| Halted | Strategy stopped (error or shutdown) |

**Implementation**:
```go
type StrategyPhase int

const (
    Initializing StrategyPhase = iota
    ActiveMonitoring
    RebalancingRequired
    WaitingForStability
    ExecutingRebalancing
    Halted
)

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
```

---

### 4. PositionRange

**Purpose**: Encapsulates concentrated liquidity position tick bounds

**Attributes**:
| Field | Type | Description |
|-------|------|-------------|
| TickLower | int32 | Lower tick bound (inclusive) |
| TickUpper | int32 | Upper tick bound (inclusive) |

**Methods**:
```go
func (pr *PositionRange) IsOutOfRange(currentTick int32) bool
func (pr *PositionRange) Width() int32
func (pr *PositionRange) Center() int32
```

**Relationships**:
- Derived from: Mint operation results
- Used by: Out-of-range detection logic
- Stored in: StrategyState (TickLower, TickUpper)

**Validation Rules**:
- TickLower < TickUpper
- Both must be multiples of tickSpacing (200 for WAVAX/USDC pool)

---

### 5. StabilityWindow

**Purpose**: Implements price stability detection algorithm

**Attributes**:
| Field | Type | Description |
|-------|------|-------------|
| Threshold | float64 | Maximum acceptable price change (0.005 = 0.5%) |
| RequiredIntervals | int | Consecutive stable intervals needed (5) |
| LastPrice | *big.Int | Previous interval's price |
| StableCount | int | Current consecutive stable count |
| PriceHistory | []*big.Int | Recent price samples (optional, for debugging) |

**Methods**:
```go
func (sw *StabilityWindow) CheckStability(currentPrice *big.Int) bool
func (sw *StabilityWindow) Reset()
func (sw *StabilityWindow) Progress() float64 // Returns stability progress (0.0 to 1.0)
```

**Relationships**:
- Used by: Rebalancing workflow (WaitingForStability phase)
- Updates: StrategyState.StableCount

**Lifecycle**: Reset whenever price exceeds threshold

---

### 6. StrategyReport

**Purpose**: Structured message sent via reporting channel

**Attributes**:
| Field | Type | Description | Required |
|-------|------|-------------|----------|
| Timestamp | time.Time | Report generation time | Yes |
| EventType | string | Event category (see Event Types below) | Yes |
| Message | string | Human-readable description | Yes |
| Phase | StrategyPhase | Current strategy phase | No |
| GasCost | *big.Int | Gas cost for this event (wei) | No |
| CumulativeGas | *big.Int | Total gas spent so far | No |
| Profit | *big.Int | Profit/reward amount (BLACK tokens) | No |
| NetPnL | *big.Int | Net P&L = rewards - gas - fees | No |
| Error | string | Error message if EventType="error" | No |
| NFTTokenID | *big.Int | Position NFT ID | No |
| PositionDetails | *PositionSnapshot | Current position info | No |

**Event Types**:
| EventType | When Sent | Required Fields |
|-----------|-----------|----------------|
| "strategy_start" | RunStrategy1 begins | Timestamp, Message |
| "monitoring" | Each price check (optional, can be noisy) | Timestamp, Message, Phase |
| "out_of_range" | Price exits active range | Timestamp, Message, NFTTokenID |
| "rebalance_start" | Beginning withdrawal | Timestamp, Message, Phase |
| "gas_cost" | After any transaction | Timestamp, GasCost, CumulativeGas, Message |
| "swap_complete" | After rebalancing swaps | Timestamp, GasCost, Message |
| "position_created" | After Mint+Stake | Timestamp, NFTTokenID, PositionDetails, GasCost |
| "profit" | After collecting rewards | Timestamp, Profit, NetPnL, Message |
| "stability_check" | During price stabilization | Timestamp, Message, Phase |
| "error" | Any error occurs | Timestamp, Error, Message, Phase |
| "halt" | Circuit breaker triggers | Timestamp, Message, Error |
| "shutdown" | Graceful shutdown | Timestamp, Message, NetPnL |

**JSON Serialization**:
```go
func (sr *StrategyReport) ToJSON() (string, error) {
    bytes, err := json.Marshal(sr)
    if err != nil {
        return "", err
    }
    return string(bytes), nil
}
```

**Example Output**:
```json
{
  "timestamp": "2025-12-23T10:30:00Z",
  "event_type": "position_created",
  "message": "New liquidity position created and staked",
  "phase": "ActiveMonitoring",
  "gas_cost": "150000000000000",
  "cumulative_gas": "450000000000000",
  "nft_token_id": "12345",
  "position_details": {
    "tick_lower": -1000,
    "tick_upper": 1000,
    "liquidity": "1000000000000000000"
  }
}
```

---

### 7. PositionSnapshot

**Purpose**: Captures position details at a point in time

**Attributes**:
| Field | Type | Description |
|-------|------|-------------|
| NFTTokenID | *big.Int | Position NFT ID |
| TickLower | int32 | Lower tick bound |
| TickUpper | int32 | Upper tick bound |
| Liquidity | *big.Int | Liquidity amount (uint128) |
| Amount0 | *big.Int | WAVAX amount in position |
| Amount1 | *big.Int | USDC amount in position |
| FeeGrowth0 | *big.Int | Accumulated fee growth (token0) |
| FeeGrowth1 | *big.Int | Accumulated fee growth (token1) |
| Timestamp | time.Time | When snapshot was taken |

**Relationships**:
- Derived from: NonfungiblePositionManager.positions() call
- Used in: StrategyReport.PositionDetails
- Stored: Not persisted, generated on-demand for reporting

---

### 8. CircuitBreaker

**Purpose**: Tracks errors and determines when to halt strategy

**Attributes**:
| Field | Type | Description |
|-------|------|-------------|
| ErrorWindow | time.Duration | Time window for error counting (5 minutes) |
| ErrorThreshold | int | Max errors in window (5) |
| LastErrors | []time.Time | Timestamps of recent errors |
| CriticalErrorOccurred | bool | Whether a critical error happened |

**Methods**:
```go
func (cb *CircuitBreaker) RecordError(err error, critical bool) bool
func (cb *CircuitBreaker) Reset()
func (cb *CircuitBreaker) ErrorRate() float64
```

**Relationships**:
- Maintained by: RunStrategy1 main loop
- Updated: On every error occurrence
- Triggers: Strategy halt when threshold exceeded

---

### 9. RebalanceWorkflow

**Purpose**: Tracks a complete rebalancing operation from start to finish

**Attributes**:
| Field | Type | Description |
|-------|------|-------------|
| StartTime | time.Time | Workflow initiation timestamp |
| OldPosition | *PositionSnapshot | Position before withdrawal |
| WithdrawResult | *WithdrawResult | Withdrawal operation result |
| SwapResults | []TransactionRecord | All swap transactions |
| MintResult | *StakingResult | New position creation result |
| StakeResult | *StakingResult | Staking operation result |
| TotalGas | *big.Int | Sum of gas costs for entire workflow |
| Duration | time.Duration | Time from start to completion |
| Success | bool | Whether workflow completed successfully |
| ErrorMessage | string | Error if failed |

**Lifecycle**:
```
1. Create RebalanceWorkflow instance
2. Populate OldPosition from current state
3. Execute Unstake → record results
4. Execute Withdraw → record WithdrawResult
5. Wait for stability
6. Execute Swap(s) → record SwapResults
7. Execute Mint → record MintResult
8. Execute Stake → record StakeResult
9. Calculate TotalGas and Duration
10. Send final report via channel
```

**Relationships**:
- Created during: RebalancingRequired phase
- Completed in: ExecutingRebalancing phase
- Reported via: StrategyReport with EventType="rebalance_complete"

---

## Existing Types (Reused from types.go)

The following types are already defined in `types.go` and will be reused:

| Type | Usage in RunStrategy1 |
|------|----------------------|
| TransactionRecord | Track all individual transactions (swap, approve, mint, stake) |
| StakingResult | Store results from Mint and Stake operations |
| UnstakeResult | Store results from Unstake operation |
| WithdrawResult | Store results from Withdraw operation |
| AMMState | Retrieve current pool price and tick from GetAMMState |
| Route | Construct swap routes (WAVAX↔USDC) |
| SWAPExactTokensForTokensParams | Parameters for Swap operations |
| MintParams | Parameters for Mint operations (used internally) |
| IncentiveKey | Identify farming incentive for Unstake |
| RewardAmounts | Track collected rewards from Unstake |

---

## Data Flow Diagram

```
RunStrategy1 Start
    ↓
[StrategyConfig] → StrategyState (Initializing)
    ↓
Check balances, create initial position
    ↓
StrategyState → ActiveMonitoring
    ↓
[Monitoring Loop]
    ↓
GetAMMState → AMMState (current price, tick)
    ↓
PositionRange.IsOutOfRange(currentTick)?
    ↓ YES
StrategyState → RebalancingRequired
    ↓
Unstake(NFTTokenID) → UnstakeResult
    ↓
Withdraw(NFTTokenID) → WithdrawResult
    ↓
StrategyState → WaitingForStability
    ↓
[Stability Loop]
    ↓
StabilityWindow.CheckStability(currentPrice)?
    ↓ YES
StrategyState → ExecutingRebalancing
    ↓
Calculate rebalance amounts
    ↓
Swap(WAVAX/USDC) → TransactionRecord
    ↓
Mint(rebalanced amounts) → StakingResult
    ↓
Stake(NFT) → StakingResult
    ↓
Update StrategyState with new position
    ↓
StrategyState → ActiveMonitoring
    ↓
[Repeat]
```

---

## Validation Rules Summary

| Entity | Key Validations |
|--------|----------------|
| StrategyConfig | MonitoringInterval >= 60s, StabilityThreshold in (0, 0.1), MaxWAVAX/MaxUSDC <= balances |
| PositionRange | TickLower < TickUpper, both divisible by tickSpacing (200) |
| StabilityWindow | Threshold > 0, RequiredIntervals >= 3 |
| CircuitBreaker | ErrorThreshold >= 3, ErrorWindow > 0 |
| StrategyReport | Timestamp always present, EventType from allowed list |

---

## Storage and Persistence

**Note**: RunStrategy1 does NOT persist data to local storage. All state is:
- **Ephemeral**: Exists only in memory during strategy execution
- **Recoverable**: Position state can be queried from blockchain via NonfungiblePositionManager.positions()
- **Reported**: Financial data sent via string channel for external persistence if needed

If persistence is required in future versions, consider:
- CSV/JSON file logging of StrategyReport events
- SQLite database for historical P&L tracking
- Prometheus metrics export for monitoring dashboards
