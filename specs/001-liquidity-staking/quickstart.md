# Quickstart Guide: Liquidity Staking

**Feature**: 001-liquidity-staking
**Date**: 2025-12-09
**Audience**: Operators managing liquidity on Blackhole DEX

## Overview

This guide explains how to stake liquidity in the WAVAX-USDC pool using the Blackhole DEX Go client.

## Prerequisites

1. **Wallet Setup**:
   - Private key with AVAX for gas fees
   - WAVAX tokens (wrapped AVAX)
   - USDC tokens
   - Minimum recommended: 1 AVAX for gas + desired staking amounts

2. **Environment**:
   - Go 1.24.10 or higher
   - Access to Avalanche C-Chain RPC endpoint
   - Blackhole DEX contracts deployed on C-Chain

3. **Configuration**:
   - RPC endpoint URL
   - Private key (securely stored, never committed to git)

## Quick Start

### 1. Initialize Client

```go
package main

import (
    "blackholego"
    "crypto/ecdsa"
    "log"
    "math/big"
    "os"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    godotenv.Load()

    // Load private key
    privateKeyHex := os.Getenv("PRIVATE_KEY")
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        log.Fatalf("Failed to load private key: %v", err)
    }

    // Derive address
    publicKey := privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    myAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

    // Initialize Blackhole client (implementation details depend on constructor)
    blackhole := initializeBlackhole(privateKey, myAddress)

    log.Printf("Initialized client for address: %s", myAddress.Hex())
}
```

### 2. Stake Liquidity

```go
func stakeLP(blackhole *blackholedex.Blackhole) {
    // Define staking parameters
    maxWAVAX := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))  // 10 WAVAX
    maxUSDC := new(big.Int).Mul(big.NewInt(500), big.NewInt(1e6))   // 500 USDC
    rangeWidth := 6                                                  // ±3 tick ranges
    slippagePct := 5                                                 // 5% slippage

    log.Printf("Staking up to %s WAVAX and %s USDC...",
        formatWAVAX(maxWAVAX), formatUSDC(maxUSDC))

    // Execute staking
    result, err := blackhole.Mint(maxWAVAX, maxUSDC, rangeWidth, slippagePct)
    if err != nil {
        log.Fatalf("Staking failed: %v", err)
    }

    // Display results
    log.Printf("✓ Success! NFT Token ID: %s", result.NFTTokenID.String())
    log.Printf("  Staked: %s WAVAX, %s USDC",
        formatWAVAX(result.ActualAmount0), formatUSDC(result.ActualAmount1))
    log.Printf("  Position: Tick %d to %d",
        result.FinalTickLower, result.FinalTickUpper)
    log.Printf("  Total Gas Cost: %s AVAX", formatAVAX(result.TotalGasCost))

    // Log transaction details
    for _, tx := range result.Transactions {
        log.Printf("  - %s: %s (gas: %s AVAX)",
            tx.Operation, tx.TxHash.Hex()[:10]+"...",
            formatAVAX(tx.GasCost))
    }
}

// Helper: Format WAVAX (18 decimals)
func formatWAVAX(amount *big.Int) string {
    divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
    whole := new(big.Int).Div(amount, divisor)
    return whole.String() + " WAVAX"
}

// Helper: Format USDC (6 decimals)
func formatUSDC(amount *big.Int) string {
    divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)
    whole := new(big.Int).Div(amount, divisor)
    return whole.String() + " USDC"
}

// Helper: Format AVAX (18 decimals)
func formatAVAX(amount *big.Int) string {
    divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
    whole := new(big.Int).Div(amount, divisor)
    remainder := new(big.Int).Mod(amount, divisor)
    fractional := new(big.Int).Div(remainder, big.NewInt(1e15)) // 3 decimals
    return whole.String() + "." + fractional.String() + " AVAX"
}
```

### 3. Run

```bash
# Set environment variables
export PRIVATE_KEY="your_private_key_without_0x_prefix"
export RPC_URL="https://api.avax.network/ext/bc/C/rpc"

# Run
go run cmd/stake/main.go
```

## Parameter Selection Guide

### Range Width

The range width parameter controls position concentration:

| Range Width | Tick Range | Use Case | Rebalancing Frequency |
|-------------|------------|----------|----------------------|
| 2 | ±200 ticks | Very concentrated, max fees | Very frequent |
| 4 | ±400 ticks | Concentrated, high fees | Frequent |
| 6 | ±600 ticks | **Balanced (recommended)** | Moderate |
| 8 | ±800 ticks | Wide, stable | Infrequent |
| 10 | ±1000 ticks | Very wide, low maintenance | Rare |

**Recommendation**: Start with rangeWidth=6 (±600 ticks) for balanced capital efficiency and rebalancing costs.

### Slippage Tolerance

Slippage tolerance protects against price movements during transaction execution:

| Slippage % | Use Case | Risk |
|------------|----------|------|
| 1-2% | Stable markets, fast execution | May fail during volatility |
| 3-5% | **Normal conditions (recommended)** | Balanced |
| 6-10% | High volatility periods | May accept unfavorable prices |
| >10% | Emergency only | High risk of value loss |

**Recommendation**: Use 5% for normal operations, increase to 8-10% only during extreme volatility.

### Token Amounts

Choose amounts based on your capital and strategy:

**Equal Value Approach** (recommended for beginners):
```go
// Example: $1000 worth of liquidity
// Assuming WAVAX = $25, USDC = $1
maxWAVAX := new(big.Int).Mul(big.NewInt(20), big.NewInt(1e18))   // 20 WAVAX = $500
maxUSDC := new(big.Int).Mul(big.NewInt(500), big.NewInt(1e6))    // 500 USDC = $500
```

