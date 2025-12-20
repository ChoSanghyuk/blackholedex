# Data Model: Unstake Liquidity from Blackhole DEX

**Feature**: 001-unstake
**Date**: 2025-12-19

## Overview

This document defines the data entities used in the unstake functionality. These entities represent inputs, intermediate queries, and outputs of the unstake workflow, enabling users to withdraw their staked liquidity positions from the FarmingCenter contract.

## Entity Definitions

### 1. IncentiveKey

**Purpose**: Identifies a specific farming incentive program that a liquidity position is participating in

**Attributes**:
| Attribute | Type | Description | Validation Rules |
|-----------|------|-------------|------------------|
| RewardToken | common.Address | Primary reward token address | Non-zero address |
| BonusRewardToken | common.Address | Secondary reward token address | Can be zero address if no bonus |
| Pool | common.Address | The Algebra pool address (WAVAX/USDC) | Must equal wavaxUsdcPair constant |
| Nonce | *big.Int | Incentive identifier/version number | >= 0 |

**Relationships**:
- Required parameter for exitFarming and collectRewards calls
- Can be queried from FarmingCenter.deposits(tokenId) and incentiveKeys(incentiveId)
- Must match the incentive the NFT was originally staked with

**State Transitions**: Immutable (retrieved from on-chain state)

**Go Representation**:
```go
type IncentiveKey struct {
    RewardToken      common.Address  // Primary reward token (e.g., BLACK token)
    BonusRewardToken common.Address  // Bonus reward token (0x0 if none)
    Pool             common.Address  // WAVAX/USDC pool address
    Nonce            *big.Int        // Incentive nonce/version
}
```

**Solidity Reference**:
```solidity
struct IncentiveKey {
  IERC20Minimal rewardToken;
  IERC20Minimal bonusRewardToken;
  IAlgebraPool pool;
  uint256 nonce;
}
```

---

### 2. FarmingStatus

**Purpose**: Represents the current farming status of an NFT position

**Attributes**:
| Attribute | Type | Description | Validation Rules |
|-----------|------|-------------|------------------|
| TokenID | *big.Int | NFT token ID | > 0 |
| IncentiveID | [32]byte | Hash of the IncentiveKey | Non-zero if staked, zero if not staked |
| IsStaked | bool | Whether NFT is currently in farming | Derived from IncentiveID != 0 |
| Owner | common.Address | Current owner of the NFT | Non-zero address |

**Relationships**:
- Queried from FarmingCenter.deposits(tokenId) → incentiveId
- Queried from NonfungiblePositionManager.ownerOf(tokenId) → owner
- Used for pre-flight validation before unstake

**State Transitions**:
```
Staked (IncentiveID != 0) → Unstaked (IncentiveID = 0) via exitFarming
```

**Go Representation**:
```go
type FarmingStatus struct {
    TokenID     *big.Int
    IncentiveID [32]byte
    IsStaked    bool
    Owner       common.Address
}
```

---

### 3. UnstakeRequest

**Purpose**: Encapsulates user's intent to unstake a liquidity position

**Attributes**:
| Attribute | Type | Description | Validation Rules |
|-----------|------|-------------|------------------|
| NFTTokenID | *big.Int | NFT position token ID to unstake | > 0, must be owned by user |
| IncentiveKey | *IncentiveKey | Farming incentive to exit from | Non-nil, pool must be WAVAX/USDC |
| CollectRewards | bool | Whether to collect rewards during unstake | N/A (user preference) |

**Relationships**:
- NFTTokenID must exist and be owned by the caller
- IncentiveKey must match the current farming incentive for the NFT

**State Transitions**: N/A (immutable input)

**Go Representation**:
```go
type UnstakeRequest struct {
    NFTTokenID     *big.Int
    IncentiveKey   *IncentiveKey
    CollectRewards bool
}
```

---

### 4. RewardAmounts

**Purpose**: Tracks rewards collected during unstake operation

**Attributes**:
| Attribute | Type | Description | Validation Rules |
|-----------|------|-------------|------------------|
| Reward | *big.Int | Primary reward token amount collected | >= 0 |
| BonusReward | *big.Int | Bonus reward token amount collected | >= 0 |
| RewardToken | common.Address | Address of primary reward token | Non-zero if Reward > 0 |
| BonusRewardToken | common.Address | Address of bonus reward token | Can be zero if no bonus |

**Relationships**:
- Returned by FarmingCenter.collectRewards(key, tokenId)
- Only populated if UnstakeRequest.CollectRewards = true
- Included in UnstakeResult for financial transparency

**State Transitions**: N/A (calculated once during unstake)

**Go Representation**:
```go
type RewardAmounts struct {
    Reward           *big.Int
    BonusReward      *big.Int
    RewardToken      common.Address
    BonusRewardToken common.Address
}
```

---

### 5. TransactionRecord

**Purpose**: Tracks individual transaction details for financial transparency (reused from existing types.go)

**Attributes**:
| Attribute | Type | Description | Validation Rules |
|-----------|------|-------------|------------------|
| TxHash | common.Hash | Transaction hash | Non-zero |
| GasUsed | uint64 | Gas consumed | > 0 |
| GasPrice | *big.Int | Effective gas price (wei) | > 0 |
| GasCost | *big.Int | Total gas cost (wei) = GasUsed * GasPrice | > 0 |
| Timestamp | time.Time | Transaction timestamp | Valid time |
| Operation | string | Operation type (e.g., "ExitFarming", "CollectRewards") | Non-empty |

**Relationships**:
- Multiple TransactionRecords comprise an UnstakeResult
- Each unstake may generate 1-2 transaction records (exitFarming, optionally collectRewards via multicall)

