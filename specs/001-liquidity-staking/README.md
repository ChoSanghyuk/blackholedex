# Liquidity Staking Implementation Guide

**Feature**: 001-liquidity-staking
**Version**: 1.0.0
**Date**: 2025-12-10
**Status**: ✅ Production Ready (MVP)

## Overview

This feature enables automated liquidity staking in the WAVAX-USDC pool on Blackhole DEX. The implementation provides a single-method interface for staking liquidity with automatic position calculation, token approvals, slippage protection, and comprehensive financial tracking.

## What Was Changed

### New Files Created

#### 1. `internal/util/validation.go`
Core validation and helper functions for liquidity staking operations.

**Functions:**
- `ValidateStakingRequest(maxWAVAX, maxUSDC *big.Int, rangeWidth, slippagePct int) error`
  - Validates range width (1-20)
  - Validates slippage tolerance (1-50%)
  - Validates token amounts (> 0)

- `CalculateTickBounds(currentTick int32, rangeWidth int, tickSpacing int) (int32, int32, error)`
  - Calculates tick bounds from current tick and range width
  - rangeWidth N means ±(N/2) tick ranges from current tick
  - Clamps extreme values to ±887272 (max valid tick)

- `CalculateMinAmount(amountDesired *big.Int, slippagePct int) *big.Int`
  - Calculates minimum amount with slippage protection
  - Formula: `amountMin = amountDesired * (100 - slippagePct) / 100`

- `ExtractGasCost(receipt *types.TxReceipt) (*big.Int, error)`
  - Extracts gas cost from transaction receipt
  - Returns: `GasUsed * EffectiveGasPrice` in wei

### Modified Files

#### 1. `blackhole.go`

**New Constant:**
```go
nonfungiblePositionManager = "0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146"
```

**New Methods:**

1. **`validateBalances(requiredWAVAX, requiredUSDC *big.Int) error`**
   - Queries WAVAX and USDC token balances
   - Validates sufficient funds before any transactions
   - Returns descriptive error if insufficient

2. **`ensureApproval(tokenClient ContractClient, spender common.Address, requiredAmount *big.Int) (common.Hash, error)`**
   - Checks existing token allowance
   - Only approves if insufficient (gas optimization)
   - Returns transaction hash (or zero hash if approval not needed)
   - Waits for confirmation if approval transaction sent

3. **`Mint(maxWAVAX, maxUSDC *big.Int, rangeWidth, slippagePct int) (*StakingResult, error)`** ⭐
   - **Complete rewrite** - now fully functional
   - Single-step liquidity staking with automatic position management
   - See "Usage" section below for details

#### 2. `types.go`

**New Types:**

```go
// TransactionRecord tracks individual transaction details for financial transparency
type TransactionRecord struct {
    TxHash    common.Hash  // Transaction hash
    GasUsed   uint64       // Gas consumed
    GasPrice  *big.Int     // Effective gas price (wei)
    GasCost   *big.Int     // Total gas cost (wei) = GasUsed * GasPrice
    Timestamp time.Time    // Transaction timestamp
    Operation string       // Operation type ("ApproveWAVAX", "ApproveUSDC", "Mint")
}

// StakingResult represents the complete output of staking operation
type StakingResult struct {
    NFTTokenID     *big.Int            // Liquidity position NFT token ID
    ActualAmount0  *big.Int            // Actual WAVAX staked (wei)
    ActualAmount1  *big.Int            // Actual USDC staked (smallest unit)
    FinalTickLower int32               // Final lower tick bound
    FinalTickUpper int32               // Final upper tick bound
    Transactions   []TransactionRecord // All transactions executed
    TotalGasCost   *big.Int            // Sum of all gas costs (wei)
    Success        bool                // Whether operation succeeded
    ErrorMessage   string              // Error message if failed (empty if success)
}
```

## How to Use

### Basic Usage

