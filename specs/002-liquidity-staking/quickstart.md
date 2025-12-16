# Quickstart: Implementing Stake Method

**Feature**: 002-liquidity-staking
**Audience**: Developers implementing the Stake method
**Time to Read**: 10 minutes
**Time to Implement**: 2-4 hours

## Overview

This guide walks you through implementing the `Stake` method that stakes liquidity position NFTs into GaugeV2 contracts. The method follows the same pattern as the existing `Mint` method for transaction tracking and error handling.

## Prerequisites

- Go 1.24.10+ installed
- Familiarity with go-ethereum library
- Understanding of ERC721 NFT standards
- Access to Avalanche C-Chain RPC
- Completed `Mint` implementation (reference pattern)

## Implementation Steps

### Step 1: Define Method Signature (5 minutes)

Add the method to `blackhole.go`:

```go
// Stake stakes a liquidity position NFT in a GaugeV2 contract to earn additional rewards
// nftTokenID: ERC721 token ID from previous Mint operation
// gaugeAddress: GaugeV2 contract address (must match pool)
// Returns StakingResult with transaction tracking and gas costs
func (b *Blackhole) Stake(
    nftTokenID *big.Int,
    gaugeAddress common.Address,
) (*StakingResult, error) {
    // Implementation here
}
```

**Why**: Follows existing naming conventions and uses standard types

### Step 2: Input Validation (15 minutes)

Validate inputs before any blockchain operations:

```go
// Validate token ID
if nftTokenID == nil || nftTokenID.Sign() <= 0 {
    return &StakingResult{
        Success:      false,
        ErrorMessage: "validation failed: invalid token ID",
    }, fmt.Errorf("validation failed: invalid token ID")
}

// Validate gauge address
if gaugeAddress == (common.Address{}) {
    return &StakingResult{
        NFTTokenID:   nftTokenID,
        Success:      false,
        ErrorMessage: "validation failed: invalid gauge address",
    }, fmt.Errorf("validation failed: invalid gauge address")
}

// Initialize transaction tracking (like Mint does)
var transactions []TransactionRecord
```

**Why**: Fail fast before expensive blockchain calls

### Step 3: Verify NFT Ownership (20 minutes)

Check that the user owns the NFT:

```go
// Get NonfungiblePositionManager client
nftManagerClient, err := b.Client(nonfungiblePositionManager)
if err != nil {
    return &StakingResult{
        NFTTokenID:   nftTokenID,
        Success:      false,
        ErrorMessage: fmt.Sprintf("failed to get NFT manager client: %v", err),
    }, fmt.Errorf("failed to get NFT manager client: %w", err)
}

// Query NFT ownership
ownerResult, err := nftManagerClient.Call(&b.myAddr, "ownerOf", nftTokenID)
if err != nil {
    return &StakingResult{
        NFTTokenID:   nftTokenID,
        Success:      false,
        ErrorMessage: fmt.Sprintf("failed to verify NFT ownership: %v", err),
    }, fmt.Errorf("failed to verify NFT ownership: %w", err)
}

owner := ownerResult[0].(common.Address)
if owner != b.myAddr {
    return &StakingResult{
        NFTTokenID:   nftTokenID,
        Success:      false,
        ErrorMessage: fmt.Sprintf("NFT not owned by wallet: owned by %s", owner.Hex()),
    }, fmt.Errorf("NFT not owned by wallet")
}
```

**Why**: Prevents gas waste on transactions that will fail

### Step 4: Check and Handle NFT Approval (30 minutes)

This is similar to token approval in `Mint`, but for ERC721:

```go
// Check current approval
approvalResult, err := nftManagerClient.Call(&b.myAddr, "getApproved", nftTokenID)
if err != nil {
    return &StakingResult{
        NFTTokenID:   nftTokenID,
        Success:      false,
        ErrorMessage: fmt.Sprintf("failed to check NFT approval: %v", err),
    }, fmt.Errorf("failed to check NFT approval: %w", err)
}

currentApproval := approvalResult[0].(common.Address)

// Only approve if not already approved for this gauge
if currentApproval != gaugeAddress {
    log.Printf("Approving NFT %s for gauge %s", nftTokenID.String(), gaugeAddress.Hex())

    approveTxHash, err := nftManagerClient.Send(
        types.Standard,
        nil, // Use automatic gas limit estimation
        &b.myAddr,
        b.privateKey,
        "approve",
        gaugeAddress,
        nftTokenID,
    )
    if err != nil {
        return &StakingResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to approve NFT: %v", err),
        }, fmt.Errorf("failed to approve NFT: %w", err)
    }

    // Wait for approval confirmation
    approvalReceipt, err := b.tl.WaitForTransaction(approveTxHash)
    if err != nil {
        return &StakingResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("NFT approval transaction failed: %v", err),
        }, fmt.Errorf("NFT approval transaction failed: %w", err)
    }

    // Track approval transaction (same pattern as Mint)
    gasCost, err := util.ExtractGasCost(approvalReceipt)
    if err != nil {
        return &StakingResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to extract approval gas cost: %v", err),
        }, fmt.Errorf("failed to extract approval gas cost: %w", err)
    }

    gasPrice := new(big.Int)
    gasPrice.SetString(approvalReceipt.EffectiveGasPrice, 0)
    gasUsed := new(big.Int)
    gasUsed.SetString(approvalReceipt.GasUsed, 0)

    transactions = append(transactions, TransactionRecord{
        TxHash:    approveTxHash,
        GasUsed:   gasUsed.Uint64(),
        GasPrice:  gasPrice,
        GasCost:   gasCost,
        Timestamp: time.Now(),
        Operation: "ApproveNFT",
    })
} else {
    log.Printf("NFT already approved for gauge, skipping approval")
}
```

