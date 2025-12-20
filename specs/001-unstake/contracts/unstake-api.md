# Unstake API Contract

**Feature**: 001-unstake
**Date**: 2025-12-19
**Version**: 1.0

## Overview

This document defines the API contract for the Unstake functionality in the Blackhole DEX Go client. The API provides a programmatic interface for withdrawing staked liquidity positions from the FarmingCenter contract.

---

## Function: Unstake

### Signature

```go
func (b *Blackhole) Unstake(
    nftTokenID *big.Int,
    incentiveKey *IncentiveKey,
    collectRewards bool,
) (*UnstakeResult, error)
```

### Description

Withdraws a staked liquidity position NFT from the FarmingCenter contract. Optionally collects accumulated farming rewards in the same transaction using multicall for gas efficiency.

### Parameters

| Parameter | Type | Required | Description | Validation |
|-----------|------|----------|-------------|------------|
| `nftTokenID` | `*big.Int` | Yes | ERC721 token ID of the staked position | Must be > 0, must be owned by caller, must be currently staked |
| `incentiveKey` | `*IncentiveKey` | Yes | Farming incentive identifier to exit from | Pool must equal WAVAX/USDC pair, nonce must match deposited incentive |
| `collectRewards` | `bool` | Yes | Whether to collect rewards during unstake | No validation (user preference) |

### IncentiveKey Structure

```go
type IncentiveKey struct {
    RewardToken      common.Address  // Primary reward token address
    BonusRewardToken common.Address  // Bonus reward token address (0x0 if none)
    Pool             common.Address  // Must be WAVAX/USDC pool (0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0)
    Nonce            *big.Int        // Incentive nonce
}
```

### Returns

#### Success: `(*UnstakeResult, nil)`

```go
type UnstakeResult struct {
    NFTTokenID   *big.Int            // Unstaked NFT token ID
    Rewards      *RewardAmounts      // Rewards collected (nil if collectRewards=false)
    Transactions []TransactionRecord // Gas tracking for all transactions
    TotalGasCost *big.Int            // Total gas cost in wei
    Success      bool                // true
    ErrorMessage string              // "" (empty)
}

type RewardAmounts struct {
    Reward           *big.Int        // Primary reward amount (if collected)
    BonusReward      *big.Int        // Bonus reward amount (if collected)
    RewardToken      common.Address  // Primary reward token address
    BonusRewardToken common.Address  // Bonus reward token address
}
```

#### Error: `(*UnstakeResult, error)`

When an error occurs:
- `*UnstakeResult`: Partial result with available information
  - `Success`: `false`
  - `ErrorMessage`: Detailed error description
  - `Transactions`: May contain gas records for partial execution
  - `TotalGasCost`: Sum of gas costs for completed transactions (if any)
- `error`: Go error with context

### Behavior

1. **Pre-flight Validation**:
   - Validates `nftTokenID` is positive
   - Validates `incentiveKey.Pool` matches WAVAX/USDC pair (constitutional constraint)
   - Verifies NFT ownership via `NonfungiblePositionManager.ownerOf(tokenId)`
   - Verifies NFT is currently staked via `FarmingCenter.deposits(tokenId)`
   - Verifies `incentiveKey` matches deposited incentive

2. **Multicall Construction**:
   - Encodes `exitFarming(incentiveKey, nftTokenID)` call
   - If `collectRewards == true`:
     - Encodes `collectRewards(incentiveKey, nftTokenID)` call
     - Encodes `claimReward(rewardToken, recipient, amount)` calls for both reward tokens
   - Combines all calls into `bytes[]` array

3. **Transaction Execution**:
   - Sends `multicall(bytes[])` transaction to FarmingCenter
   - Uses automatic gas estimation (no manual gas limit)
   - Waits for transaction confirmation via `TxListener`

4. **Result Processing**:
   - Extracts gas cost from transaction receipt
   - Parses multicall results to extract reward amounts (if applicable)
   - Constructs and returns `UnstakeResult`

### Error Conditions

| Error | Condition | ErrorMessage Example |
|-------|-----------|---------------------|
| Invalid Token ID | `nftTokenID <= 0` or `nil` | "validation failed: invalid token ID" |
| Wrong Pool | `incentiveKey.Pool != wavaxUsdcPair` | "pool must be WAVAX/USDC pair: 0xA02..." |
| Not Owner | Caller doesn't own NFT | "NFT not owned by wallet: owned by 0x..." |
| Not Staked | NFT not in farming | "NFT is not currently staked in farming" |
| Incentive Mismatch | Provided key doesn't match deposits | "Invalid incentiveId" (contract revert) |
| Transaction Failed | On-chain revert | "deposit transaction failed: <revert reason>" |
| Network Error | RPC timeout/failure | "failed to verify NFT ownership: <network error>" |

