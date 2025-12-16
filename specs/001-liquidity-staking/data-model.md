# Data Model: Liquidity Staking

**Feature**: 001-liquidity-staking
**Date**: 2025-12-09

## Overview

This document defines the data entities used in the liquidity staking feature. These entities represent inputs, intermediate calculations, and outputs of the staking workflow.

## Entity Definitions

### 1. StakingRequest

**Purpose**: Encapsulates operator's intent to stake liquidity

**Attributes**:
| Attribute | Type | Description | Validation Rules |
|-----------|------|-------------|------------------|
| MaxWAVAX | *big.Int | Maximum WAVAX amount to stake (wei) | > 0, <= wallet balance |
| MaxUSDC | *big.Int | Maximum USDC amount to stake (smallest unit) | > 0, <= wallet balance |
| RangeWidth | int | Number of tick ranges for position (e.g., 6 = ±3 ranges) | > 0, <= 20 |
| SlippagePct | int | Slippage tolerance percentage (e.g., 5 = 5%) | > 0, <= 50 |

**Relationships**: None (input entity)

**State Transitions**: N/A (immutable input)

**Go Representation**:
```go
type StakingRequest struct {
    MaxWAVAX    *big.Int
    MaxUSDC     *big.Int
    RangeWidth  int
    SlippagePct int
}
```

---

### 2. PoolState

**Purpose**: Represents current state of the WAVAX-USDC pool (from on-chain query)

**Attributes**:
| Attribute | Type | Description | Validation Rules |
|-----------|------|-------------|------------------|
| SqrtPrice | *big.Int | Current sqrt price in Q96 format | > 0 |
| Tick | int32 | Current tick | Within ±887272 |
| LastFee | uint16 | Last swap fee | N/A (informational) |
| PluginConfig | uint8 | Plugin configuration | N/A (informational) |
| ActiveLiquidity | *big.Int | Active liquidity in pool | >= 0 |
| NextTick | int32 | Next initialized tick | Within ±887272 |
| PreviousTick | int32 | Previous initialized tick | Within ±887272 |

**Relationships**: Used to calculate PositionBounds

**State Transitions**: Read-only (queried from chain)

**Go Representation**: Existing `AMMState` struct in types.go:150-160

---

### 3. PositionBounds

**Purpose**: Calculated tick range for liquidity position based on current pool state and range width

**Attributes**:
| Attribute | Type | Description | Validation Rules |
|-----------|------|-------------|------------------|
| TickLower | int32 | Lower tick bound | Within ±887272, < TickUpper |
| TickUpper | int32 | Upper tick bound | Within ±887272, > TickLower |
| TickSpacing | int | Tick spacing for pool (always 200) | == 200 |

**Relationships**:
- Derived from PoolState.Tick and StakingRequest.RangeWidth
- Used to calculate OptimalAmounts

**State Transitions**: N/A (calculated once)

**Calculation**:
```
tickIndex = currentTick / tickSpacing
tickLower = (tickIndex - rangeWidth/2) * tickSpacing
tickUpper = (tickIndex + rangeWidth/2) * tickSpacing
```

**Go Representation**:
```go
type PositionBounds struct {
    TickLower   int32
    TickUpper   int32
    TickSpacing int
}
```

---

### 4. OptimalAmounts

**Purpose**: Calculated token amounts needed for the specified position, with slippage protection

**Attributes**:
| Attribute | Type | Description | Validation Rules |
|-----------|------|-------------|------------------|
| Amount0Desired | *big.Int | WAVAX amount to stake (wei) | > 0, <= MaxWAVAX |
| Amount1Desired | *big.Int | USDC amount to stake (smallest unit) | > 0, <= MaxUSDC |
| Amount0Min | *big.Int | Minimum WAVAX after slippage | > 0, < Amount0Desired |
| Amount1Min | *big.Int | Minimum USDC after slippage | > 0, < Amount1Desired |
| Liquidity | *big.Int | Calculated liquidity value | > 0 |

**Relationships**:
- Calculated from PoolState, PositionBounds, StakingRequest
- Used to construct MintParams

**State Transitions**: N/A (calculated once)

**Calculation**:
```
// Using ComputeAmounts from internal/util/amm.go
amount0Desired, amount1Desired, liquidity = ComputeAmounts(
    sqrtPrice, currentTick, tickLower, tickUpper, maxWAVAX, maxUSDC
)

// Slippage protection
amount0Min = amount0Desired * (100 - slippagePct) / 100
amount1Min = amount1Desired * (100 - slippagePct) / 100
```

**Go Representation**:
```go
type OptimalAmounts struct {
    Amount0Desired *big.Int
    Amount1Desired *big.Int
    Amount0Min     *big.Int
    Amount1Min     *big.Int
    Liquidity      *big.Int
}
```

---

### 5. TransactionRecord

**Purpose**: Tracks individual transaction details for financial transparency

**Attributes**:
| Attribute | Type | Description | Validation Rules |
|-----------|------|-------------|------------------|
| TxHash | common.Hash | Transaction hash | Non-zero |
| GasUsed | uint64 | Gas consumed | > 0 |
| GasPrice | *big.Int | Effective gas price (wei) | > 0 |
| GasCost | *big.Int | Total gas cost (wei) = GasUsed * GasPrice | > 0 |
| Timestamp | time.Time | Transaction timestamp | Not future |
| Operation | string | Operation type (e.g., "ApproveWAVAX", "Mint") | Non-empty |

