# Research: Liquidity Staking

**Feature**: 001-liquidity-staking
**Date**: 2025-12-09
**Status**: Complete

## Research Questions

This document consolidates research findings for technical decisions needed to implement the liquidity staking feature.

## Q1: NonfungiblePositionManager Contract Address

**Question**: What is the correct NonfungiblePositionManager contract address for Blackhole DEX on Avalanche C-Chain?

**Research**:
- Examined README.md and found reference to NonfungiblePositionManager in mint example transaction
- Transaction 0x9e2247a0210448cab301475eef741eba0ee9a9351188a92b8127fce27206b9d0 shows contractAddr: `0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146`
- This address is used in the decoded mint transaction data in README.md line 136

**Decision**: Use `0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146` as the NonfungiblePositionManager contract address

**Rationale**: This address is verified from actual on-chain transactions and matches the spec's constitutional constraints

**Alternatives Considered**:
- None - address is definitively identified from historical transactions

**Implementation**: Add constant to blackhole.go:
```go
const nonfungiblePositionManager = "0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146"
```

---

## Q2: Approval Optimization Strategy

**Question**: How should we optimize ERC20 approvals to minimize gas costs while maintaining safety?

**Research**:
- Examined existing `Swap` method in blackhole.go (lines 43-103) which always calls `approve()` before swap
- Constitutional Principle 4 requires approval reuse when safe
- ERC20 standard provides `allowance(owner, spender)` view function to check existing allowance
- Unnecessary approve() calls waste gas (~46,000 gas per approval)

**Decision**: Check existing allowance before approving; only call approve() if current allowance < required amount

**Rationale**:
- Reduces gas costs when approvals already exist
- Maintains safety (never reduces existing allowance without explicit approval)
- Aligns with constitutional gas optimization principle

**Alternatives Considered**:
1. **Always approve**: Simple but wastes gas on repeat operations (rejected - violates Principle 4)
2. **Approve unlimited (type(uint256).max)**: One-time approval but security risk (rejected - violates Principle 5 fail-safe)
3. **Approve exact amount + check**: Optimal gas and safe (selected)

**Implementation**:
```go
func (b *Blackhole) ensureApproval(
    tokenClient ContractClient,
    spender common.Address,
    requiredAmount *big.Int,
) (txHash common.Hash, error) {
    // Check existing allowance
    result, err := tokenClient.Call(&b.myAddr, "allowance", b.myAddr, spender)
    if err != nil {
        return common.Hash{}, fmt.Errorf("failed to check allowance: %w", err)
    }

    currentAllowance := result[0].(*big.Int)

    // Only approve if insufficient
    if currentAllowance.Cmp(requiredAmount) >= 0 {
        return common.Hash{}, nil // No approval needed
    }

    // Approve required amount
    return tokenClient.Send(
        types.Standard,
        nil,
        &b.myAddr,
        b.privateKey,
        "approve",
        spender,
        requiredAmount,
    )
}
```

---

## Q3: Slippage Protection Implementation

**Question**: How should we calculate minimum amounts (amount0Min, amount1Min) from slippage tolerance percentage?

