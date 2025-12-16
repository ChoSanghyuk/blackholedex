# API Contract: Mint Operation

**Feature**: 001-liquidity-staking
**Date**: 2025-12-09
**Version**: 1.0.0

## Overview

This document defines the public API for the liquidity staking (`Mint`) operation in the Blackhole DEX client library.

## Core Operation

### Mint

Stakes liquidity in the WAVAX-USDC concentrated liquidity pool with automatic position calculation and comprehensive tracking.

**Signature**:
```go
func (b *Blackhole) Mint(
    maxWAVAX *big.Int,
    maxUSDC *big.Int,
    rangeWidth int,
    slippagePct int,
) (*StakingResult, error)
```

**Parameters**:

| Parameter | Type | Description | Constraints |
|-----------|------|-------------|-------------|
| maxWAVAX | *big.Int | Maximum WAVAX amount to stake (wei) | > 0, <= wallet balance |
| maxUSDC | *big.Int | Maximum USDC amount to stake (smallest unit, 6 decimals) | > 0, <= wallet balance |
| rangeWidth | int | Position range width (e.g., 6 = ±3 tick ranges) | > 0, <= 20 |
| slippagePct | int | Slippage tolerance percentage (e.g., 5 = 5%) | > 0, <= 50 |

**Returns**:

| Return | Type | Description |
|--------|------|-------------|
| result | *StakingResult | Complete staking operation result including tx hashes, gas costs, and position details |
| error | error | Error if operation failed, nil if successful |

**Behavior**:

1. Validates all input parameters
2. Queries current WAVAX-USDC pool state
3. Calculates optimal tick bounds based on current price and range width
4. Calculates optimal token amounts using ComputeAmounts utility
5. Validates wallet balances
6. Ensures token approvals for NonfungiblePositionManager (optimized to reuse existing allowances)
7. Submits mint transaction to create liquidity position NFT
8. Waits for transaction confirmation
9. Extracts gas costs from all transactions
10. Returns complete StakingResult

**Errors**:

| Error Type | Condition | Example Message |
|------------|-----------|-----------------|
| Validation | Invalid inputs | "range width must be between 1 and 20, got 0" |
| Validation | Invalid slippage | "slippage tolerance must be between 1 and 50 percent, got 100" |
| Balance | Insufficient funds | "insufficient WAVAX balance: have 1000000000000000000, need 5000000000000000000" |
| Network | RPC failure | "failed to query pool state: connection timeout" |
| Transaction | Approval failed | "failed to approve WAVAX: transaction reverted" |
| Transaction | Mint failed | "failed to mint position: slippage exceeded" |
| Calculation | Tick out of range | "tick bounds out of valid range: [-900000, -700000]" |

**Example Usage**:

```go
// Initialize Blackhole client
blackhole := &Blackhole{
    privateKey: privateKey,
    myAddr:     myAddress,
    tl:         txListener,
    ccm:        contractClientMap,
}

// Stake 10 WAVAX and 500 USDC with moderate range
maxWAVAX := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))      // 10 WAVAX
maxUSDC := new(big.Int).Mul(big.NewInt(500), big.NewInt(1e6))       // 500 USDC
rangeWidth := 6                                                      // ±3 tick ranges
slippagePct := 5                                                     // 5% slippage tolerance

result, err := blackhole.Mint(maxWAVAX, maxUSDC, rangeWidth, slippagePct)
if err != nil {
    log.Fatalf("Staking failed: %v", err)
}

// Log results
log.Printf("✓ Position created: NFT token ID %s", result.NFTTokenID.String())
log.Printf("✓ Staked: %s WAVAX, %s USDC", result.ActualAmount0.String(), result.ActualAmount1.String())
log.Printf("✓ Total gas cost: %s wei (%d AVAX)", result.TotalGasCost.String(), result.TotalGasCost.Uint64()/1e18)

for _, tx := range result.Transactions {
    log.Printf("  - %s: %s (gas: %s wei)", tx.Operation, tx.TxHash.Hex(), tx.GasCost.String())
}
```

---

## Supporting Types

### StakingResult

Complete output of staking operation.

**Definition**:
```go
type StakingResult struct {
    NFTTokenID     *big.Int              // Liquidity position NFT token ID
    ActualAmount0  *big.Int              // Actual WAVAX staked (wei)
    ActualAmount1  *big.Int              // Actual USDC staked (smallest unit)
    FinalTickLower int32                 // Final lower tick bound
    FinalTickUpper int32                 // Final upper tick bound
    Transactions   []TransactionRecord   // All transactions executed
    TotalGasCost   *big.Int              // Sum of all gas costs (wei)
    Success        bool                  // Whether operation succeeded
    ErrorMessage   string                // Error message if failed (empty if success)
}
```