**State Transitions**: Immutable (created after transaction confirmation)

**Go Representation**: Existing in types.go:166-173

---

### 6. UnstakeResult

**Purpose**: Comprehensive output of the unstake operation, tracking all financial and operational data

**Attributes**:
| Attribute | Type | Description | Validation Rules |
|-----------|------|-------------|------------------|
| NFTTokenID | *big.Int | NFT position token ID that was unstaked | > 0 |
| Rewards | *RewardAmounts | Rewards collected (if any) | nil if CollectRewards=false |
| Transactions | []TransactionRecord | All transactions executed | At least 1 transaction |
| TotalGasCost | *big.Int | Sum of all gas costs (wei) | > 0 |
| Success | bool | Whether operation succeeded | true or false |
| ErrorMessage | string | Error message if failed | Empty if Success=true |

**Relationships**:
- Contains aggregated TransactionRecords for complete gas tracking
- Includes RewardAmounts if rewards were collected
- Returned to caller for financial reporting and verification

**State Transitions**: N/A (output entity)

**Go Representation**:
```go
type UnstakeResult struct {
    NFTTokenID   *big.Int            // Unstaked NFT token ID
    Rewards      *RewardAmounts      // Rewards collected (nil if not collected)
    Transactions []TransactionRecord // All transactions executed
    TotalGasCost *big.Int            // Sum of all gas costs (wei)
    Success      bool                // Operation success status
    ErrorMessage string              // Error description (empty if success)
}
```

**Alternative**: Could reuse existing `StakingResult` with adapted field semantics:
- ActualAmount0 → Rewards.Reward
- ActualAmount1 → Rewards.BonusReward
- FinalTickLower/Upper → 0 (not applicable)

**Decision**: Create dedicated UnstakeResult for clarity, avoiding semantic confusion with StakingResult fields.

---

## Data Flow

### Unstake Workflow

```
1. Input: UnstakeRequest (tokenID, incentiveKey, collectRewards)
   ↓
2. Query: FarmingStatus (verify ownership and staking status)
   ↓
3. Encode: multicall(bytes[]) with exitFarming + optional collectRewards
   ↓
4. Execute: Send multicall transaction to FarmingCenter
   ↓
5. Track: TransactionRecord (gas costs, tx hash, timestamp)
   ↓
6. Parse: RewardAmounts from transaction events (if collected)
   ↓
7. Output: UnstakeResult (success, rewards, gas costs)
```

### Validation Pipeline

```
UnstakeRequest
  → Validate NFTTokenID > 0
  → Validate IncentiveKey.Pool == wavaxUsdcPair
  → Query FarmingStatus.Owner == caller
  → Query FarmingStatus.IsStaked == true
  → Query FarmingStatus.IncentiveID matches IncentiveKey hash
  ↓ (if all pass)
Execute unstake
```

---

## Entity Relationships Diagram

```
UnstakeRequest ──────┐
    │                │
    ├─ NFTTokenID ───┼──→ FarmingStatus (query ownership, staking)
    │                │
    └─ IncentiveKey ─┴──→ FarmingCenter.exitFarming(key, tokenId)
                     └──→ FarmingCenter.collectRewards(key, tokenId)
                                    ↓
                              RewardAmounts
                                    ↓
                            UnstakeResult ← TransactionRecord(s)
```

---

## State Transition Diagram

### NFT Farming State

```
┌─────────────┐    Stake()     ┌─────────────┐
│   Unstaked  │───────────────→│   Staked    │
│ (wallet)    │                │ (farming)   │
└─────────────┘                └─────────────┘
       ↑                              │
       │         Unstake()            │
       └──────────────────────────────┘
```

### IncentiveID State

```
deposits[tokenId] = 0x00...00  ──Stake()──→  deposits[tokenId] = incentiveId (hash)
                                                        │
                                                        │ Unstake()
                                                        ↓
deposits[tokenId] = 0x00...00  ←───────────  deposits[tokenId] reset to zero
```

---

## Implementation Notes

### Type Reuse vs. New Types

**Existing Types (reused)**:
- `TransactionRecord` - Already defined in types.go:166-173
- `common.Address`, `common.Hash`, `*big.Int` - Standard go-ethereum types

**New Types (to be added to types.go)**:
- `IncentiveKey` - Required for FarmingCenter interaction
- `FarmingStatus` - Optional helper for validation
- `RewardAmounts` - Tracks collected rewards
- `UnstakeRequest` - Optional (parameters can be passed directly to Unstake function)
- `UnstakeResult` - Primary output structure

**Minimal Implementation**:
At minimum, add:
1. `IncentiveKey` (required for exitFarming call)
2. `UnstakeResult` (output structure for financial tracking)

Optional types like `FarmingStatus` and `RewardAmounts` can be inline structs or embedded in UnstakeResult.

---

## Constitutional Compliance

### Principle 1: Pool-Specific Scope
- ✅ IncentiveKey.Pool validated against WAVAX/USDC pool address
- ✅ No support for other pools introduced

### Principle 3: Financial Transparency
- ✅ TransactionRecord tracks all gas costs
- ✅ RewardAmounts tracks all collected incentives
- ✅ UnstakeResult provides complete financial picture
- ✅ All data structured for export/reporting

### Principle 5: Fail-Safe Operation
- ✅ FarmingStatus validation prevents invalid unstake attempts
- ✅ UnstakeResult.ErrorMessage captures all failure details
- ✅ Partial gas costs tracked even on failure

---

## References

- Existing types.go:1-187 for established patterns
- FarmingCenter.sol exitFarming signature
- IncentiveKey.sol for Solidity struct definition
- Liquidity staking data-model.md for TransactionRecord pattern
