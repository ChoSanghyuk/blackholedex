# Research: Unstake Implementation

**Feature**: 001-unstake
**Created**: 2025-12-18
**Status**: Complete

## Overview

This document consolidates research findings for implementing the unstake functionality in the Blackhole DEX Go client. All technical decisions are based on analysis of the existing `Stake()` function in `blackhole.go`, the FarmingCenter contract interface, and constitutional requirements.

## Research Areas

### 1. FarmingCenter Contract Interface

**Question**: What is the exact interface for the FarmingCenter contract's unstake operation?

**Decision**: Use `exitFarming(IncentiveKey memory key, uint256 tokenId)` wrapped in `multicall(bytes[] data)`

**Rationale**:
- Analysis of FarmingCenter.sol (lines 58-67) reveals the `exitFarming` function as the primary unstake method
- The function signature: `exitFarming(IncentiveKey memory key, uint256 tokenId) external isApprovedOrOwner(tokenId)`
- The multicall interface (Multicall.sol lines 10-26) accepts `bytes[] calldata data` and executes via delegatecall
- This allows batching multiple operations atomically (e.g., exitFarming + collectRewards)

**Alternatives Considered**:
- **Direct exitFarming call**: Simpler but lacks batching capability, not aligned with user's requirement to use multicall
- **Multiple separate transactions**: Higher gas costs, not atomic, rejected for efficiency and safety reasons

**Implementation Details**:
```go
// IncentiveKey struct matches Solidity definition
type IncentiveKey struct {
    RewardToken common.Address  // Token used for rewards
    BonusRewardToken common.Address  // Bonus token (can be zero address)
    Pool common.Address  // WAVAX/USDC pool address
    Nonce *big.Int  // Incentive nonce/identifier
}

// Multicall encoding:
// 1. Encode exitFarming call with IncentiveKey and tokenId
// 2. Optionally encode collectRewards call
// 3. Pass encoded calls as bytes[] to multicall function
```

---

### 2. Gas Tracking Pattern

**Question**: How should gas costs be tracked following the existing Stake function pattern?

**Decision**: Reuse the `TransactionRecord` struct and gas extraction utilities from the Stake function

**Rationale**:
- The Stake function (blackhole.go lines 543-772) demonstrates the established pattern:
  - `TransactionRecord` captures: TxHash, GasUsed, GasPrice, GasCost, Timestamp, Operation
  - `util.ExtractGasCost(receipt)` extracts the actual gas cost from transaction receipts
  - All transactions are appended to a slice and returned in `StakingResult`
- This pattern ensures constitutional compliance with Principle 3 (Financial Transparency)
- Consistency with existing codebase reduces maintenance burden

**Alternatives Considered**:
- **New tracking struct**: Rejected - introduces unnecessary inconsistency
- **No tracking**: Rejected - violates constitutional Principle 3 requirements

**Implementation Pattern** (from Stake function lines 629-651):
```go
gasCost, err := util.ExtractGasCost(receipt)
if err != nil {
    return &StakingResult{
        Success: false,
        ErrorMessage: fmt.Sprintf("failed to extract gas cost: %v", err),
    }, fmt.Errorf("failed to extract gas cost: %w", err)
}

gasPrice := new(big.Int)
gasPrice.SetString(receipt.EffectiveGasPrice, 0)
gasUsed := new(big.Int)
gasUsed.SetString(receipt.GasUsed, 0)

transactions = append(transactions, TransactionRecord{
    TxHash:    txHash,
    GasUsed:   gasUsed.Uint64(),
    GasPrice:  gasPrice,
    GasCost:   gasCost,
    Timestamp: time.Now(),
    Operation: "ExitFarming",
})
```

---

### 3. Reward Collection During Unstake

**Question**: Should rewards be automatically collected when unstaking?

**Decision**: Make reward collection optional via a boolean parameter `collectRewards`

**Rationale**:
- FarmingCenter.sol shows two separate functions:
  - `exitFarming` (lines 58-67): Exits farming position
  - `collectRewards` (lines 111-116): Collects accumulated rewards
- Users may want to:
  1. **Exit only**: Unstake position but leave rewards for later collection (tax optimization)
  2. **Exit + collect**: Unstake and immediately claim rewards (full withdrawal)