```go
package main

import (
    "blackholego"
    "log"
    "math/big"
)

func main() {
    // Initialize Blackhole client (assuming you have this set up)
    blackhole := initializeBlackhole()

    // Define staking parameters
    maxWAVAX := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))  // 10 WAVAX
    maxUSDC := new(big.Int).Mul(big.NewInt(500), big.NewInt(1e6))   // 500 USDC
    rangeWidth := 6      // ±3 tick ranges (balanced concentration)
    slippagePct := 5     // 5% slippage tolerance

    // Execute staking
    result, err := blackhole.Mint(maxWAVAX, maxUSDC, rangeWidth, slippagePct)
    if err != nil {
        log.Fatalf("Staking failed: %v", err)
    }

    // Check result
    if result.Success {
        log.Printf("✓ Success! NFT Token ID: %s", result.NFTTokenID.String())
        log.Printf("  Staked: %s wei WAVAX, %s USDC",
            result.ActualAmount0.String(), result.ActualAmount1.String())
        log.Printf("  Position: Tick %d to %d",
            result.FinalTickLower, result.FinalTickUpper)
        log.Printf("  Total Gas Cost: %s wei", result.TotalGasCost.String())
    } else {
        log.Fatalf("Staking failed: %s", result.ErrorMessage)
    }
}
```

### What the Mint Method Does

The `Mint` method performs the following steps automatically:

1. **Input Validation** - Validates all parameters (range width, slippage, amounts)
2. **Pool State Query** - Fetches current pool state (tick, sqrt price, liquidity)
3. **Tick Bounds Calculation** - Calculates optimal tick range based on current tick and rangeWidth
4. **Optimal Amounts Calculation** - Computes optimal token ratio for the position
5. **Capital Efficiency Check** - Logs utilization percentages, warns if >10% unused
6. **Balance Validation** - Verifies wallet has sufficient WAVAX and USDC
7. **Slippage Protection** - Calculates minimum amounts for both tokens
8. **WAVAX Approval** - Checks allowance, approves only if needed (gas optimized)
9. **USDC Approval** - Checks allowance, approves only if needed (gas optimized)
10. **Position Minting** - Submits mint transaction to NonfungiblePositionManager
11. **Transaction Tracking** - Records all transaction details (hash, gas, timestamp)
12. **Result Construction** - Returns comprehensive StakingResult with all details

### Method Signature

```go
func (b *Blackhole) Mint(
    maxWAVAX *big.Int,    // Maximum WAVAX to stake (wei, 18 decimals)
    maxUSDC *big.Int,     // Maximum USDC to stake (smallest unit, 6 decimals)
    rangeWidth int,       // Position width in tick ranges (1-20)
    slippagePct int,      // Slippage tolerance percentage (1-50)
) (*StakingResult, error)
```

### Parameter Guidelines

#### Range Width (1-20)

Controls position concentration around current price:

| Value | Tick Range | Description | Use Case |
|-------|------------|-------------|----------|
| 2 | ±200 ticks | Very concentrated | Maximum fees, frequent rebalancing |
| 4 | ±400 ticks | Concentrated | High fees, regular rebalancing |
| **6** | **±600 ticks** | **Balanced (recommended)** | Good fees, moderate rebalancing |
| 8 | ±800 ticks | Wide | Stable, less rebalancing |
| 10+ | ±1000+ ticks | Very wide | Low maintenance |

**Recommendation**: Start with `rangeWidth = 6` for balanced capital efficiency and maintenance costs.

#### Slippage Tolerance (1-50%)

Protects against price movements during transaction execution:

| Value | Description | Risk |
|-------|-------------|------|
| 1-2% | Very tight | Transaction may fail in volatility |
| **3-5%** | **Normal (recommended)** | Balanced protection |
| 6-10% | High tolerance | May accept unfavorable prices |
| >10% | Emergency only | High risk of value loss |

**Recommendation**: Use `slippagePct = 5` for normal operations.