**Why**: Gas optimization - reuse existing approval when possible

### Step 5: Execute Gauge Deposit (30 minutes)

Deposit the NFT into the gauge:

```go
// Get gauge client
gaugeClient, err := b.Client(gaugeAddress.Hex())
if err != nil {
    // Return with partial transaction records if approval was sent
    return &StakingResult{
        NFTTokenID:   nftTokenID,
        Transactions: transactions,
        TotalGasCost: sumGasCosts(transactions),
        Success:      false,
        ErrorMessage: fmt.Sprintf("failed to get gauge client: %v", err),
    }, fmt.Errorf("failed to get gauge client: %w", err)
}

// Submit deposit transaction
// Note: deposit(uint256 amount) where amount is the NFT token ID
log.Printf("Depositing NFT %s into gauge %s", nftTokenID.String(), gaugeAddress.Hex())

depositTxHash, err := gaugeClient.Send(
    types.Standard,
    nil, // Use automatic gas limit estimation
    &b.myAddr,
    b.privateKey,
    "deposit",
    nftTokenID, // Token ID is the "amount" parameter
)
if err != nil {
    return &StakingResult{
        NFTTokenID:   nftTokenID,
        Transactions: transactions,
        TotalGasCost: sumGasCosts(transactions),
        Success:      false,
        ErrorMessage: fmt.Sprintf("failed to submit deposit transaction: %v", err),
    }, fmt.Errorf("failed to submit deposit transaction: %w", err)
}

// Wait for deposit confirmation
depositReceipt, err := b.tl.WaitForTransaction(depositTxHash)
if err != nil {
    return &StakingResult{
        NFTTokenID:   nftTokenID,
        Transactions: transactions,
        TotalGasCost: sumGasCosts(transactions),
        Success:      false,
        ErrorMessage: fmt.Sprintf("deposit transaction failed: %v", err),
    }, fmt.Errorf("deposit transaction failed: %w", err)
}

// Track deposit transaction
gasCost, err := util.ExtractGasCost(depositReceipt)
if err != nil {
    return &StakingResult{
        NFTTokenID:   nftTokenID,
        Transactions: transactions,
        TotalGasCost: sumGasCosts(transactions),
        Success:      false,
        ErrorMessage: fmt.Sprintf("failed to extract deposit gas cost: %v", err),
    }, fmt.Errorf("failed to extract deposit gas cost: %w", err)
}

gasPrice := new(big.Int)
gasPrice.SetString(depositReceipt.EffectiveGasPrice, 0)
gasUsed := new(big.Int)
gasUsed.SetString(depositReceipt.GasUsed, 0)

transactions = append(transactions, TransactionRecord{
    TxHash:    depositTxHash,
    GasUsed:   gasUsed.Uint64(),
    GasPrice:  gasPrice,
    GasCost:   gasCost,
    Timestamp: time.Now(),
    Operation: "DepositNFT",
})
```

**Why**: Core functionality - transfers NFT and starts earning rewards

### Step 6: Construct Success Result (15 minutes)

Build the final result:

```go
// Calculate total gas cost
totalGasCost := big.NewInt(0)
for _, tx := range transactions {
    totalGasCost.Add(totalGasCost, tx.GasCost)
}

result := &StakingResult{
    NFTTokenID:     nftTokenID,
    ActualAmount0:  big.NewInt(0), // Not populated by Stake
    ActualAmount1:  big.NewInt(0), // Not populated by Stake
    FinalTickLower: 0,              // Not populated by Stake
    FinalTickUpper: 0,              // Not populated by Stake
    Transactions:   transactions,
    TotalGasCost:   totalGasCost,
    Success:        true,
    ErrorMessage:   "",
}

// Log success (matching Mint pattern)
fmt.Printf("✓ NFT staked successfully\n")
fmt.Printf("  Token ID: %s\n", nftTokenID.String())
fmt.Printf("  Gauge: %s\n", gaugeAddress.Hex())
fmt.Printf("  Total Gas Cost: %s wei\n", totalGasCost.String())
for _, tx := range transactions {
    fmt.Printf("  - %s: %s (gas: %s wei)\n", tx.Operation, tx.TxHash.Hex(), tx.GasCost.String())
}

return result, nil
```

**Why**: Consistent output format with comprehensive tracking

### Step 7: Add Helper Function (10 minutes)

If not already present, add a helper for gas summation:

```go
// sumGasCosts calculates total gas cost across multiple transactions
func sumGasCosts(transactions []TransactionRecord) *big.Int {
    total := big.NewInt(0)
    for _, tx := range transactions {
        total.Add(total, tx.GasCost)
    }
    return total
}
```

**Why**: Reusable utility for gas tracking

## Testing Your Implementation

### Unit Test Structure

```go
func TestStake(t *testing.T) {
    // Setup mock clients and test wallet
    blackhole := setupTestBlackhole(t)

    // Test case 1: Valid stake with approval
    t.Run("ValidStakeWithApproval", func(t *testing.T) {
        nftTokenID := big.NewInt(12345)
        gaugeAddr := common.HexToAddress("0x...")

        result, err := blackhole.Stake(nftTokenID, gaugeAddr)

        require.NoError(t, err)
        assert.True(t, result.Success)
        assert.Len(t, result.Transactions, 2) // Approval + Deposit
        assert.Greater(t, result.TotalGasCost.Int64(), int64(0))
    })

    // Test case 2: Valid stake without approval (already approved)
    t.Run("ValidStakeNoApproval", func(t *testing.T) {
        // ... similar structure but expect 1 transaction
    })

    // Test case 3: Invalid token ID
    t.Run("InvalidTokenID", func(t *testing.T) {
        result, err := blackhole.Stake(big.NewInt(0), validGauge)

        require.Error(t, err)
        assert.False(t, result.Success)
        assert.Contains(t, result.ErrorMessage, "invalid token ID")
    })
}
```

### Integration Test with Real Contracts

```go
func TestStakeIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Use testnet or mainnet fork
    blackhole := setupRealBlackhole(t)

    // First mint a position
    mintResult, err := blackhole.Mint(wanaxAmount, usdcAmount, 6, 5)
    require.NoError(t, err)
    require.True(t, mintResult.Success)

    // Then stake it
    gaugeAddr := common.HexToAddress("0x...") // Real gauge address
    stakeResult, err := blackhole.Stake(mintResult.NFTTokenID, gaugeAddr)

    require.NoError(t, err)
    assert.True(t, stakeResult.Success)

    // Verify on-chain state
    owner := queryNFTOwner(t, mintResult.NFTTokenID)
    assert.Equal(t, gaugeAddr, owner, "NFT should be owned by gauge")
}
```

## Common Pitfalls

### Pitfall 1: Using Wrong Approval Method
❌ **Wrong**: `setApprovalForAll(gauge, true)` - Approves ALL NFTs
✅ **Right**: `approve(gauge, tokenId)` - Approves single NFT only

### Pitfall 2: Not Checking Existing Approval
❌ **Wrong**: Always sending approval transaction
✅ **Right**: Check `getApproved(tokenId)` first, skip if already approved

### Pitfall 3: Inconsistent Error Handling
❌ **Wrong**: Returning `nil` result on error
✅ **Right**: Always return `&StakingResult{}` with `Success=false` and `ErrorMessage`

### Pitfall 4: Missing Gas Tracking
❌ **Wrong**: Tracking only successful transactions
✅ **Right**: Track all transactions, even if later steps fail

### Pitfall 5: Wrong Deposit Parameter
❌ **Wrong**: `deposit(0)` or `deposit(amount0 + amount1)`
✅ **Right**: `deposit(nftTokenID)` - Token ID is the parameter

## Verification Checklist

Before marking implementation complete:

- [ ] Method signature matches API contract
- [ ] Input validation prevents invalid calls
- [ ] NFT ownership checked before transactions
- [ ] Existing approval reused when possible
- [ ] Both approval and deposit transactions tracked
- [ ] Gas costs accurately calculated and summed
- [ ] Error messages are descriptive and actionable
- [ ] Success logging matches Mint pattern
- [ ] Unit tests cover success and failure paths
- [ ] Integration test verifies on-chain state
- [ ] Code follows existing patterns from Mint method
- [ ] Constitutional principles satisfied (see research.md)

## Next Steps

After implementing Stake:

1. **Test thoroughly** with unit and integration tests
2. **Document edge cases** encountered during testing
3. **Run gas benchmarks** to verify optimization
4. **Create examples** in cmd/ directory for users
5. **Consider unstaking** as follow-up feature

## Reference Files

- **Specification**: `specs/002-liquidity-staking/spec.md`
- **Research**: `specs/002-liquidity-staking/research.md`
- **Data Model**: `specs/002-liquidity-staking/data-model.md`
- **API Contract**: `specs/002-liquidity-staking/contracts/stake-api.md`
- **Existing Mint**: `blackhole.go` lines 221-533
- **Approval Pattern**: `blackhole.go` lines 183-219

## Getting Help

- Review the Mint method implementation for patterns
- Check GaugeV2 ABI: `blackholedex-contracts/artifacts/contracts/GaugeV2.sol/GaugeV2.json`
- Check NFT Manager ABI: Interface in artifacts directory
- Constitution: `.specify/memory/constitution.md`

## Estimated Timeline

- Implementation: 2-4 hours
- Testing: 1-2 hours
- Integration: 30 minutes
- Documentation: 30 minutes

**Total**: ~4-7 hours for complete feature delivery