**Unbalanced Approach** (for specific strategies):
```go
// The system will use maximum possible from available capital
// Example: More USDC than WAVAX if expecting price to rise
maxWAVAX := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))   // 10 WAVAX
maxUSDC := new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e6))   // 1000 USDC
```

## Common Operations

### Check Current Position

```go
// Query pool state to see current tick
state, err := blackhole.GetAMMState(common.HexToAddress(wavaxUsdcPair))
if err != nil {
    log.Fatalf("Failed to get pool state: %v", err)
}

log.Printf("Current pool tick: %d", state.Tick)
log.Printf("Current sqrt price: %s", state.SqrtPrice.String())
log.Printf("Active liquidity: %s", state.ActiveLiquidity.String())
```

### Calculate Expected Position

```go
// Preview position before staking
currentTick := int32(state.Tick)
rangeWidth := 6
tickSpacing := 200

tickLower := (int(currentTick)/tickSpacing - rangeWidth/2) * tickSpacing
tickUpper := (int(currentTick)/tickSpacing + rangeWidth/2) * tickSpacing

log.Printf("Position will be: tick %d to %d", tickLower, tickUpper)
log.Printf("Current tick %d is within range: %v",
    currentTick, currentTick >= int32(tickLower) && currentTick <= int32(tickUpper))
```

## Error Handling

### Common Errors and Solutions

#### 1. Insufficient Balance

```
Error: insufficient WAVAX balance: have 5000000000000000000, need 10000000000000000000
```

**Solution**: Reduce maxWAVAX amount or add more WAVAX to wallet

#### 2. Slippage Exceeded

```
Error: failed to mint position: slippage exceeded
```

**Solution**: Increase slippage tolerance or wait for lower volatility

#### 3. Range Width Too Large

```
Error: range width must be between 1 and 20, got 50
```

**Solution**: Use rangeWidth between 1 and 20

#### 4. Invalid Slippage

```
Error: slippage tolerance must be between 1 and 50 percent, got 100
```

**Solution**: Use slippagePct between 1 and 50

#### 5. Network Timeout

```
Error: failed to query pool state: connection timeout
```

**Solution**: Check RPC endpoint connectivity, retry operation

#### 6. Gas Estimation Failed

```
Error: failed to estimate gas: insufficient funds for gas
```

**Solution**: Ensure wallet has enough AVAX for gas (recommend minimum 0.1 AVAX)

### Partial Failure Recovery

If approval succeeds but mint fails:

1. **Funds are safe**: Tokens remain in your wallet (approved but not staked)
2. **Approvals persist**: Next mint attempt will reuse approvals (saves gas)
3. **No manual intervention needed**: Simply retry with adjusted parameters

```go
// Retry with increased slippage
result, err := blackhole.Mint(maxWAVAX, maxUSDC, rangeWidth, 10) // Increase from 5% to 10%
if err != nil {
    log.Printf("Retry failed: %v", err)
    return
}
```

## Gas Cost Expectations

Typical gas costs for staking operation:

| Transaction | Gas Used | Cost (GWEI=25) | Cost (GWEI=50) |
|-------------|----------|----------------|----------------|
| Approve WAVAX | ~46,000 | ~0.00115 AVAX | ~0.0023 AVAX |
| Approve USDC | ~46,000 | ~0.00115 AVAX | ~0.0023 AVAX |
| Mint Position | ~250,000 | ~0.00625 AVAX | ~0.0125 AVAX |
| **Total** | **~342,000** | **~0.00855 AVAX** | **~0.0171 AVAX** |

Note: Costs vary with network congestion. If approvals already exist, only mint gas is paid.

## Best Practices

1. **Start Small**: Test with small amounts first (e.g., 1 WAVAX, 50 USDC)
2. **Monitor Gas**: Check network gas prices before large operations
3. **Verify Balances**: Always verify token balances before staking
4. **Save NFT Token ID**: Record the returned NFTTokenID for position tracking
5. **Log Transactions**: Keep transaction hashes for audit trail
6. **Use Appropriate Slippage**: Balance between failure risk and price protection
7. **Check Pool State**: Query current tick before staking to verify position will be in-range

## Security Considerations

1. **Private Key Protection**:
   - Never hardcode private keys
   - Use environment variables or secure key management
   - Never commit .env files to git

2. **Amount Validation**:
   - Double-check token amounts (decimals: WAVAX=18, USDC=6)
   - Verify amounts in UI before execution

3. **Transaction Monitoring**:
   - Always wait for transaction confirmation
   - Verify transaction success on block explorer
   - Save transaction hashes for records

4. **Slippage Protection**:
   - Never use >20% slippage unless absolutely necessary
   - Understand that high slippage allows unfavorable execution prices

## Next Steps

After successfully staking:

1. **Monitor Position**: Track position performance and fee accrual
2. **Plan Rebalancing**: Watch for price moving outside position range
3. **Collect Fees**: Future feature will enable fee collection
4. **Unstake**: Future feature will enable position withdrawal

## Support

For issues or questions:
- Check error message and refer to "Common Errors" section above
- Verify all parameters are within valid ranges
- Ensure sufficient balances (tokens + gas)
- Check Avalanche C-Chain network status

## Appendix: Contract Addresses

| Contract | Address | Purpose |
|----------|---------|---------|
| WAVAX | 0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7 | Token 0 in pool |
| USDC | 0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E | Token 1 in pool |
| WAVAX/USDC Pool | 0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0 | Liquidity pool |
| NonfungiblePositionManager | 0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146 | Mints positions |
| Deployer | 0x5d433a94a4a2aa8f9aa34d8d15692dc2e9960584 | Pool deployer |

All addresses are on Avalanche C-Chain (Chain ID: 43114).