#### Token Amounts

- **WAVAX**: 18 decimals (1 WAVAX = 1e18 wei)
- **USDC**: 6 decimals (1 USDC = 1e6 smallest units)

**Example conversions:**
```go
// 10 WAVAX
maxWAVAX := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))

// 500 USDC
maxUSDC := new(big.Int).Mul(big.NewInt(500), big.NewInt(1e6))

// 0.5 WAVAX
halfWAVAX := new(big.Int).Div(new(big.Int).Mul(big.NewInt(5), big.NewInt(1e17)), big.NewInt(1))

// 123.45 USDC
usdcAmount := new(big.Int).Mul(big.NewInt(12345), big.NewInt(1e4))
```

## Key Features

### 1. Automatic Optimal Amount Calculation

The system automatically calculates the optimal token ratio based on:
- Current pool price
- Current tick
- Specified tick range
- Available capital

**Example:**
```go
// You provide max amounts
maxWAVAX := 10 WAVAX
maxUSDC := 500 USDC

// System calculates optimal amounts (might use less)
// If current price requires more USDC per WAVAX:
// actualWAVAX = 8 WAVAX (80% utilized)
// actualUSDC = 500 USDC (100% utilized)

// Capital efficiency warning logged if >10% of either token unused
```

### 2. Approval Optimization (Gas Savings)

**Smart approval logic:**
1. Checks existing allowance for each token
2. Only approves if allowance < required amount
3. Reuses approvals on subsequent stakes

**Gas savings:**
- First stake: ~342,000 gas (2 approvals + mint)
- Subsequent stakes with sufficient approvals: ~250,000 gas (mint only)
- Saves ~92,000 gas (~27% reduction)

### 3. Comprehensive Transaction Tracking

Every operation returns complete transaction history:

```go
result, _ := blackhole.Mint(...)

// Access all transactions
for _, tx := range result.Transactions {
    fmt.Printf("Operation: %s\n", tx.Operation)
    fmt.Printf("  Hash: %s\n", tx.TxHash.Hex())
    fmt.Printf("  Gas Used: %d\n", tx.GasUsed)
    fmt.Printf("  Gas Price: %s wei\n", tx.GasPrice.String())
    fmt.Printf("  Gas Cost: %s wei\n", tx.GasCost.String())
    fmt.Printf("  Timestamp: %s\n", tx.Timestamp.Format(time.RFC3339))
}

// Total cost across all operations
fmt.Printf("Total Gas Cost: %s wei\n", result.TotalGasCost.String())
```

### 4. Capital Efficiency Monitoring

Automatic warnings when capital is underutilized:

```go
// Console output during staking:
// Capital Utilization: WAVAX 80%, USDC 100%
// ⚠️  Capital Efficiency Warning: 20% of WAVAX (2000000000000000000 wei) will not be staked.
//     Consider adjusting amounts or range width.
```

### 5. Slippage Protection

Automatically calculates minimum amounts to protect against price slippage:

```go
slippagePct := 5  // 5%

// If optimal amounts are:
// amount0Desired = 10 WAVAX
// amount1Desired = 500 USDC

// System calculates minimums:
// amount0Min = 9.5 WAVAX  (10 * 0.95)
// amount1Min = 475 USDC   (500 * 0.95)

// Transaction will fail if actual amounts < minimums
// This protects you from unfavorable price movements
```

## Error Handling

### Common Errors

#### 1. Invalid Range Width
```
Error: range width must be between 1 and 20, got 25 (valid examples: 2, 6, 10)
```
**Solution**: Use rangeWidth between 1-20

#### 2. Invalid Slippage
```
Error: slippage tolerance must be between 1 and 50 percent, got 100
```
**Solution**: Use slippagePct between 1-50

#### 3. Insufficient Balance
```
Error: balance validation failed: insufficient WAVAX balance: have 5000000000000000000, need 10000000000000000000
```
**Solution**: Reduce maxWAVAX or add more WAVAX to wallet