### TransactionRecord

Individual transaction details for financial tracking.

**Definition**:
```go
type TransactionRecord struct {
    TxHash    common.Hash   // Transaction hash
    GasUsed   uint64        // Gas consumed
    GasPrice  *big.Int      // Effective gas price (wei)
    GasCost   *big.Int      // Total gas cost (wei) = GasUsed * GasPrice
    Timestamp time.Time     // Transaction timestamp
    Operation string        // Operation type ("ApproveWAVAX", "ApproveUSDC", "Mint")
}
```

---

## Internal Helper Functions

These are not part of the public API but are documented for implementation reference.

### validateStakingRequest

Validates input parameters before workflow execution.

**Signature**:
```go
func validateStakingRequest(maxWAVAX, maxUSDC *big.Int, rangeWidth, slippagePct int) error
```

**Returns**: Error if validation fails, nil otherwise

---

### calculateTickBounds

Calculates tick bounds from current tick and range width.

**Signature**:
```go
func calculateTickBounds(currentTick int32, rangeWidth int, tickSpacing int) (tickLower, tickUpper int32, err error)
```

**Returns**:
- `tickLower`: Lower tick bound
- `tickUpper`: Upper tick bound
- `err`: Error if bounds invalid

---

### calculateMinAmounts

Calculates minimum amounts with slippage protection.

**Signature**:
```go
func calculateMinAmounts(amount0Desired, amount1Desired *big.Int, slippagePct int) (amount0Min, amount1Min *big.Int)
```

**Returns**:
- `amount0Min`: Minimum WAVAX amount
- `amount1Min`: Minimum USDC amount

---

### ensureApproval

Ensures token approval exists, optimizing to reuse existing allowances.

**Signature**:
```go
func (b *Blackhole) ensureApproval(
    tokenClient ContractClient,
    spender common.Address,
    requiredAmount *big.Int,
) (txHash common.Hash, err error)
```

**Returns**:
- `txHash`: Approval transaction hash (zero if approval not needed)
- `err`: Error if approval failed

---

### extractGasCost

Extracts gas cost from transaction receipt.

**Signature**:
```go
func extractGasCost(receipt *types.TxReceipt) (*big.Int, error)
```

**Returns**:
- Gas cost in wei
- Error if receipt invalid

---

### validateBalances

Validates wallet has sufficient token balances.

**Signature**:
```go
func (b *Blackhole) validateBalances(requiredWAVAX, requiredUSDC *big.Int) error
```

**Returns**: Error if insufficient balance, nil otherwise

---

## Constitutional Compliance

This API design satisfies all constitutional principles:

**Principle 1 (Pool-Specific Scope)**:
- Mint method hardcoded to WAVAX/USDC pool
- No token address parameters (prevents other pools)

**Principle 2 (Autonomous Rebalancing)**:
- N/A for initial staking (foundation for future automation)

**Principle 3 (Financial Transparency)**:
- StakingResult includes all transaction hashes
- Gas costs tracked for every transaction
- Actual staked amounts returned

**Principle 4 (Gas Optimization)**:
- Approval reuse via allowance check
- Gas estimation handled by ContractClient
- Single pool state query

**Principle 5 (Fail-Safe Operation)**:
- Comprehensive input validation
- Balance validation before transactions
- Slippage protection enforced
- Sequential execution (approval → mint)
- Detailed error messages

---

## Versioning

**Current Version**: 1.0.0

**Breaking Changes**:
- Changes to Mint signature
- Changes to StakingResult structure

**Non-Breaking Changes**:
- Adding new helper functions
- Enhanced error messages
- Performance optimizations

---

## Testing Contract

The Mint operation must pass these contract tests:

1. **Valid staking with balanced amounts**: Returns success with NFTTokenID > 0
2. **Valid staking with unbalanced amounts**: Uses maximum possible capital without exceeding balances
3. **Insufficient balance**: Returns error before any transactions submitted
4. **Invalid range width (0)**: Returns validation error immediately
5. **Invalid slippage (>50%)**: Returns validation error immediately
6. **Existing approvals**: Does not submit redundant approval transactions
7. **Price movement during execution**: Fails with slippage error if movement exceeds tolerance
8. **Network failure during approval**: Returns error without attempting mint
9. **Gas cost tracking**: StakingResult.TotalGasCost matches sum of individual transaction costs
10. **Tick bounds calculation**: Tick bounds are symmetric around current tick with specified width

See integration tests in `blackhole_test.go` for complete test implementation.