- Using multicall allows batching both operations atomically when desired
- Flexibility aligns with user story priority levels (P1: basic unstake, P2: emergency exit scenarios)

**Alternatives Considered**:
- **Always collect rewards**: Rejected - removes user flexibility, may have tax implications
- **Never collect rewards**: Rejected - users expect option to claim everything in one transaction
- **Two separate functions**: Rejected - multicall provides better UX and gas efficiency

**Implementation Approach**:
```go
func (b *Blackhole) Unstake(
    nftTokenID *big.Int,
    incentiveKey *IncentiveKey,
    collectRewards bool,  // Flag to control reward collection
) (*UnstakeResult, error) {
    // Build multicall data
    var multicallData [][]byte

    // Always include exitFarming
    exitFarmingData := encodeExitFarming(incentiveKey, nftTokenID)
    multicallData = append(multicallData, exitFarmingData)

    // Conditionally include collectRewards
    if collectRewards {
        collectRewardsData := encodeCollectRewards(incentiveKey, nftTokenID)
        multicallData = append(multicallData, collectRewardsData)
    }

    // Execute multicall
    // ...
}
```

---

### 4. Error Handling and Validation

**Question**: What validations and error handling are required for unstake operations?

**Decision**: Implement pre-flight checks matching the Stake function pattern plus NFT farming status verification

**Rationale**:
- Constitutional Principle 5 (Fail-Safe Operation) requires comprehensive error handling
- Stake function (lines 547-553) demonstrates input validation pattern
- FarmingCenter contract (lines 40-43) includes `isApprovedOrOwner` modifier
- exitFarming function (line 63) validates incentiveId matches deposits mapping

**Required Validations**:
1. **Input validation**:
   - NFT token ID must be non-nil and positive
   - IncentiveKey must have valid addresses (non-zero pool address)

2. **Ownership verification**:
   - User must own the NFT or be approved operator
   - Query `nonfungiblePositionManager.ownerOf(tokenId)`

3. **Farming status verification**:
   - NFT must be currently staked in FarmingCenter
   - Query `farmingCenter.deposits(tokenId)` to get current incentiveId
   - Verify incentiveId matches the provided IncentiveKey

4. **Transaction failure handling**:
   - Catch and log all transaction failures with full context
   - Return detailed error messages in UnstakeResult
   - Track partial gas costs even on failure (constitutional requirement)

**Alternatives Considered**:
- **Minimal validation**: Rejected - risks fund loss and violates Principle 5
- **On-chain validation only**: Rejected - wastes gas on preventable errors

**Implementation Example** (based on Stake pattern lines 567-584):
```go
// Verify NFT ownership
ownerResult, err := nftManagerClient.Call(&b.myAddr, "ownerOf", nftTokenID)
if err != nil {
    return &UnstakeResult{
        NFTTokenID:   nftTokenID,
        Success:      false,
        ErrorMessage: fmt.Sprintf("failed to verify NFT ownership: %v", err),
    }, fmt.Errorf("failed to verify NFT ownership: %w", err)
}

owner := ownerResult[0].(common.Address)
if owner != b.myAddr {
    return &UnstakeResult{
        NFTTokenID:   nftTokenID,
        Success:      false,
        ErrorMessage: fmt.Sprintf("NFT not owned by wallet: owned by %s", owner.Hex()),
    }, fmt.Errorf("NFT not owned by wallet")
}

// Verify NFT is currently farmed
farmingCenterClient, err := b.Client(farmingcenter)
depositsResult, err := farmingCenterClient.Call(&b.myAddr, "deposits", nftTokenID)
if err != nil {
    return &UnstakeResult{
        NFTTokenID:   nftTokenID,
        Success:      false,
        ErrorMessage: fmt.Sprintf("failed to check farming status: %v", err),
    }, fmt.Errorf("failed to check farming status: %w", err)
}

currentIncentiveId := depositsResult[0].([32]byte)
if currentIncentiveId == [32]byte{} {
    return &UnstakeResult{
        NFTTokenID:   nftTokenID,
        Success:      false,
        ErrorMessage: "NFT is not currently staked in farming",
    }, fmt.Errorf("NFT is not currently staked")
}
```

---

### 5. ABI Encoding for Multicall

**Question**: How should the exitFarming and collectRewards calls be encoded for multicall?