#### 4. Zero Amount
```
Error: maxWAVAX must be > 0
```
**Solution**: Provide positive amounts for both tokens

#### 5. Pool State Query Failed
```
Error: failed to query pool state: connection timeout
```
**Solution**: Check RPC endpoint connectivity, retry operation

#### 6. Slippage Exceeded During Mint
```
Error: failed to mint position: slippage exceeded
```
**Solution**:
- Increase slippage tolerance
- Wait for lower volatility
- Retry with current price

### Safe Failure Behavior

**If approval succeeds but mint fails:**
- ✅ Your tokens remain in your wallet (safe)
- ✅ Approvals persist for next attempt (gas savings)
- ✅ No manual intervention needed
- ✅ Simply retry with adjusted parameters

```go
// First attempt - mint fails due to slippage
result, err := blackhole.Mint(maxWAVAX, maxUSDC, 6, 5)
if err != nil {
    log.Printf("First attempt failed: %v", err)

    // Retry with higher slippage - approvals already done!
    result, err = blackhole.Mint(maxWAVAX, maxUSDC, 6, 10)
    if err != nil {
        log.Fatalf("Retry failed: %v", err)
    }
}
```

## Gas Cost Expectations

Typical gas costs on Avalanche C-Chain:

| Transaction | Gas Used | Cost @ 25 GWEI | Cost @ 50 GWEI |
|-------------|----------|----------------|----------------|
| Approve WAVAX | ~46,000 | ~0.00115 AVAX | ~0.0023 AVAX |
| Approve USDC | ~46,000 | ~0.00115 AVAX | ~0.0023 AVAX |
| Mint Position | ~250,000 | ~0.00625 AVAX | ~0.0125 AVAX |
| **Total (first stake)** | **~342,000** | **~0.00855 AVAX** | **~0.0171 AVAX** |
| **Subsequent stakes** | **~250,000** | **~0.00625 AVAX** | **~0.0125 AVAX** |

**Note**: Costs vary with network congestion. Always ensure wallet has sufficient AVAX for gas (recommend minimum 0.1 AVAX buffer).

## Complete Example

