# Data Model: Blackhole DEX Liquidity Repositioning Agent

**Date**: 2025-12-06
**Purpose**: Define data entities and their relationships for the liquidity repositioning agent

---

## Entity Overview

This document defines the core data entities used by the repositioning agent to track positions, pool states, repositioning events, and user configurations.

```
┌──────────────────┐      monitors      ┌──────────────┐
│  UserConfiguration├──────────────────>│   Position   │
└──────────────────┘                     └──────┬───────┘
                                               │
                                               │ queries
                                               │
                                               ▼
                                        ┌──────────────┐
                                        │  PoolState   │
                                        └──────────────┘
                                               │
                                               │ triggers
                                               │
                                               ▼
                                        ┌────────────────────┐
                                        │ RepositioningEvent │
                                        └────────────────────┘
```

---

## 1. Position

Represents a concentrated liquidity position on Blackhole DEX managed by the NonfungiblePositionManager contract.

### Attributes

| Field | Type | Description | Validation Rules |
|-------|------|-------------|------------------|
| `PositionID` | uint256 | NFT token ID representing the position | Must be > 0, unique per position |
| `Owner` | address | Wallet address that owns the position NFT | Must be valid Ethereum address |
| `Pool` | address | Address of the Algebra pool contract | Must be valid pool address from supported pairs |
| `Token0` | address | Lower-address token in the pair | Must be < Token1 (canonical ordering) |
| `Token1` | address | Higher-address token in the pair | Must be > Token0 (canonical ordering) |
| `TickLower` | int24 | Lower tick bound of the position range | Must be multiple of tick spacing (200), < TickUpper |
| `TickUpper` | int24 | Upper tick bound of the position range | Must be multiple of tick spacing (200), > TickLower |
| `Liquidity` | uint128 | Amount of liquidity in the position | >= 0, 0 indicates empty position |
| `Amount0` | uint256 | Current amount of token0 in position | >= 0, calculated value |
| `Amount1` | uint256 | Current amount of token1 in position | >= 0, calculated value |
| `UnclaimedFees0` | uint256 | Uncollected fees in token0 | >= 0 |
| `UnclaimedFees1` | uint256 | Uncollected fees in token1 | >= 0 |
| `TotalFeesEarned0` | uint256 | Lifetime fees earned in token0 | Monotonically increasing |
| `TotalFeesEarned1` | uint256 | Lifetime fees earned in token1 | Monotonically increasing |
| `CreatedAt` | timestamp | When the position was minted | Unix timestamp |
| `LastUpdated` | timestamp | Last time position data was refreshed | Unix timestamp, >= CreatedAt |
| `LastInRangeTime` | timestamp | Last time position was in active range | Unix timestamp, used for trigger logic |
| `Status` | enum | Current position status | One of: InRange, BelowRange, AboveRange |
| `TimeInRange` | duration | Total time position has been in-range | Cumulative, used for performance tracking |
| `TimeOutOfRange` | duration | Total time position has been out-of-range | Cumulative, triggers repositioning |

### Status Enum Values

```go
type PositionStatus int

const (
    InRange     PositionStatus = iota  // tickLower <= currentTick < tickUpper
    BelowRange                         // currentTick < tickLower
    AboveRange                         // currentTick >= tickUpper
)
```

### State Transitions

```
        price moves up
BelowRange ──────────────> InRange ──────────────> AboveRange
           <──────────────         <──────────────
         price moves down        price moves down
```

### Derivable Fields

These fields are calculated from other attributes:

- **CurrentValue** = `Amount0 * Price0 + Amount1 * Price1` (in USD or reference token)
- **FeeAPR** = `(TotalFeesEarned / CurrentValue) * (365 days / position age)`
- **TickDistance** = `min(abs(currentTick - TickLower), abs(TickUpper - currentTick))` (proximity to range boundary)

### Relationships

- **One-to-One** with PoolState: Each position queries exactly one pool
- **One-to-Many** with RepositioningEvent: A position can be repositioned multiple times
- **Many-to-One** with UserConfiguration: Multiple positions belong to one configuration

---

## 2. PoolState

Represents the current state of a concentrated liquidity pool on Blackhole DEX.

### Attributes