**Research**:
- Spec requires slippage protection with configurable tolerance (default 5% per user's updated spec line 166)
- AMM minting can fail if price moves significantly between calculation and execution
- Minimum amounts protect against excessive slippage
- Must use big.Int arithmetic to avoid precision loss

**Decision**: Calculate `amountMin = amountDesired * (100 - slippagePct) / 100` using big.Int multiplication and division

**Rationale**:
- Simple, auditable calculation
- Preserves precision with big.Int arithmetic
- Standard approach used in DEX interfaces (Uniswap, etc.)

**Alternatives Considered**:
1. **Fixed percentage (e.g., always 2%)**: Not configurable (rejected - violates spec FR-015)
2. **Floating point calculation**: Risk of precision loss (rejected - can cause transaction failures)
3. **Big.Int percentage calculation**: Precise and safe (selected)

**Implementation**:
```go
func calculateMinAmount(amountDesired *big.Int, slippagePct int) *big.Int {
    // amountMin = amountDesired * (100 - slippagePct) / 100
    multiplier := big.NewInt(int64(100 - slippagePct))
    divisor := big.NewInt(100)

    result := new(big.Int).Mul(amountDesired, multiplier)
    result.Div(result, divisor)

    return result
}
```

---

## Q4: Gas Cost Tracking

**Question**: How do we extract actual gas cost in native tokens from completed transactions?

**Research**:
- Examined pkg/types/transaction.go for TxReceipt structure
- Receipt contains: GasUsed (uint64) and EffectiveGasPrice (*big.Int)
- Actual cost in wei = GasUsed * EffectiveGasPrice
- This is native AVAX cost on Avalanche C-Chain
- Must convert GasUsed to *big.Int for multiplication

**Decision**: Extract gas cost from receipt as `gasCost = new(big.Int).SetUint64(receipt.GasUsed).Mul(gasCost, receipt.EffectiveGasPrice)`

**Rationale**:
- Accurate calculation using actual on-chain values
- Accounts for dynamic gas pricing (EIP-1559)
- Returns cost in wei (native token smallest unit)

**Alternatives Considered**:
1. **Use estimated gas**: Inaccurate, doesn't reflect actual cost (rejected)
2. **Use GasUsed * fixed price**: Doesn't account for EIP-1559 dynamics (rejected)
3. **GasUsed * EffectiveGasPrice**: Accurate actual cost (selected)

**Implementation**:
```go
func extractGasCost(receipt *types.TxReceipt) (*big.Int, error) {
    if receipt == nil {
        return nil, fmt.Errorf("receipt is nil")
    }

    gasUsed := new(big.Int).SetUint64(receipt.GasUsed)
    gasCost := new(big.Int).Mul(gasUsed, receipt.EffectiveGasPrice)

    return gasCost, nil
}
```

---

## Q5: Range Width Parameter Semantics

**Question**: How should the range width parameter map to actual tick bounds?

**Research**:
- Spec user scenarios show "range width parameter of 6 (±3 tick ranges from current)" (spec line 20)
- Spec acceptance scenario shows width=2 mapping to tickLower = 9800, tickUpper = 10200 with current tick 10000 (spec line 37)
- Tick spacing is 200 (CLAUDE.md line 165 assumption, verified in types.go comment line 38)
- Current implementation uses hardcoded `lpg = 3` (liquidity providing gap) in blackhole.go:141

**Decision**: Range width `N` means ±(N/2) tick ranges from current tick
- `tickLower = (currentTick / tickSpacing - rangeWidth/2) * tickSpacing`
- `tickUpper = (currentTick / tickSpacing + rangeWidth/2) * tickSpacing`

**Rationale**:
- Matches spec examples (width=6 → ±3 ranges, width=2 → ±1 range)
- Simple, intuitive parameter
- Allows fine-grained control (width=1 is minimum, width=20 is maximum per validation)

**Alternatives Considered**:
1. **Width = total number of ranges (width=6 means 6 ranges on each side)**: Too wide, contradicts spec examples (rejected)
2. **Width = tick offset directly**: Not aligned with tick spacing semantics (rejected)
3. **Width = ±(width/2) tick ranges**: Matches spec precisely (selected)

**Verification**:
- Example 1: currentTick=10000, width=6, spacing=200
  - tickLower = (10000/200 - 6/2) * 200 = (50 - 3) * 200 = 9400
  - tickUpper = (10000/200 + 6/2) * 200 = (50 + 3) * 200 = 10600
  - Range: ±3 tick ranges (600 ticks each direction) ✓

- Example 2: currentTick=10000, width=2, spacing=200
  - tickLower = (10000/200 - 2/2) * 200 = (50 - 1) * 200 = 9800
  - tickUpper = (10000/200 + 2/2) * 200 = (50 + 1) * 200 = 10200
  - Matches spec line 37 ✓

**Implementation**:
```go
func calculateTickBounds(currentTick int32, rangeWidth int, tickSpacing int) (int32, int32, error) {
    halfWidth := rangeWidth / 2
    tickIndex := int(currentTick) / tickSpacing

    tickLower := int32((tickIndex - halfWidth) * tickSpacing)
    tickUpper := int32((tickIndex + halfWidth) * tickSpacing)

    // Validate bounds
    const maxTick = 887272
    if tickLower < -maxTick || tickUpper > maxTick {
        return 0, 0, fmt.Errorf("tick bounds out of valid range: [%d, %d]", tickLower, tickUpper)
    }

    return tickLower, tickUpper, nil
}
```

---

## Q6: Input Validation Requirements

**Question**: What validation checks are required before executing staking operation?

**Research**:
- Constitutional Principle 5 requires fail-safe operation
- Spec FR-014 requires range width validation
- Spec FR-015 requires slippage tolerance validation
- Spec FR-005 requires balance verification
- Examined existing blackhole.go Mint method (line 139-180) which lacks validation

**Decision**: Implement comprehensive validation before any blockchain interaction

**Validation Checklist**:
1. **Range width**: Must be > 0 and <= 20 (prevents extremely wide ranges)
2. **Slippage tolerance**: Must be > 0 and <= 50 (prevents accidental 100% slippage acceptance)
3. **Token amounts**: Must be > 0
4. **Wallet balances**: Must be >= requested amounts for both tokens
5. **Tick bounds**: Must be within ±887272 (maximum valid tick range)

**Rationale**:
- Prevents common operator errors
- Fails fast before gas expenditure
- Provides clear error messages for troubleshooting
- Aligns with fail-safe principle

**Alternatives Considered**:
1. **Minimal validation (only check > 0)**: Insufficient, allows invalid operations (rejected)
2. **On-chain validation only**: Wastes gas on predictable failures (rejected)
3. **Comprehensive pre-flight validation**: Optimal UX and safety (selected)

**Implementation**:
```go
func validateStakingRequest(maxWAVAX, maxUSDC *big.Int, rangeWidth, slippagePct int) error {
    // Range width validation
    if rangeWidth <= 0 || rangeWidth > 20 {
        return fmt.Errorf("range width must be between 1 and 20, got %d", rangeWidth)
    }

    // Slippage validation
    if slippagePct <= 0 || slippagePct > 50 {
        return fmt.Errorf("slippage tolerance must be between 1 and 50 percent, got %d", slippagePct)
    }

    // Amount validation
    if maxWAVAX.Cmp(big.NewInt(0)) <= 0 {
        return fmt.Errorf("maxWAVAX must be > 0")
    }
    if maxUSDC.Cmp(big.NewInt(0)) <= 0 {
        return fmt.Errorf("maxUSDC must be > 0")
    }

    return nil
}

func (b *Blackhole) validateBalances(requiredWAVAX, requiredUSDC *big.Int) error {
    wavaxClient, _ := b.Client(wavax)
    usdcClient, _ := b.Client(usdc)

    wavaxBalance, err := wavaxClient.Call(&b.myAddr, "balanceOf", b.myAddr)
    if err != nil {
        return fmt.Errorf("failed to get WAVAX balance: %w", err)
    }

    usdcBalance, err := usdcClient.Call(&b.myAddr, "balanceOf", b.myAddr)
    if err != nil {
        return fmt.Errorf("failed to get USDC balance: %w", err)
    }

    if wavaxBalance[0].(*big.Int).Cmp(requiredWAVAX) < 0 {
        return fmt.Errorf("insufficient WAVAX balance: have %s, need %s",
            wavaxBalance[0].(*big.Int).String(), requiredWAVAX.String())
    }

    if usdcBalance[0].(*big.Int).Cmp(requiredUSDC) < 0 {
        return fmt.Errorf("insufficient USDC balance: have %s, need %s",
            usdcBalance[0].(*big.Int).String(), requiredUSDC.String())
    }

    return nil
}
```

---

## Summary

All research questions have been resolved. The implementation can proceed with:

1. **NonfungiblePositionManager**: 0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146
2. **Approval strategy**: Check allowance first, approve only if needed
3. **Slippage calculation**: `amountMin = amountDesired * (100 - slippagePct) / 100`
4. **Gas tracking**: `GasUsed * EffectiveGasPrice` from receipt
5. **Range width**: ±(width/2) tick ranges from current tick
6. **Validation**: Comprehensive pre-flight checks on all inputs and balances

These decisions satisfy all constitutional principles and functional requirements from the spec.
