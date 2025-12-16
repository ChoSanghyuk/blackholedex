# Data Model: Liquidity Position Staking

**Feature**: 002-liquidity-staking
**Date**: 2025-12-16
**Purpose**: Define data structures and state transitions for NFT staking operations

## Entity Overview

This feature reuses existing data structures from the Mint operation (`StakingResult` and `TransactionRecord`) but uses them in a different workflow context. No new entities are required.

---

## Entity 1: StakingResult (Existing - Reused)

**Purpose**: Comprehensive output structure for staking operations, tracking success/failure and all financial costs

**Source**: `types.go` lines 175-186

### Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `NFTTokenID` | `*big.Int` | Liquidity position NFT token ID | Required, > 0 |
| `ActualAmount0` | `*big.Int` | Actual WAVAX staked (wei) | Not modified by Stake (set by Mint) |
| `ActualAmount1` | `*big.Int` | Actual USDC staked (smallest unit) | Not modified by Stake (set by Mint) |
| `FinalTickLower` | `int32` | Final lower tick bound | Not modified by Stake (set by Mint) |
| `FinalTickUpper` | `int32` | Final upper tick bound | Not modified by Stake (set by Mint) |
| `Transactions` | `[]TransactionRecord` | All transactions executed (approvals + deposit) | 1-2 records for Stake operation |
| `TotalGasCost` | `*big.Int` | Sum of all gas costs (wei) | Sum of Transactions[*].GasCost |
| `Success` | `bool` | Whether operation succeeded | true on success, false on any error |
| `ErrorMessage` | `string` | Error message if failed | Empty on success, descriptive on failure |

### Usage in Stake Operation

The `Stake` method returns a `StakingResult` with these field semantics:

- **NFTTokenID**: Set to the token ID being staked (input parameter)
- **ActualAmount0/1**: Not populated (remains zero) since Stake doesn't modify liquidity amounts
- **FinalTickLower/Upper**: Not populated (remains zero) since Stake doesn't modify tick bounds
- **Transactions**: Contains 1-2 records:
  - If approval needed: [ApproveNFT, DepositNFT]
  - If approval existed: [DepositNFT]
- **TotalGasCost**: Sum of approval gas (if sent) + deposit gas
- **Success**: true if deposit confirmed, false if any step failed
- **ErrorMessage**: Empty on success, contains error context on failure

### State Transitions

```
Initial State → Validation → Approval (conditional) → Deposit → Final State

1. Initial State:
   - NFT owned by user wallet
   - NFT may or may not be approved for gauge
   - Gauge contract ready to accept deposits

2. Validation State:
   - Check NFT ownership (ownerOf call)
   - Check gauge address validity
   - If invalid: transition to Failed State

3. Approval Check State:
   - Query getApproved(tokenId)
   - If approved for gauge: skip to Deposit State
   - If not approved: transition to Approval State

4. Approval State:
   - Submit approve(gauge, tokenId) transaction
   - Wait for confirmation
   - If failed: transition to Failed State (NFT still owned by user)
   - If confirmed: transition to Deposit State

5. Deposit State:
   - Submit deposit(tokenId) transaction to gauge
   - Wait for confirmation
   - If failed: transition to Failed State (NFT still owned by user)
   - If confirmed: transition to Success State

6. Success State:
   - NFT owned by gauge contract
   - User's position earning gauge rewards
   - All transactions recorded

7. Failed State:
   - NFT owned by user wallet (never transferred)
   - Partial transaction records logged
   - Error message explains failure
```

---

## Entity 2: TransactionRecord (Existing - Reused)

**Purpose**: Immutable record of blockchain transaction with full gas cost tracking

**Source**: `types.go` lines 166-173

### Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `TxHash` | `common.Hash` | Transaction hash | Required, non-zero |
| `GasUsed` | `uint64` | Gas consumed by transaction | > 0 |
| `GasPrice` | `*big.Int` | Effective gas price (wei) | > 0 |
| `GasCost` | `*big.Int` | Total gas cost = GasUsed * GasPrice | = GasUsed * GasPrice |
| `Timestamp` | `time.Time` | Transaction confirmation time | Real-time on receipt |
| `Operation` | `string` | Operation type identifier | "ApproveNFT" or "DepositNFT" |

### Operation Type Values for Stake

- **"ApproveNFT"**: Transaction approving NFT transfer to gauge
- **"DepositNFT"**: Transaction depositing NFT into gauge

### Record Creation Pattern

```go
// After each transaction confirmation:
1. Extract receipt from TxListener.WaitForTransaction(txHash)
2. Extract gas cost using util.ExtractGasCost(receipt)
3. Parse receipt.EffectiveGasPrice → *big.Int
4. Parse receipt.GasUsed → *big.Int → uint64
5. Create TransactionRecord{
     TxHash:    txHash,
     GasUsed:   gasUsed.Uint64(),
     GasPrice:  gasPrice,
     GasCost:   gasCost,
     Timestamp: time.Now(),
     Operation: "ApproveNFT" or "DepositNFT",
   }
6. Append to StakingResult.Transactions
```