| Field | Type | Description | Validation Rules |
|-------|------|-------------|------------------|
| `PoolAddress` | address | Address of the Algebra pool contract | Must be valid pool contract |
| `Token0` | address | Lower-address token in the pair | Canonical ordering (Token0 < Token1) |
| `Token1` | address | Higher-address token in the pair | Canonical ordering (Token1 > Token0) |
| `SqrtPriceX96` | uint160 | Current price in sqrt(token1/token0) * 2^96 format | > 0, Q64.96 fixed point |
| `Tick` | int24 | Current active tick | Valid tick value for the pool |
| `CurrentPrice` | float64 | Human-readable price (token1/token0) | Derived from SqrtPriceX96 |
| `LastFee` | uint16 | Current swap fee in basis points | Typically 100-3000 (0.01%-0.3%) |
| `ActiveLiquidity` | uint128 | Total liquidity available at current tick | >= 0 |
| `NextTick` | int24 | Next initialized tick above current | > Tick |
| `PreviousTick` | int24 | Next initialized tick below current | < Tick |
| `Volume24h` | uint256 | Trading volume in last 24 hours (token1) | >= 0, optional for monitoring |
| `TVL` | uint256 | Total value locked in pool (USD) | >= 0, optional for monitoring |
| `LastUpdated` | timestamp | When this state was last queried | Unix timestamp |

### Derivable Fields

- **Price** = `(SqrtPriceX96 / 2^96)^2` (convert from Q64.96 format)
- **InversePrice** = `1 / CurrentPrice` (token0/token1 instead of token1/token0)
- **TickSpacing** = 200 (fixed for Blackhole DEX Algebra pools)

### State Updates

Pool state is **read-only** from the agent's perspective (queried via RPC, not modified). State changes occur through:
- User swaps (move price/tick)
- Liquidity adds/removes (change ActiveLiquidity)
- Protocol parameter updates (change LastFee)

### Relationships

- **One-to-Many** with Position: One pool can have multiple positions
- **One-to-Many** with RepositioningEvent: Pool state at time of repositioning is recorded

---

## 3. RepositioningEvent

Represents a completed or in-progress repositioning operation for a liquidity position.

### Attributes

| Field | Type | Description | Validation Rules |
|-------|------|-------------|------------------|
| `EventID` | string | Unique identifier for this event | UUID or sequential ID |
| `PositionID` | uint256 | NFT token ID of position being repositioned | References Position.PositionID |
| `TriggerTime` | timestamp | When repositioning was initiated | Unix timestamp |
| `TriggerReason` | enum | Why repositioning was triggered | One of: OutOfRange, NearBoundary, Manual |
| `OldTickLower` | int24 | Previous lower tick bound | Multiple of tick spacing |
| `OldTickUpper` | int24 | Previous upper tick bound | Multiple of tick spacing |
| `NewTickLower` | int24 | New lower tick bound | Multiple of tick spacing |
| `NewTickUpper` | int24 | New upper tick bound | Multiple of tick spacing |
| `PoolTickAtTrigger` | int24 | Pool's current tick when triggered | Snapshot for analysis |
| `OldLiquidity` | uint128 | Liquidity amount before repositioning | From old position |
| `NewLiquidity` | uint128 | Liquidity amount after repositioning | From new position (may differ due to swaps) |
| `SwapExecuted` | bool | Whether token rebalancing swap occurred | true if token ratio needed adjustment |
| `SwapDetails` | struct | Swap parameters if executed | See SwapDetails sub-entity |
| `UnstakeTxHash` | bytes32 | Transaction hash for decreaseLiquidity call | Ethereum transaction hash |
| `CollectTxHash` | bytes32 | Transaction hash for collect call | Ethereum transaction hash |
| `SwapTxHash` | bytes32 | Transaction hash for swap (if executed) | Ethereum transaction hash or null |
| `MintTxHash` | bytes32 | Transaction hash for new position mint | Ethereum transaction hash |
| `GasCostTotal` | uint256 | Total gas spent across all transactions (wei) | Sum of all tx gas costs |
| `GasPriceAvg` | uint256 | Average gas price used (wei) | Weighted average across txs |
| `Outcome` | enum | Final status of repositioning | One of: Success, Partial, Failed |
| `ErrorDetails` | string | Error message if failed | Empty if Success, detailed if Failed/Partial |
| `CompletedAt` | timestamp | When repositioning finished (success or fail) | Unix timestamp, >= TriggerTime |
| `DurationSeconds` | uint32 | Time from trigger to completion | CompletedAt - TriggerTime |

