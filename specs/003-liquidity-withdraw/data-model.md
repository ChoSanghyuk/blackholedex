# Data Model: Liquidity Position Withdrawal

**Feature**: 003-liquidity-withdraw
**Date**: 2025-12-21

## Overview

This document defines the data structures required for withdrawing liquidity positions. The model follows existing patterns in the codebase (StakingResult, UnstakeResult) and ensures compliance with constitutional transparency requirements.

## Core Entities

### WithdrawResult

**Purpose**: Represents the complete outcome of a withdrawal operation

**Attributes**:
- `NFTTokenID` (*big.Int): The NFT token ID that was withdrawn
- `Amount0` (*big.Int): Amount of token0 (WAVAX) withdrawn in wei
- `Amount1` (*big.Int): Amount of token1 (USDC) withdrawn in smallest unit
- `Transactions` ([]TransactionRecord): Detailed record of all blockchain transactions
- `TotalGasCost` (*big.Int): Sum of gas costs across all transactions in wei
- `Success` (bool): Whether the withdrawal completed successfully
- `ErrorMessage` (string): Human-readable error description (empty if success)

**Relationships**:
- Contains multiple TransactionRecord instances
- Created by Blackhole.Withdraw method
- Returned to caller for reporting and logging

**Validation Rules**:
- NFTTokenID must be positive (validated before execution)
- Success is true only if all operations complete
- ErrorMessage populated only when Success is false
- TotalGasCost is sum of all Transaction gas costs

**State Transitions**:
```
Initial → Validating (ownership check)
Validating → Executing (multicall submitted)
Executing → Success (multicall confirmed, NFT burned)
Executing → Failed (validation error, transaction revert, or network failure)
```

### DecreaseLiquidityParams

**Purpose**: Parameters for the decreaseLiquidity operation in the multicall

**Attributes**:
- `TokenId` (*big.Int): NFT position token ID
- `Liquidity` (*big.Int): Amount of liquidity to remove (uint128, always full position for Withdraw)
- `Amount0Min` (*big.Int): Minimum token0 to receive (slippage protection)
- `Amount1Min` (*big.Int): Minimum token1 to receive (slippage protection)
- `Deadline` (*big.Int): Unix timestamp after which transaction becomes invalid

**Relationships**:
- Encoded into multicall data array (first operation)
- Parameters derived from positions() query and slippage calculations

**Validation Rules**:
- TokenId must be positive
- Liquidity must match full position liquidity (from positions query)
- Amount0Min/Amount1Min calculated as (expected - slippage%)
- Deadline must be future timestamp (typically now + 20 minutes)

### CollectParams

**Purpose**: Parameters for the collect operation in the multicall