**Decision**: Use go-ethereum's `abi.Pack()` method with FarmingCenter ABI definitions

**Rationale**:
- The existing codebase uses ContractClient which abstracts ABI encoding
- For multicall, we need raw encoded bytes to pass as `bytes[]` parameter
- go-ethereum provides `abi.Method.Inputs.Pack()` for encoding function calls
- FarmingCenter ABI can be extracted from compiled contract artifacts

**Implementation Pattern**:
```go
import (
    "github.com/ethereum/go-ethereum/accounts/abi"
)

// Load FarmingCenter ABI (one-time initialization)
farmingCenterABI, err := abi.JSON(strings.NewReader(FarmingCenterABIJson))

// Encode exitFarming call
exitFarmingData, err := farmingCenterABI.Pack(
    "exitFarming",
    incentiveKey,  // IncentiveKey struct
    nftTokenID,    // uint256
)

// Encode collectRewards call (if needed)
collectRewardsData, err := farmingCenterABI.Pack(
    "collectRewards",
    incentiveKey,
    nftTokenID,
)

// Create multicall bytes array
multicallData := [][]byte{exitFarmingData}
if collectRewards {
    multicallData = append(multicallData, collectRewardsData)
}

// Execute multicall
txHash, err := farmingCenterClient.Send(
    types.Standard,
    nil,
    &b.myAddr,
    b.privateKey,
    "multicall",
    multicallData,  // bytes[] calldata data
)
```

**Alternatives Considered**:
- **Manual byte encoding**: Rejected - error-prone and difficult to maintain
- **Separate transactions**: Rejected - not atomic, higher gas costs

---

### 6. Result Structure

**Question**: Should unstake use the same StakingResult struct or create a new UnstakeResult?

**Decision**: Reuse `StakingResult` struct with field semantics adapted for unstake operations

**Rationale**:
- StakingResult already contains all necessary fields:
  - `NFTTokenID`: Identifies the unstaked position
  - `ActualAmount0/1`: Can represent LP tokens withdrawn
  - `Transactions`: Tracks all gas costs
  - `TotalGasCost`: Sum of all transaction costs
  - `Success/ErrorMessage`: Operation outcome
- Reusing existing struct reduces code duplication
- Both operations return similar data for financial tracking
- Type system makes it clear these are related operations

**Alternatives Considered**:
- **New UnstakeResult struct**: Rejected - creates unnecessary duplication
- **Generic OperationResult**: Rejected - loses type specificity

**Field Semantics for Unstake**:
```go
// StakingResult when used for Unstake:
&StakingResult{
    NFTTokenID:     nftTokenID,              // ID of unstaked NFT
    ActualAmount0:  rewardsAmount0,          // WAVAX rewards claimed (if collectRewards=true)
    ActualAmount1:  rewardsAmount1,          // USDC rewards claimed (if collectRewards=true)
    FinalTickLower: 0,                       // Not applicable for unstake
    FinalTickUpper: 0,                       // Not applicable for unstake
    Transactions:   []TransactionRecord{...},// Gas tracking for multicall
    TotalGasCost:   totalGas,                // Sum of all gas costs
    Success:        true,
    ErrorMessage:   "",
}
```

---

## Implementation Checklist

Based on research findings, the implementation must include:

- [ ] IncentiveKey struct definition in types.go
- [ ] Unstake function in blackhole.go following Stake pattern
- [ ] NFT ownership validation
- [ ] NFT farming status verification
- [ ] ABI encoding for exitFarming and collectRewards
- [ ] Multicall execution via FarmingCenter
- [ ] Gas tracking for all transactions
- [ ] Comprehensive error handling with StakingResult
- [ ] Logging output following Stake pattern (lines 763-769)
- [ ] Unit tests with mock contract clients
- [ ] Integration tests with mainnet transaction validation

## References

- `blackhole.go` lines 543-772: Stake function implementation pattern
- `FarmingCenter.sol` lines 58-67: exitFarming function
- `FarmingCenter.sol` lines 111-116: collectRewards function
- `Multicall.sol` lines 10-26: multicall interface
- Constitution Principle 3: Financial Transparency requirements
- Constitution Principle 5: Fail-Safe Operation requirements

## Open Questions

None - all research complete and decisions documented.