### TriggerReason Enum

```go
type TriggerReason int

const (
    OutOfRange    TriggerReason = iota  // Position fully out of active range
    NearBoundary                        // Position within X ticks of boundary
    Manual                              // User manually triggered repositioning
)
```

### Outcome Enum

```go
type RepositioningOutcome int

const (
    Success  RepositioningOutcome = iota  // All steps completed successfully
    Partial                                // Some steps failed, position in inconsistent state
    Failed                                 // Entire operation failed, no changes made
)
```

### SwapDetails Sub-Entity

| Field | Type | Description |
|-------|------|-------------|
| `TokenIn` | address | Input token for swap |
| `TokenOut` | address | Output token for swap |
| `AmountIn` | uint256 | Input token amount |
| `AmountOut` | uint256 | Output token amount received |
| `AmountOutMin` | uint256 | Minimum output (slippage protection) |
| `SlippageActual` | float64 | Actual slippage experienced (%) |

### Relationships

- **Many-to-One** with Position: Multiple events can occur for same position over time
- **One-to-One** with PoolState: Each event snapshots pool state at trigger time

### Event Lifecycle

```
Triggered → Unstaking → Collecting → [Swapping] → Minting → Completed
                                         ↓
                                     (optional)
```

**State transitions:**
1. **Triggered**: Event created, trigger reason recorded
2. **Unstaking**: `decreaseLiquidity()` transaction submitted
3. **Collecting**: `collect()` transaction submitted to claim tokens/fees
4. **Swapping** (conditional): If token ratio needs adjustment, swap executed
5. **Minting**: New position `mint()` transaction submitted
6. **Completed**: All transactions confirmed, outcome set to Success/Partial/Failed

---

## 4. UserConfiguration

Represents agent settings and risk parameters for automated repositioning.

### Attributes

| Field | Type | Description | Validation Rules |
|-------|------|-------------|------------------|
| `ConfigID` | string | Unique identifier for this configuration | UUID or user wallet address |
| `Enabled` | bool | Whether automated repositioning is active | true/false |
| `WalletAddress` | address | User's wallet address | Valid Ethereum address |
| `MonitoredPositions` | []uint256 | List of position NFT IDs to monitor | Array of position IDs |
| `MonitoredPools` | []address | List of pool addresses to track | Array of pool addresses |
| `OutOfRangeDuration` | duration | Time threshold before triggering reposition | >= 0, default 1 hour |
| `PriceMovementThreshold` | float64 | Price change % to trigger repositioning | 0-100, e.g. 5.0 for 5% |
| `TickDistanceThreshold` | int32 | Min tick distance from boundary to trigger | >= 0, e.g. 10 ticks |
| `MaxSlippagePercent` | float64 | Maximum acceptable slippage for swaps | 0-100, default 0.5% |
| `MaxSlippageLiquidity` | float64 | Max slippage for liquidity operations | 0-100, default 1.0% |
| `MinPositionSize` | uint256 | Minimum position value to manage (USD) | >= 0, prevents tiny positions |
| `MaxGasPrice` | uint256 | Maximum gas price willing to pay (gwei) | > 0, prevents high-gas repositioning |
| `MultiPositionStrategy` | enum | How to handle multiple out-of-range positions | See MultiPositionStrategy enum |
| `NotificationPreferences` | struct | How to notify user of events | See NotificationPreferences sub-entity |
| `RPCEndpoint` | string | Avalanche RPC URL | Valid HTTP/WSS URL |
| `PrivateKeyPath` | string | Path to encrypted private key file | File path, never log this value |
| `CreatedAt` | timestamp | When configuration was created | Unix timestamp |
| `LastModified` | timestamp | When configuration was last updated | Unix timestamp, >= CreatedAt |

### MultiPositionStrategy Enum

```go
type MultiPositionStrategy int

const (
    LargestFirst    MultiPositionStrategy = iota  // Reposition highest-value positions first
    LongestOutFirst                               // Reposition positions out-of-range longest
    Sequential                                    // Process in order of position ID
    Parallel                                      // Handle all positions concurrently (future)
)
```