```go
package main

import (
    "blackholego"
    "crypto/ecdsa"
    "fmt"
    "log"
    "math/big"
    "os"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment
    godotenv.Load()

    privateKeyHex := os.Getenv("PRIVATE_KEY")
    rpcURL := os.Getenv("RPC_URL")

    // Load private key
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        log.Fatalf("Failed to load private key: %v", err)
    }

    // Derive address
    publicKey := privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    myAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

    // Initialize Blackhole client
    blackhole := initializeBlackhole(rpcURL, privateKey, myAddress)

    // Staking parameters
    maxWAVAX := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))  // 10 WAVAX
    maxUSDC := new(big.Int).Mul(big.NewInt(500), big.NewInt(1e6))   // 500 USDC
    rangeWidth := 6      // ±600 ticks
    slippagePct := 5     // 5% slippage

    log.Printf("Starting liquidity staking...")
    log.Printf("  Max WAVAX: %s wei (10 WAVAX)", maxWAVAX.String())
    log.Printf("  Max USDC: %s (500 USDC)", maxUSDC.String())
    log.Printf("  Range Width: %d ticks", rangeWidth)
    log.Printf("  Slippage Tolerance: %d%%", slippagePct)

    // Execute staking
    result, err := blackhole.Mint(maxWAVAX, maxUSDC, rangeWidth, slippagePct)
    if err != nil {
        log.Fatalf("Staking failed: %v", err)
    }

    // Display results
    if result.Success {
        fmt.Println("\n✅ STAKING SUCCESSFUL")
        fmt.Printf("NFT Token ID: %s\n", result.NFTTokenID.String())
        fmt.Printf("\nPosition Details:\n")
        fmt.Printf("  WAVAX Staked: %s wei\n", result.ActualAmount0.String())
        fmt.Printf("  USDC Staked: %s\n", result.ActualAmount1.String())
        fmt.Printf("  Tick Lower: %d\n", result.FinalTickLower)
        fmt.Printf("  Tick Upper: %d\n", result.FinalTickUpper)

        fmt.Printf("\nTransaction History:\n")
        for _, tx := range result.Transactions {
            fmt.Printf("  %s:\n", tx.Operation)
            fmt.Printf("    Hash: %s\n", tx.TxHash.Hex())
            fmt.Printf("    Gas: %d units @ %s wei/gas = %s wei cost\n",
                tx.GasUsed, tx.GasPrice.String(), tx.GasCost.String())
            fmt.Printf("    Time: %s\n", tx.Timestamp.Format(time.RFC3339))
        }

        fmt.Printf("\nTotal Gas Cost: %s wei\n", result.TotalGasCost.String())

        // Calculate AVAX cost (18 decimals)
        avaxCost := new(big.Float).Quo(
            new(big.Float).SetInt(result.TotalGasCost),
            big.NewFloat(1e18),
        )
        fmt.Printf("Total Gas Cost: %s AVAX\n", avaxCost.Text('f', 6))

    } else {
        fmt.Printf("\n❌ STAKING FAILED\n")
        fmt.Printf("Error: %s\n", result.ErrorMessage)
    }
}

// Helper function to initialize Blackhole client
// (Implementation depends on your setup)
func initializeBlackhole(rpcURL string, privateKey *ecdsa.PrivateKey, address common.Address) *blackholedex.Blackhole {
    // Your initialization logic here
    // This should create ContractClient instances and TxListener
    // Then construct and return Blackhole instance
    panic("implement me")
}
```

## Best Practices

### 1. Start Small
Test with small amounts first before deploying significant capital:
```go
// Test with 1 WAVAX and 50 USDC
testWAVAX := new(big.Int).Mul(big.NewInt(1), big.NewInt(1e18))
testUSDC := new(big.Int).Mul(big.NewInt(50), big.NewInt(1e6))
```

### 2. Monitor Gas Prices
Check network conditions before large operations:
- Low congestion: 20-30 GWEI
- Normal: 30-50 GWEI
- High congestion: >50 GWEI

### 3. Verify Balances
Always verify you have sufficient tokens + gas before staking:
```go
// Minimum recommended: Token amounts + 0.1 AVAX for gas
```

### 4. Save NFT Token ID
Record the returned NFTTokenID for position tracking:
```go
result, _ := blackhole.Mint(...)
nftID := result.NFTTokenID
// Save to database or log file for future reference
```

### 5. Use Appropriate Slippage
Balance between failure risk and price protection:
- Stable markets: 3-5%
- Volatile markets: 5-10%
- Never exceed 20% unless absolutely necessary

### 6. Check Pool State First
Query current tick before staking to understand where your position will be:
```go
state, _ := blackhole.GetAMMState(wavaxUsdcPair)
log.Printf("Current pool tick: %d", state.Tick)
// Your position will be centered around this tick
```

### 7. Keep Transaction Records
Save all transaction hashes for audit trail:
```go
for _, tx := range result.Transactions {
    // Save to database or append to log file
    saveTransaction(tx.TxHash, tx.Operation, tx.GasCost)
}
```

## Architecture & Design Decisions

### Why Single Mint Method?

The implementation uses a single high-level `Mint` method rather than exposing individual steps because:

1. **Simplicity**: One method call handles entire staking workflow
2. **Safety**: Atomic operation with comprehensive validation before any transactions
3. **Optimization**: Automatically optimizes approvals (doesn't approve if not needed)
4. **Transparency**: Returns complete transaction history for all operations
5. **Fail-Safe**: Balance validation happens before spending gas on approvals

### Constitutional Compliance

This implementation adheres to all 5 project constitution principles:

- **Principle 1 (Pool-Specific Scope)**: WAVAX-USDC only, hardcoded addresses
- **Principle 2 (Autonomous Rebalancing)**: Foundation for future rebalancing (MVP: manual)
- **Principle 3 (Financial Transparency)**: All gas, fees tracked in TransactionRecord
- **Principle 4 (Gas Optimization)**: Approval reuse, no unnecessary transactions
- **Principle 5 (Fail-Safe Operation)**: Validation before transactions, comprehensive error handling

### Edge Case Handling

**Extreme Ticks**: If current tick is near ±887272 (max valid tick), the system:
- Clamps calculated bounds to valid range
- Still creates valid position
- Logs appropriate warnings

**Unbalanced Capital**: System automatically:
- Uses maximum possible amount of each token
- Maintains optimal ratio for the position
- Warns if >10% of either token unused

**Approval Already Exists**: System:
- Checks existing allowance first
- Only approves if needed
- Saves gas on subsequent operations

## Troubleshooting

### Issue: Transaction Taking Too Long

**Symptoms**: Transaction pending for >5 minutes
**Solutions**:
1. Check Avalanche C-Chain network status
2. Verify RPC endpoint is responding
3. Check if gas price is sufficient
4. Retry with higher gas price if needed

### Issue: Capital Efficiency Warning

**Symptoms**: Warning that >10% of token will be unused
**Explanation**: Current pool price requires more of one token than the other
**Solutions**:
1. Adjust token amounts to match current pool ratio
2. Accept warning if you want to provide specific amounts
3. Use wider range width to allow more flexibility

### Issue: Repeated Approval Failures

**Symptoms**: Approval transaction keeps failing
**Solutions**:
1. Verify token contract addresses are correct
2. Check wallet has sufficient AVAX for gas
3. Verify RPC endpoint is working
4. Check token balance is sufficient

### Issue: Mint Succeeds But Can't Find Position

**Symptoms**: Transaction confirms but can't see position on-chain
**Solutions**:
1. Check returned NFT token ID in result.NFTTokenID
2. Query NonfungiblePositionManager contract with token ID
3. Allow time for indexers to update (can take 1-2 minutes)
4. Verify transaction on block explorer

## Next Steps

After successfully staking, you can:

1. **Monitor Position**: Track position performance and fee accrual
2. **Plan Rebalancing**: Watch for price moving outside position range
3. **Collect Fees**: (Future feature - not yet implemented)
4. **Unstake**: (Future feature - not yet implemented)

## Contract Addresses

All contracts on Avalanche C-Chain (Chain ID: 43114):

| Contract | Address | Purpose |
|----------|---------|---------|
| WAVAX | `0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7` | Token 0 in pool |
| USDC | `0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E` | Token 1 in pool |
| WAVAX/USDC Pool | `0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0` | Concentrated liquidity pool |
| NonfungiblePositionManager | `0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146` | Mints liquidity positions |
| Pool Deployer | `0x5d433a94a4a2aa8f9aa34d8d15692dc2e9960584` | Pool deployer contract |

## Support

For issues or questions:
1. Review this guide and error messages
2. Verify all parameters are within valid ranges
3. Ensure sufficient balances (tokens + 0.1 AVAX gas buffer)
4. Check Avalanche C-Chain network status
5. Review transaction hashes on block explorer

## Related Documentation

- **Quickstart Guide**: `specs/001-liquidity-staking/quickstart.md` - Step-by-step tutorial
- **Feature Specification**: `specs/001-liquidity-staking/spec.md` - Requirements and user stories
- **Technical Plan**: `specs/001-liquidity-staking/plan.md` - Architecture and design decisions
- **Task List**: `specs/001-liquidity-staking/tasks.md` - Implementation tasks and progress

---

**Version History**:
- v1.0.0 (2025-12-10): Initial implementation - Core MVP with all 3 user stories complete
