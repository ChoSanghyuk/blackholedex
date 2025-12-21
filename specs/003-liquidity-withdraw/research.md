# Research: Liquidity Position Withdrawal

**Feature**: 003-liquidity-withdraw
**Date**: 2025-12-21
**Status**: Complete

## Overview

This document captures research findings for implementing the `Withdraw` function that exits liquidity positions created by the `Mint` function. The research focuses on understanding the NonfungiblePositionManager contract interface, multicall transaction patterns, and integration with existing codebase patterns.

## Research Areas

### 1. NonfungiblePositionManager Contract Interface

**Decision**: Use INonfungiblePositionManager ABI with three key functions: `decreaseLiquidity`, `collect`, and `burn`

**Rationale**:
- These three functions must be called in sequence to fully withdraw a position
- `decreaseLiquidity` removes liquidity but doesn't transfer tokens
- `collect` actually transfers the tokens and fees to the recipient
- `burn` destroys the NFT after all tokens are collected
- All three are `payable` functions that can be batched in a multicall

**Alternatives considered**:
- Individual function calls: Rejected due to gas inefficiency and lack of atomicity
- Using only `collect` and `burn`: Rejected because liquidity must first be decreased to 0

**ABI Details**:

```solidity
// Get position details before withdrawal
function positions(uint256 tokenId) external view returns (
    uint88 nonce,
    address operator,
    address token0,
    address token1,
    address deployer,
    int24 tickLower,
    int24 tickUpper,
    uint128 liquidity,        // Current liquidity - needed for decreaseLiquidity
    uint256 feeGrowthInside0LastX128,
    uint256 feeGrowthInside1LastX128,
    uint128 tokensOwed0,      // Accumulated fees
    uint128 tokensOwed1
);

// Step 1: Remove all liquidity
struct DecreaseLiquidityParams {
    uint256 tokenId;
    uint128 liquidity;        // Must match position.liquidity for full withdrawal
    uint256 amount0Min;       // Slippage protection
    uint256 amount1Min;       // Slippage protection
    uint256 deadline;
}
function decreaseLiquidity(DecreaseLiquidityParams params) external payable returns (uint256 amount0, uint256 amount1);

// Step 2: Collect tokens and fees
struct CollectParams {
    uint256 tokenId;
    address recipient;
    uint128 amount0Max;       // Use type(uint128).max to collect all
    uint128 amount1Max;       // Use type(uint128).max to collect all
}
function collect(CollectParams params) external payable returns (uint256 amount0, uint256 amount1);

// Step 3: Burn the NFT
function burn(uint256 tokenId) external payable;

// Batch execution
function multicall(bytes[] calldata data) external payable returns (bytes[] memory results);
```

### 2. Multicall Transaction Pattern

**Decision**: Encode all three operations (decreaseLiquidity, collect, burn) into a single multicall transaction

**Rationale**:
- Ensures atomic execution - either all operations succeed or all revert
- Minimizes gas costs by combining operations
- Matches existing pattern in Unstake function which uses multicall for farming operations
- Prevents intermediate states where liquidity is removed but NFT still exists

**Alternatives considered**:
- Sequential individual transactions: Rejected due to non-atomicity and higher gas costs
- Custom contract for withdrawal: Rejected as overkill for simple operation sequence

**Implementation approach**:
```go
// 1. Get NFT position details to determine liquidity amount
positionResult := nftManagerClient.Call(&b.myAddr, "positions", nftTokenID)
liquidity := positionResult[7].(uint128)

// 2. Build multicall data
var multicallData [][]byte

// Encode decreaseLiquidity
decreaseParams := DecreaseLiquidityParams{
    TokenId:    nftTokenID,
    Liquidity:  liquidity,  // Full withdrawal
    Amount0Min: calculateMin(amount0Expected, slippagePct),
    Amount1Min: calculateMin(amount1Expected, slippagePct),
    Deadline:   deadline,
}
decreaseData, _ := abi.Pack("decreaseLiquidity", decreaseParams)
multicallData = append(multicallData, decreaseData)

// Encode collect
collectParams := CollectParams{
    TokenId:    nftTokenID,
    Recipient:  b.myAddr,
    Amount0Max: type(uint128).max,  // Collect all
    Amount1Max: type(uint128).max,
}
collectData, _ := abi.Pack("collect", collectParams)
multicallData = append(multicallData, collectData)

// Encode burn
burnData, _ := abi.Pack("burn", nftTokenID)
multicallData = append(multicallData, burnData)

// 3. Execute multicall
txHash := nftManagerClient.Send(..., "multicall", multicallData)
```

### 3. Result Type Design

**Decision**: Create `WithdrawResult` struct following the pattern of `StakingResult` and `UnstakeResult`