### NotificationPreferences Sub-Entity

| Field | Type | Description |
|-------|------|-------------|
| `EnableSlack` | bool | Send notifications to Slack webhook |
| `SlackWebhookURL` | string | Slack incoming webhook URL |
| `EnableEmail` | bool | Send email notifications |
| `EmailAddress` | string | Recipient email address |
| `NotifyOnTrigger` | bool | Notify when repositioning starts |
| `NotifyOnComplete` | bool | Notify when repositioning completes |
| `NotifyOnFailure` | bool | Notify when repositioning fails |

### Validation Rules

- **MaxSlippagePercent**: Must be between 0.01% and 10% (prevent excessive slippage or no protection)
- **OutOfRangeDuration**: Should be >= 5 minutes (prevent excessive rebalancing)
- **MaxGasPrice**: Should be reasonable for Avalanche (e.g., 25-100 gwei max)
- **MonitoredPositions**: All position IDs must exist and be owned by WalletAddress
- **PrivateKeyPath**: File must exist and be encrypted (never store plaintext keys)

### Relationships

- **One-to-Many** with Position: One config can manage multiple positions
- **One-to-Many** with RepositioningEvent: Configuration triggers multiple events over time

---

## Entity Relationships Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    UserConfiguration                         │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ ConfigID, Enabled, WalletAddress                        │ │
│ │ MonitoredPositions[], OutOfRangeDuration               │ │
│ │ MaxSlippagePercent, MaxGasPrice                        │ │
│ │ MultiPositionStrategy, NotificationPreferences         │ │
│ └─────────────────────────────────────────────────────────┘ │
└────────────────┬────────────────────────────────────────────┘
                 │ manages
                 ▼
┌─────────────────────────────────────────────────────────────┐
│                         Position                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ PositionID, Owner, Pool, Token0, Token1                │ │
│ │ TickLower, TickUpper, Liquidity                        │ │
│ │ Amount0, Amount1, UnclaimedFees0/1                     │ │
│ │ Status, TimeInRange, TimeOutOfRange                    │ │
│ └─────────────────────────────────────────────────────────┘ │
└────────┬────────────────────┬───────────────────────────────┘
         │ queries            │ triggers
         ▼                    ▼
┌──────────────────┐  ┌──────────────────────────────────────┐
│    PoolState     │  │      RepositioningEvent              │
│ ┌──────────────┐ │  │ ┌──────────────────────────────────┐ │
│ │ PoolAddress  │ │  │ │ EventID, PositionID              │ │
│ │ Token0/1     │ │  │ │ TriggerTime, TriggerReason       │ │
│ │ SqrtPriceX96 │ │  │ │ Old/NewTickLower, Old/NewTickUpper│ │
│ │ Tick         │ │  │ │ SwapDetails, TxHashes            │ │
│ │ LastFee      │ │  │ │ GasCostTotal, Outcome            │ │
│ │ ActiveLiq    │ │  │ └──────────────────────────────────┘ │
│ └──────────────┘ │  └──────────────────────────────────────┘
└──────────────────┘
```

---

## Data Storage Strategy

### In-Memory (During Runtime)

- **PoolState**: Cached for 30 seconds to reduce RPC calls
- **Position**: Cached for 60 seconds, refreshed before repositioning checks

### Persistent Storage (Config Files)

- **UserConfiguration**: Stored in `~/.blackhole-agent/config.yaml`
  - Never store private keys in plaintext
  - Use encrypted keystore files referenced by PrivateKeyPath

- **RepositioningEvent**: Append-only log file at `~/.blackhole-agent/events.jsonl`
  - One JSON object per line
  - Used for performance tracking and debugging

### Example Config File Structure

```yaml
# ~/.blackhole-agent/config.yaml
enabled: true
wallet_address: "0xb4dd4fb3d4bced984cce972991fb100488b59223"
monitored_positions:
  - 12345
  - 67890
monitored_pools:
  - "0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0"  # WAVAX/USDC
  - "0x14e4a5bed2e5e688ee1a5ca3a4914250d1abd573"  # WAVAX/BLACK
triggers:
  out_of_range_duration: 1h
  tick_distance_threshold: 10
risk_limits:
  max_slippage_percent: 0.5
  max_slippage_liquidity: 1.0
  max_gas_price_gwei: 50
  min_position_size_usd: 100