---

## Entity 3: NFT Position Token (External - Read-Only)

**Purpose**: Represents liquidity position in WAVAX/USDC pool, implemented as ERC721 NFT

**Ownership**: User wallet initially, transferred to gauge after successful stake

### On-Chain Properties (Read-Only)

| Property | Query Method | Description |
|----------|--------------|-------------|
| Token ID | Input parameter | Unique identifier for liquidity position |
| Owner | `ownerOf(tokenId)` | Current owner address (user or gauge) |
| Approved Address | `getApproved(tokenId)` | Address authorized to transfer token |
| Liquidity Amount | Not queried in Stake | Determined at mint time, not modified |
| Tick Range | Not queried in Stake | Determined at mint time, not modified |

### Validation Rules

- **Existence**: Token ID must exist (ownerOf doesn't revert)
- **Ownership**: Owner must match user's wallet address before staking
- **Approval**: Approval must match gauge address before deposit
- **No re-staking**: Cannot stake an NFT already owned by a gauge

---

## Entity 4: Gauge Contract (External - Interaction Point)

**Purpose**: Accepts staked NFT positions and distributes rewards to stakers

**Type**: GaugeV2 smart contract

### Interface Methods Used

| Method | Parameters | Returns | Purpose |
|--------|------------|---------|---------|
| `deposit` | `uint256 amount` (token ID) | void | Transfer NFT to gauge and start earning |

### State Changes

**Before Deposit**:
- NFT owned by user
- User's gauge balance = 0

**After Deposit**:
- NFT owned by gauge
- User's gauge balance = 1 position
- User earning gauge rewards on this position

---

## Data Flow Diagram

```
Input: nftTokenID (*big.Int), gaugeAddress (common.Address)
  │
  ├─> Validation
  │     ├─> ownerOf(nftTokenID) → userAddress
  │     └─> gaugeAddress != 0x0
  │
  ├─> Approval Check
  │     ├─> getApproved(nftTokenID) → approvedAddress
  │     └─> if approvedAddress != gaugeAddress:
  │           ├─> approve(gaugeAddress, nftTokenID) → txHash1
  │           ├─> WaitForTransaction(txHash1) → receipt1
  │           └─> Create TransactionRecord("ApproveNFT")
  │
  ├─> Deposit
  │     ├─> gaugeClient.Send("deposit", nftTokenID) → txHash2
  │     ├─> WaitForTransaction(txHash2) → receipt2
  │     └─> Create TransactionRecord("DepositNFT")
  │
  └─> Output: StakingResult
        ├─> NFTTokenID = nftTokenID
        ├─> Transactions = [ApproveNFT?, DepositNFT]
        ├─> TotalGasCost = sum(GasCost)
        ├─> Success = true
        └─> ErrorMessage = ""

On Error at Any Stage:
  └─> Output: StakingResult
        ├─> NFTTokenID = nftTokenID (or 0 if before validation)
        ├─> Transactions = [partial records if any tx sent]
        ├─> TotalGasCost = sum(GasCost) [may be 0]
        ├─> Success = false
        └─> ErrorMessage = "context: error details"
```

---

## Invariants

These conditions must always hold true:

1. **NFT Ownership**: NFT is either owned by user (pre-stake or failed) OR gauge (post-stake success), never in intermediate state
2. **Approval Safety**: Approval only granted to specific gauge for specific token ID (never all tokens)
3. **Transaction Completeness**: Every sent transaction appears in Transactions array with complete gas data
4. **Gas Cost Accuracy**: TotalGasCost exactly equals sum of all TransactionRecord.GasCost values
5. **Success Consistency**: Success=true implies deposit transaction confirmed; Success=false implies NFT still owned by user
6. **Error Clarity**: Success=false always has non-empty ErrorMessage
7. **No Duplicate Stakes**: Cannot stake an NFT that is already owned by a gauge

---

## Assumptions

- **Single-threaded execution**: One stake operation per user at a time (no concurrency)
- **Gauge permanence**: Gauge contract address doesn't change during operation
- **Network availability**: RPC connection remains available for transaction confirmation
- **Sufficient gas**: User wallet has enough native token to pay gas fees
- **NFT hasn't moved**: NFT remains in user's wallet between ownership check and deposit (no concurrent transfers)

---

## Future Enhancements (Out of Scope)

- **Batch staking**: Stake multiple NFTs in single operation
- **Unstaking**: Withdraw NFT from gauge back to wallet
- **Reward claiming**: Collect earned gauge rewards
- **Position history**: Persistent tracking of all stake/unstake events
- **Multi-gauge support**: Stake same position across multiple gauges (if protocol supports)
