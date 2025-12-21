# Quickstart: Liquidity Position Withdrawal

**Feature**: 003-liquidity-withdraw
**Date**: 2025-12-21

## Overview

This guide provides step-by-step instructions for implementing and using the `Withdraw` function to exit liquidity positions on Blackhole DEX.

## Prerequisites

- Go 1.24.10+ installed
- Access to Avalanche C-Chain RPC endpoint
- Wallet with:
  - Private key
  - An existing NFT position (from previous `Mint` operation)
  - Sufficient AVAX for gas fees
- Blackhole DEX contracts deployed and accessible

## Quick Start (5 Minutes)

### 1. Get Your NFT Token ID

If you minted a position earlier, you should have received an NFT token ID:

```go
// From previous Mint operation
mintResult, _ := blackhole.Mint(...)
nftTokenID := mintResult.NFTTokenID

log.Printf("Your NFT Token ID: %s", nftTokenID.String())
// Example output: "12345"
```

### 2. Execute Withdrawal

```go
package main

import (
    "fmt"
    "log"
    "math/big"
    "blackholego/blackholedex"
)

func main() {
    // Initialize Blackhole instance (see existing setup)
    blackhole := initializeBlackhole()

    // Your NFT token ID
    nftTokenID := big.NewInt(12345)

    // Execute withdrawal
    result, err := blackhole.Withdraw(nftTokenID)
    if err != nil {
        log.Fatalf("Withdrawal failed: %v", err)
    }

    // Print results
    fmt.Printf("✓ Withdrawal successful!\n")
    fmt.Printf("  WAVAX withdrawn: %s wei\n", result.Amount0.String())
    fmt.Printf("  USDC withdrawn: %s\n", result.Amount1.String())
    fmt.Printf("  Gas cost: %s wei\n", result.TotalGasCost.String())
}
```

### 3. Verify NFT is Burned

After withdrawal, verify the NFT no longer exists:

```go
// This should fail
_, err := nftManagerClient.Call(&blackhole.myAddr, "ownerOf", nftTokenID)
if err != nil {
    fmt.Println("✓ NFT successfully burned")
} else {
    fmt.Println("⚠ NFT still exists - unexpected!")
}
```

## Detailed Implementation Guide

### Step 1: Add Types to types.go

Add the new types required for withdrawal:

```go
// File: types.go

// WithdrawResult represents the complete output of withdrawal operation
type WithdrawResult struct {
    NFTTokenID     *big.Int            // Withdrawn NFT token ID
    Amount0        *big.Int            // WAVAX withdrawn (wei)
    Amount1        *big.Int            // USDC withdrawn (smallest unit)
    Transactions   []TransactionRecord // All transactions executed
    TotalGasCost   *big.Int            // Sum of all gas costs (wei)
    Success        bool                // Whether operation succeeded
    ErrorMessage   string              // Error message if failed
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
    Amount0Max *big.Int       `json:"amount0Max"`  // uint128
    Amount1Max *big.Int       `json:"amount1Max"`  // uint128
}
```

### Step 2: Implement Withdraw Method in blackhole.go

Add the `Withdraw` method to the `Blackhole` struct:

```go
// File: blackhole.go

// Withdraw removes all liquidity from an NFT position and burns the NFT
// nftTokenID: ERC721 token ID from previous Mint operation
// Returns WithdrawResult with transaction tracking and gas costs
func (b *Blackhole) Withdraw(nftTokenID *big.Int) (*WithdrawResult, error) {
    // Step 1: Input validation
    if nftTokenID == nil || nftTokenID.Sign() <= 0 {
        return &WithdrawResult{
            Success:      false,
            ErrorMessage: "validation failed: invalid token ID",
        }, fmt.Errorf("validation failed: invalid token ID")
    }

    // Step 2: Get NFT manager client
    nftManagerClient, err := b.Client(nonfungiblePositionManager)
    if err != nil {
        return &WithdrawResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to get NFT manager client: %v", err),
        }, fmt.Errorf("failed to get NFT manager client: %w", err)
    }

    // Step 3: Verify ownership
    ownerResult, err := nftManagerClient.Call(&b.myAddr, "ownerOf", nftTokenID)
    if err != nil {
        return &WithdrawResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to verify NFT ownership: %v", err),
        }, fmt.Errorf("failed to verify NFT ownership: %w", err)
    }

    owner := ownerResult[0].(common.Address)
    if owner != b.myAddr {
        return &WithdrawResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("NFT not owned by wallet: owned by %s", owner.Hex()),
        }, fmt.Errorf("NFT not owned by wallet")
    }

    // Step 4: Get position details
    positionsResult, err := nftManagerClient.Call(&b.myAddr, "positions", nftTokenID)
    if err != nil {
        return &WithdrawResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to query position: %v", err),
        }, fmt.Errorf("failed to query position: %w", err)
    }

    liquidity := positionsResult[7].(*big.Int)  // uint128 liquidity

    // Step 5: Build multicall data
    var multicallData [][]byte
    deadline := big.NewInt(time.Now().Add(20 * time.Minute).Unix())

    // TODO: Calculate expected amounts and slippage
    // For now, use zero minimums (production should calculate properly)
    amount0Min := big.NewInt(0)
    amount1Min := big.NewInt(0)

    // Encode decreaseLiquidity
    decreaseParams := &DecreaseLiquidityParams{
        TokenId:    nftTokenID,
        Liquidity:  liquidity,
        Amount0Min: amount0Min,
        Amount1Min: amount1Min,
        Deadline:   deadline,
    }

    nftManagerABI := nftManagerClient.Abi()
    decreaseData, err := nftManagerABI.Pack("decreaseLiquidity", decreaseParams)
    if err != nil {
        return &WithdrawResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to encode decreaseLiquidity: %v", err),
        }, fmt.Errorf("failed to encode decreaseLiquidity: %w", err)
    }
    multicallData = append(multicallData, decreaseData)

    // Encode collect
    maxUint128 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
    collectParams := &CollectParams{
        TokenId:    nftTokenID,
        Recipient:  b.myAddr,
        Amount0Max: maxUint128,
        Amount1Max: maxUint128,
    }

    collectData, err := nftManagerABI.Pack("collect", collectParams)
    if err != nil {
        return &WithdrawResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to encode collect: %v", err),
        }, fmt.Errorf("failed to encode collect: %w", err)
    }
    multicallData = append(multicallData, collectData)

    // Encode burn
    burnData, err := nftManagerABI.Pack("burn", nftTokenID)
    if err != nil {
        return &WithdrawResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to encode burn: %v", err),
        }, fmt.Errorf("failed to encode burn: %w", err)
    }
    multicallData = append(multicallData, burnData)

    // Step 6: Execute multicall
    txHash, err := nftManagerClient.Send(
        types.Standard,
        nil,
        &b.myAddr,
        b.privateKey,
        "multicall",
        multicallData,
    )
    if err != nil {
        return &WithdrawResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to submit multicall: %v", err),
        }, fmt.Errorf("failed to submit multicall: %w", err)
    }

    // Step 7: Wait for confirmation
    receipt, err := b.tl.WaitForTransaction(txHash)
    if err != nil {
        return &WithdrawResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("multicall transaction failed: %v", err),
        }, fmt.Errorf("multicall transaction failed: %w", err)
    }

    // Step 8: Extract gas cost
    gasCost, err := util.ExtractGasCost(receipt)
    if err != nil {
        return &WithdrawResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to extract gas cost: %v", err),
        }, fmt.Errorf("failed to extract gas cost: %w", err)
    }

    gasPrice := new(big.Int)
    gasPrice.SetString(receipt.EffectiveGasPrice, 0)
    gasUsed := new(big.Int)
    gasUsed.SetString(receipt.GasUsed, 0)

    // Step 9: Build result
    var transactions []TransactionRecord
    transactions = append(transactions, TransactionRecord{
        TxHash:    txHash,
        GasUsed:   gasUsed.Uint64(),
        GasPrice:  gasPrice,
        GasCost:   gasCost,
        Timestamp: time.Now(),
        Operation: "Withdraw",
    })

    // TODO: Parse multicall results to get actual amounts
    result := &WithdrawResult{
        NFTTokenID:   nftTokenID,
        Amount0:      big.NewInt(0),  // Should parse from multicall results
        Amount1:      big.NewInt(0),  // Should parse from multicall results
        Transactions: transactions,
        TotalGasCost: gasCost,
        Success:      true,
        ErrorMessage: "",
    }

    fmt.Printf("✓ Liquidity withdrawn successfully\n")
    fmt.Printf("  NFT ID: %s\n", nftTokenID.String())
    fmt.Printf("  Gas cost: %s wei\n", gasCost.String())

    return result, nil
}
```

### Step 3: Add Tests

Add integration tests to verify the implementation:

```go
// File: pkg/contractclient/contractclient_test.go

func TestWithdraw(t *testing.T) {
    // Setup
    blackhole := setupBlackhole(t)

    // Mint a position first
    mintResult, err := blackhole.Mint(
        big.NewInt(1e18),  // 1 WAVAX
        big.NewInt(1e6),   // 1 USDC
        6,
        5,
    )
    require.NoError(t, err)
    require.True(t, mintResult.Success)

    nftTokenID := mintResult.NFTTokenID

    // Execute withdrawal
    withdrawResult, err := blackhole.Withdraw(nftTokenID)
    require.NoError(t, err)
    require.True(t, withdrawResult.Success)

    // Verify NFT is burned
    _, err = nftManagerClient.Call(&blackhole.myAddr, "ownerOf", nftTokenID)
    require.Error(t, err)  // Should fail because NFT is burned

    // Verify gas tracking
    require.NotNil(t, withdrawResult.TotalGasCost)
    require.Greater(t, withdrawResult.TotalGasCost.Uint64(), uint64(0))
}
```

## Common Use Cases

### Use Case 1: Emergency Exit

Quickly exit a position due to market conditions:

```go
func emergencyExit(blackhole *Blackhole, nftTokenID *big.Int) error {
    result, err := blackhole.Withdraw(nftTokenID)
    if err != nil {
        log.Printf("Emergency exit failed: %v", err)
        return err
    }

    log.Printf("Emergency exit complete - %s WAVAX, %s USDC withdrawn",
        result.Amount0.String(), result.Amount1.String())
    return nil
}
```

### Use Case 2: Rebalancing Workflow

Withdraw and re-mint at different price range:

```go
func rebalance(blackhole *Blackhole, oldNFT *big.Int, newRange int) error {
    // Step 1: Withdraw old position
    withdrawResult, err := blackhole.Withdraw(oldNFT)
    if err != nil {
        return fmt.Errorf("withdraw failed: %w", err)
    }

    // Step 2: Mint new position with different range
    mintResult, err := blackhole.Mint(
        withdrawResult.Amount0,
        withdrawResult.Amount1,
        newRange,
        5,
    )
    if err != nil {
        return fmt.Errorf("re-mint failed: %w", err)
    }

    log.Printf("Rebalanced: Old NFT %s → New NFT %s",
        oldNFT.String(), mintResult.NFTTokenID.String())
    return nil
}
```

### Use Case 3: Batch Withdrawal

Withdraw multiple positions:

```go
func withdrawAll(blackhole *Blackhole, nftIDs []*big.Int) error {
    var totalGas = big.NewInt(0)

    for _, nftID := range nftIDs {
        result, err := blackhole.Withdraw(nftID)
        if err != nil {
            log.Printf("Failed to withdraw NFT %s: %v", nftID.String(), err)
            continue
        }

        totalGas.Add(totalGas, result.TotalGasCost)
        log.Printf("Withdrawn NFT %s - Gas: %s", nftID.String(), result.TotalGasCost.String())
    }

    log.Printf("Total gas spent: %s wei", totalGas.String())
    return nil
}
```

## Troubleshooting

### Error: "NFT not owned by wallet"

**Cause**: You're trying to withdraw an NFT you don't own.

**Solution**:
- Verify the NFT ID is correct
- Check wallet address matches the owner
- If staked in farming, unstake first

### Error: "failed to query position"

**Cause**: NFT doesn't exist or was already burned.

**Solution**:
- Verify NFT ID is correct
- Check if NFT was already withdrawn
- Query blockchain explorer for NFT status

### Error: "multicall transaction failed"

**Cause**: Transaction reverted during execution.

**Solution**:
- Check wallet has sufficient AVAX for gas
- Verify network connectivity
- Check if position is locked in farming contract
- Review transaction revert reason in receipt

## Next Steps

1. **Implement Slippage Calculation**: Add proper amount0Min/amount1Min calculations
2. **Parse Multicall Results**: Extract actual withdrawn amounts from transaction results
3. **Add Logging**: Implement structured logging for auditing
4. **Monitor Gas Costs**: Track and optimize gas consumption
5. **Test on Testnet**: Verify functionality before mainnet deployment

## Resources

- Feature Spec: [spec.md](./spec.md)
- API Contract: [contracts/withdraw-api.md](./contracts/withdraw-api.md)
- Data Model: [data-model.md](./data-model.md)
- Research: [research.md](./research.md)
- Constitution: [.specify/memory/constitution.md](../../.specify/memory/constitution.md)