**Attributes**:
- `TokenId` (*big.Int): NFT position token ID
- `Recipient` (common.Address): Address to receive withdrawn tokens (user's wallet)
- `Amount0Max` (*big.Int): Maximum token0 to collect (uint128, use MaxUint128 for all)
- `Amount1Max` (*big.Int): Maximum token1 to collect (uint128, use MaxUint128 for all)

**Relationships**:
- Encoded into multicall data array (second operation)
- Follows decreaseLiquidity in execution order

**Validation Rules**:
- TokenId must match decreaseLiquidity TokenId
- Recipient must be valid address (user's wallet)
- Amount0Max/Amount1Max set to MaxUint128 to collect all available tokens

### TransactionRecord (Existing)

**Purpose**: Tracks individual blockchain transaction details

**Attributes**:
- `TxHash` (common.Hash): Transaction hash for blockchain lookup
- `GasUsed` (uint64): Actual gas consumed by transaction
- `GasPrice` (*big.Int): Effective gas price in wei
- `GasCost` (*big.Int): Total cost (GasUsed * GasPrice) in wei
- `Timestamp` (time.Time): When transaction was executed
- `Operation` (string): Human-readable operation type

**Relationships**:
- Part of WithdrawResult.Transactions array
- One record per blockchain transaction (typically 1 for multicall)

**Operation Types**:
- "Withdraw": Multicall transaction containing decreaseLiquidity, collect, burn

### Position (Read-Only)

**Purpose**: Represents NFT position state queried from blockchain

**Attributes** (from positions() view function):
- `Nonce` (uint88): Position nonce
- `Operator` (address): Approved operator
- `Token0` (address): First token (WAVAX)
- `Token1` (address): Second token (USDC)
- `Deployer` (address): Pool deployer address
- `TickLower` (int24): Lower tick boundary
- `TickUpper` (int24): Upper tick boundary
- `Liquidity` (uint128): Current liquidity amount (needed for withdrawal)
- `FeeGrowthInside0LastX128` (uint256): Fee tracking
- `FeeGrowthInside1LastX128` (uint256): Fee tracking
- `TokensOwed0` (uint128): Accumulated fees in token0
- `TokensOwed1` (uint128): Accumulated fees in token1

**Relationships**:
- Queried before withdrawal to get liquidity amount
- Not persisted, only used for calculation

**Usage**:
- Liquidity field used to populate DecreaseLiquidityParams.Liquidity
- TokensOwed0/TokensOwed1 indicate accumulated fees to be collected

## Data Flow

### Withdrawal Operation Flow

```
1. Input: NFT Token ID
   ↓
2. Query Position:
   positions(tokenId) → Position{liquidity, tokensOwed0, tokensOwed1, ...}
   ↓
3. Build Parameters:
   - DecreaseLiquidityParams{tokenId, liquidity, amount0Min, amount1Min, deadline}
   - CollectParams{tokenId, recipient, MaxUint128, MaxUint128}
   - BurnParams{tokenId}
   ↓
4. Encode Multicall:
   multicallData = [decreaseLiquidityData, collectData, burnData]
   ↓
5. Execute Transaction:
   nftManagerClient.Send("multicall", multicallData) → txHash
   ↓
6. Wait for Confirmation:
   TxListener.WaitForTransaction(txHash) → receipt
   ↓
7. Extract Results:
   - Parse receipt for gas costs
   - Create TransactionRecord
   - Decode multicall results for amounts
   ↓
8. Return: WithdrawResult
```

### Error Handling Flow

```
Validation Error → WithdrawResult{Success: false, ErrorMessage: "..."}
Transaction Revert → WithdrawResult{Success: false, ErrorMessage: "...", Transactions: [...]}
Network Failure → WithdrawResult{Success: false, ErrorMessage: "timeout..."}
Success → WithdrawResult{Success: true, Amount0: X, Amount1: Y, ...}
```

## Type Mappings

### Solidity to Go Type Conversions

| Solidity Type | Go Type | Notes |
|---------------|---------|-------|
| uint256 | *big.Int | Standard large integer |
| uint128 | *big.Int | Stored as big.Int, validated in range |
| address | common.Address | Ethereum address (20 bytes) |
| int24 | int32 | Tick values, stored as int32 |
| bytes[] | [][]byte | Multicall data array |

### Constants

```go
// Maximum uint128 value for collecting all tokens
MaxUint128 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))

// Default deadline offset
DefaultDeadlineMinutes = 20
```

## Storage Considerations

**No Persistent Storage Required**:
- All state exists on blockchain (NFT position, token balances)
- WithdrawResult is ephemeral (returned to caller, optionally logged)
- TransactionRecords are created on-the-fly from receipts

**Memory Lifecycle**:
1. Function entry: Allocate WithdrawResult, TransactionRecord slice
2. Query phase: Allocate Position struct (discarded after use)
3. Encode phase: Allocate multicall byte arrays (garbage collected)
4. Return: WithdrawResult returned to caller

## Validation Summary

### Pre-Execution Validation
- NFT token ID > 0
- User owns the NFT (ownerOf check)
- Position exists and has liquidity (positions check)

### Transaction Validation
- Multicall execution within deadline
- Slippage bounds respected (amount0Min, amount1Min)
- NFT successfully burned (verified post-transaction)

### Post-Execution Validation
- Transaction confirmed on blockchain
- Gas costs calculated correctly
- All tokens transferred to recipient

## Constitutional Compliance

This data model supports constitutional principles:

- **Principle 1 (Scope)**: Only operates on WAVAX/USDC positions (token0/token1 addresses validated)
- **Principle 3 (Transparency)**: TransactionRecord tracks all gas costs and operations
- **Principle 4 (Optimization)**: Multicall batches operations, minimizing transactions
- **Principle 5 (Safety)**: Validation prevents invalid operations, slippage protection included