multi_position_strategy: largest_first
rpc_endpoint: "https://api.avax.network/ext/bc/C/rpc"
private_key_path: "~/.blackhole-agent/keystore/key.json"
```

### Example Event Log Entry

```json
{
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "position_id": 12345,
  "trigger_time": 1733500800,
  "trigger_reason": "OutOfRange",
  "old_tick_lower": -249600,
  "old_tick_upper": -248400,
  "new_tick_lower": -250000,
  "new_tick_upper": -248800,
  "pool_tick_at_trigger": -250500,
  "swap_executed": true,
  "swap_details": {
    "token_in": "0xb31f66aa3c1e785363f0875a1b74e27b85fd66c7",
    "token_out": "0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E",
    "amount_in": 1000000000000000000,
    "amount_out": 25500000,
    "slippage_actual": 0.3
  },
  "unstake_tx_hash": "0xabc123...",
  "collect_tx_hash": "0xdef456...",
  "swap_tx_hash": "0x789ghi...",
  "mint_tx_hash": "0x012jkl...",
  "gas_cost_total": 5000000000000000,
  "outcome": "Success",
  "completed_at": 1733500920,
  "duration_seconds": 120
}
```

---

## Data Flow

### Position Monitoring Cycle

```
1. Load UserConfiguration
    ↓
2. For each MonitoredPosition:
    ├─ Query Position data (batch RPC)
    ├─ Query PoolState data (batch RPC)
    └─ Calculate Status (InRange/OutOfRange)
    ↓
3. Evaluate Triggers:
    ├─ Check OutOfRangeDuration
    ├─ Check TickDistanceThreshold
    └─ Check GasCost vs FeeOpportunity
    ↓
4. If trigger met:
    ├─ Create RepositioningEvent
    └─ Execute repositioning workflow
    ↓
5. Update RepositioningEvent with outcome
    ↓
6. Save event to events.jsonl log
```

### Repositioning Workflow Data Flow

```
RepositioningEvent (Triggered)
    ↓
Query current Position + PoolState
    ↓
Calculate new TickLower/TickUpper
    ↓
Execute decreaseLiquidity → Update event.UnstakeTxHash
    ↓
Execute collect → Update event.CollectTxHash
    ↓
If rebalance needed:
    Execute swap → Update event.SwapDetails + SwapTxHash
    ↓
Calculate optimal amounts for new range
    ↓
Execute mint → Update event.MintTxHash
    ↓
Set event.Outcome = Success/Failed
Set event.CompletedAt = now
    ↓
Persist event to events.jsonl
```

---

## Performance Considerations

### Batch RPC Queries

To meet SC-001 (position status within 5 seconds), use batch queries:

```go
// Query 10 positions + 2 pools = 22 RPC calls
// Batch request reduces from 4.4s (sequential) to 200ms (batched)
batch := []rpc.BatchElem{
    {Method: "eth_call", Args: buildPositionQuery(positionID1), Result: &pos1},
    {Method: "eth_call", Args: buildPoolQuery(poolAddr1), Result: &pool1},
    // ... repeat for all positions/pools
}
client.BatchCallContext(ctx, batch)
```

### Caching Strategy

- **PoolState**: Cache for 30 seconds (pools change frequently due to swaps)
- **Position**: Cache for 60 seconds (positions only change on user actions)
- **UserConfiguration**: Cache for entire runtime (only reload on SIGHUP signal)

### Data Size Estimates

- **Position**: ~300 bytes per position
- **PoolState**: ~200 bytes per pool
- **RepositioningEvent**: ~500 bytes per event
- **UserConfiguration**: ~1 KB per config

For 50 positions monitored, total memory usage: ~15 KB (negligible)

---

## Summary

This data model provides:

1. **Complete position tracking** (ticks, liquidity, fees, status)
2. **Pool state monitoring** (price, tick, liquidity)
3. **Event audit trail** (all repositioning operations logged)
4. **User configuration** (risk limits, triggers, notifications)

All entities align with:
- ✅ Constitution Principle I (Go types: int24, uint256, address, etc.)
- ✅ Constitution Principle III (Clear status tracking for UX)
- ✅ Constitution Principle IV (Batch queries for performance)
- ✅ Constitution Principle V (Validation rules for security)