**Relationships**: Multiple TransactionRecords form StakingResult.Transactions

**State Transitions**:
1. Pending (submitted)
2. Confirmed (receipt received)
3. Failed (if reverted)

**Go Representation**:
```go
type TransactionRecord struct {
    TxHash    common.Hash
    GasUsed   uint64
    GasPrice  *big.Int
    GasCost   *big.Int
    Timestamp time.Time
    Operation string
}
```

---

### 6. StakingResult

**Purpose**: Complete output of staking operation including all transactions and final position

**Attributes**:
| Attribute | Type | Description | Validation Rules |
|-----------|------|-------------|------------------|
| NFTTokenID | *big.Int | Liquidity position NFT token ID | > 0 (after successful mint) |
| ActualAmount0 | *big.Int | Actual WAVAX staked (may differ from desired) | > 0, <= Amount0Desired |
| ActualAmount1 | *big.Int | Actual USDC staked (may differ from desired) | > 0, <= Amount1Desired |
| FinalTickLower | int32 | Final lower tick bound | == PositionBounds.TickLower |
| FinalTickUpper | int32 | Final upper tick bound | == PositionBounds.TickUpper |
| Transactions | []TransactionRecord | All transactions executed | Len >= 1 (at minimum mint) |
| TotalGasCost | *big.Int | Sum of all gas costs | >= 0 |
| Success | bool | Whether operation succeeded | true/false |
| ErrorMessage | string | Error message if failed | Empty if Success==true |

**Relationships**:
- Contains multiple TransactionRecords
- Final state of staking workflow

**State Transitions**: N/A (final output entity)

**Go Representation**:
```go
type StakingResult struct {
    NFTTokenID     *big.Int
    ActualAmount0  *big.Int
    ActualAmount1  *big.Int
    FinalTickLower int32
    FinalTickUpper int32
    Transactions   []TransactionRecord
    TotalGasCost   *big.Int
    Success        bool
    ErrorMessage   string
}
```

---

## Entity Relationship Diagram

```
┌─────────────────┐
│ StakingRequest  │ (Input)
│  - MaxWAVAX     │
│  - MaxUSDC      │
│  - RangeWidth   │
│  - SlippagePct  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   PoolState     │ (Query from chain)
│  - SqrtPrice    │
│  - Tick         │
│  - ...          │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ PositionBounds  │ (Calculated)
│  - TickLower    │
│  - TickUpper    │
│  - TickSpacing  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ OptimalAmounts  │ (Calculated)
│  - Amount0Des.  │
│  - Amount1Des.  │
│  - Amount0Min   │
│  - Amount1Min   │
│  - Liquidity    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐       ┌───────────────────┐
│   MintParams    │──────>│ TransactionRecord │ (Multiple)
│ (Existing type) │       │  - TxHash         │
│  - Token0       │       │  - GasCost        │
│  - Token1       │       │  - Operation      │
│  - TickLower    │       │  - ...            │
│  - TickUpper    │       └─────────┬─────────┘
│  - Amount0Des.  │                 │
│  - Amount1Des.  │                 │
│  - Amount0Min   │                 │
│  - Amount1Min   │                 │
│  - Recipient    │                 │
│  - Deadline     │                 │
└────────┬────────┘                 │
         │                          │
         ▼                          ▼
┌─────────────────────────────────────┐
│         StakingResult               │ (Output)
│  - NFTTokenID                       │
│  - ActualAmount0/1                  │
│  - FinalTickLower/Upper             │
│  - Transactions []TransactionRecord │
│  - TotalGasCost                     │
│  - Success                          │
│  - ErrorMessage                     │
└─────────────────────────────────────┘
```

## Data Flow

1. **Input**: Operator provides `StakingRequest`
2. **Query**: System retrieves `PoolState` from blockchain
3. **Calculate Bounds**: System computes `PositionBounds` from PoolState.Tick and RangeWidth
4. **Calculate Amounts**: System computes `OptimalAmounts` using ComputeAmounts utility
5. **Validate**: System validates balances and bounds
6. **Approve**: System ensures token approvals (records `TransactionRecord` for each)
7. **Mint**: System constructs `MintParams` and calls mint (records `TransactionRecord`)
8. **Output**: System returns `StakingResult` with all transaction details and final position

## Validation Rules Summary

| Entity | Validation Point | Rules |
|--------|------------------|-------|
| StakingRequest | Before workflow starts | Range width [1-20], Slippage [1-50], Amounts > 0 |
| PositionBounds | After calculation | Ticks within ±887272, TickLower < TickUpper |
| OptimalAmounts | Before balance check | Amounts > 0, <= Max amounts from request |
| MintParams | Before submission | All amounts > 0, Deadline in future |
| TransactionRecord | After receipt | GasUsed > 0, GasPrice > 0 |
| StakingResult | After completion | If Success, NFTTokenID > 0; if !Success, ErrorMessage non-empty |

## Implementation Notes

- Existing `MintParams` type (types.go:124-138) is reused for mint transaction
- `TransactionRecord` and `StakingResult` are new types needed for constitutional compliance (financial tracking)
- `StakingRequest`, `PositionBounds`, `OptimalAmounts` are intermediate types (may be implemented as local variables or structs depending on complexity)
- All big.Int operations must use pointer methods to avoid mutations
- Gas cost calculations must use uint64->big.Int conversions safely