**Rationale**:
- Maintains consistency with existing codebase
- Provides comprehensive transaction tracking for Principle 3 (Financial Transparency)
- Includes gas cost details, success status, and error messages

**Alternatives considered**:
- Reuse `UnstakeResult`: Rejected because withdrawal is conceptually different (no farming rewards)
- Minimal return type (just error): Rejected because violates transparency principle

**Type definition**:
```go
type WithdrawResult struct {
    NFTTokenID     *big.Int            // Withdrawn NFT token ID
    Amount0        *big.Int            // WAVAX withdrawn (wei)
    Amount1        *big.Int            // USDC withdrawn (smallest unit)
    Transactions   []TransactionRecord // All transactions executed
    TotalGasCost   *big.Int            // Sum of all gas costs (wei)
    Success        bool                // Whether operation succeeded
    ErrorMessage   string              // Error message if failed
}

// CollectParams for collect operation
type CollectParams struct {
    TokenId    *big.Int       `json:"tokenId"`
    Recipient  common.Address `json:"recipient"`
    Amount0Max *big.Int       `json:"amount0Max"`  // Use MaxUint128
    Amount1Max *big.Int       `json:"amount1Max"`  // Use MaxUint128
}

// DecreaseLiquidityParams for decreaseLiquidity operation
type DecreaseLiquidityParams struct {
    TokenId    *big.Int `json:"tokenId"`
    Liquidity  *big.Int `json:"liquidity"`   // uint128
    Amount0Min *big.Int `json:"amount0Min"`
    Amount1Min *big.Int `json:"amount1Min"`
    Deadline   *big.Int `json:"deadline"`
}
```

### 4. Slippage Protection

**Decision**: Calculate minimum token amounts based on expected amounts with configurable slippage tolerance (default 5%)

**Rationale**:
- Protects against sandwich attacks and price manipulation
- Follows existing pattern in Mint function
- Matches constitutional Principle 5 (Fail-Safe Operation)

**Alternatives considered**:
- No slippage protection: Rejected as unsafe
- Fixed slippage percentage: Rejected as inflexible for volatile market conditions

**Implementation**:
```go
// Use existing util.CalculateMinAmount pattern from Mint
amount0Min := util.CalculateMinAmount(amount0Expected, slippagePct)
amount1Min := util.CalculateMinAmount(amount1Expected, slippagePct)
```

### 5. Error Handling Strategy

**Decision**: Validate ownership and position state before executing multicall, return detailed error context

**Rationale**:
- Prevents wasted gas on failed transactions
- Provides clear diagnostic information
- Matches patterns in Stake and Unstake functions

**Alternatives considered**:
- Let blockchain validate: Rejected due to wasted gas and poor error messages
- Minimal validation: Rejected because violates fail-safe principle

**Validation sequence**:
1. Input validation (NFT ID > 0)
2. Ownership verification (ownerOf call)
3. Position state check (positions call to get liquidity)
4. Multicall execution with timeout

### 6. Integration with Existing Code

**Decision**: Add `Withdraw` method to `Blackhole` struct, reuse existing ContractClient infrastructure

**Rationale**:
- Consistent with existing Mint, Stake, Unstake methods
- Leverages existing TxListener for transaction confirmation
- Uses existing util package for validation and gas extraction

**Alternatives considered**:
- Standalone function: Rejected as inconsistent with codebase architecture
- New package: Rejected as unnecessary complexity

**Dependencies**:
- Existing: `blackhole.Blackhole` struct, `ContractClient` interface, `TxListener` interface
- Existing: `util.ExtractGasCost`, `util.CalculateMinAmount`
- New: `WithdrawResult`, `CollectParams`, `DecreaseLiquidityParams` types

## Key Findings Summary

1. **Multicall is essential** for atomic withdrawal execution
2. **Three-step process** is mandatory: decreaseLiquidity → collect → burn
3. **Position query** must precede withdrawal to determine liquidity amount
4. **Slippage protection** required via amount0Min/amount1Min
5. **Gas tracking** maintains constitutional transparency requirements
6. **Ownership validation** prevents wasted gas on invalid operations

## Testing Strategy

1. **Unit tests**: Validate input parameters, slippage calculations
2. **Integration tests**: Execute full withdrawal on testnet with real NFT positions
3. **Edge case tests**:
   - Withdrawing already burned NFT
   - Withdrawing NFT owned by different address
   - Position with zero liquidity
   - Network failures during multicall

## References

- NonfungiblePositionManager ABI: `blackholedex-contracts/artifacts/@cryptoalgebra/integral-periphery/contracts/interfaces/INonfungiblePositionManager.sol/`
- Existing Mint implementation: `blackhole.go:231-538`
- Existing Unstake multicall pattern: `blackhole.go:786-981`
- Constitutional principles: `.specify/memory/constitution.md`