### Example Usage

#### Basic Unstake (No Reward Collection)

```go
import (
    "math/big"
    "github.com/ethereum/go-ethereum/common"
)

tokenID := big.NewInt(12345)

incentiveKey := &IncentiveKey{
    RewardToken:      common.HexToAddress("0xcd94a87696fac69edae3a70fe5725307ae1c43f6"), // BLACK
    BonusRewardToken: common.Address{}, // No bonus
    Pool:             common.HexToAddress("0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0"), // WAVAX/USDC
    Nonce:            big.NewInt(1),
}

result, err := blackhole.Unstake(tokenID, incentiveKey, false)
if err != nil {
    log.Fatalf("Unstake failed: %v", err)
}

if !result.Success {
    log.Fatalf("Unstake unsuccessful: %s", result.ErrorMessage)
}

fmt.Printf("Unstaked NFT #%s\n", result.NFTTokenID.String())
fmt.Printf("Gas cost: %s wei\n", result.TotalGasCost.String())
```

#### Unstake With Reward Collection

```go
result, err := blackhole.Unstake(tokenID, incentiveKey, true)
if err != nil {
    log.Fatalf("Unstake failed: %v", err)
}

if result.Rewards != nil {
    fmt.Printf("Collected rewards:\n")
    fmt.Printf("  Primary: %s (token: %s)\n",
        result.Rewards.Reward.String(),
        result.Rewards.RewardToken.Hex())
    if result.Rewards.BonusReward.Sign() > 0 {
        fmt.Printf("  Bonus: %s (token: %s)\n",
            result.Rewards.BonusReward.String(),
            result.Rewards.BonusRewardToken.Hex())
    }
}
```

---

## Constitutional Compliance

### Principle 1: Pool-Specific Scope
-  Function validates `incentiveKey.Pool` matches WAVAX/USDC pool constant
-  Rejects any attempt to unstake from other pools with clear error message

### Principle 3: Financial Transparency
-  `TransactionRecord` in result tracks gas used, gas price, total gas cost
-  `RewardAmounts` tracks all collected incentives with token addresses
-  All financial data exportable via JSON marshaling of `UnstakeResult`

### Principle 4: Gas Optimization
-  Uses multicall to batch exitFarming + collectRewards in single transaction
-  Automatic gas estimation prevents over-estimation
-  Pre-flight validation prevents wasted gas on doomed transactions

### Principle 5: Fail-Safe Operation
-  Comprehensive pre-flight checks before transaction submission
-  All errors captured in `UnstakeResult.ErrorMessage` and returned `error`
-  Partial gas costs tracked even on failure
-  Transaction confirmation via `TxListener` with timeout/retry logic
-  FarmingCenter.exitFarming is atomic (on-chain revert on any failure)

---

## Testing Contract

### Unit Test Cases

```go
func TestUnstake_Success(t *testing.T) {
    // Test successful unstake without reward collection
}

func TestUnstake_WithRewards(t *testing.T) {
    // Test successful unstake with reward collection
}

func TestUnstake_InvalidTokenID(t *testing.T) {
    // Test rejection of nil/zero/negative token ID
}

func TestUnstake_WrongPool(t *testing.T) {
    // Test rejection of incentiveKey with non-WAVAX/USDC pool
}

func TestUnstake_NotOwner(t *testing.T) {
    // Test rejection when caller doesn't own NFT
}

func TestUnstake_NotStaked(t *testing.T) {
    // Test rejection when NFT is not in farming
}

func TestUnstake_IncentiveMismatch(t *testing.T) {
    // Test rejection when incentiveKey doesn't match deposits
}

func TestUnstake_NetworkError(t *testing.T) {
    // Test error handling when RPC calls fail
}

func TestUnstake_TransactionRevert(t *testing.T) {
    // Test error handling when on-chain transaction reverts
}

func TestUnstake_GasTracking(t *testing.T) {
    // Test that gas costs are accurately recorded
}
```

### Integration Test Requirements

- Real mainnet transaction execution with testnet funds
- Validation against actual FarmingCenter contract
- Gas cost benchmarking (must be < 150% of direct contract call)
- Reward collection accuracy verification

---

## Versioning

**Version 1.0** (2025-12-19)
- Initial API definition
- Support for exitFarming via multicall
- Optional reward collection
- WAVAX/USDC pool scope constraint

---

## Related Documentation

- [data-model.md](../data-model.md) - Entity definitions (IncentiveKey, UnstakeResult)
- [research.md](../research.md) - FarmingCenter contract analysis
- [quickstart.md](../quickstart.md) - Step-by-step usage guide
- [spec.md](../spec.md) - Feature requirements and acceptance criteria
