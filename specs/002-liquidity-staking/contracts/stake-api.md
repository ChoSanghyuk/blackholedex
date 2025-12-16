# API Contract: Stake Method

**Feature**: 002-liquidity-staking
**Date**: 2025-12-16
**Type**: Go Method Signature
**Purpose**: Define the public interface for staking liquidity position NFTs in gauge contracts

## Method Signature

```go
func (b *Blackhole) Stake(
    nftTokenID *big.Int,
    gaugeAddress common.Address,
) (*StakingResult, error)
```

## Description

Stakes a liquidity position NFT (created by previous `Mint` operation) into a GaugeV2 contract to earn additional gauge rewards beyond trading fees. The method handles NFT approval (if needed) and deposit in a single atomic workflow with complete transaction tracking.

## Parameters

### Input Parameters

#### `nftTokenID` - `*big.Int` (required)
- **Description**: The ERC721 token ID representing the liquidity position to stake
- **Source**: Returned by previous `Mint()` call in `StakingResult.NFTTokenID`
- **Constraints**:
  - Must be > 0
  - Must exist in NonfungiblePositionManager contract
  - Must be owned by the caller's wallet address
  - Must not already be staked in any gauge
- **Example**: `big.NewInt(12345)`

#### `gaugeAddress` - `common.Address` (required)
- **Description**: The GaugeV2 contract address where the NFT will be staked
- **Source**: Configuration (typically hardcoded for WAVAX/USDC pool gauge)
- **Constraints**:
  - Must be non-zero address
  - Should correspond to a valid GaugeV2 contract (basic validation only)
  - Must match the pool for which the NFT was minted (WAVAX/USDC)
- **Example**: `common.HexToAddress("0x...")`

## Return Values

### Success Case - `(*StakingResult, nil)`

Returns a populated `StakingResult` structure with:

```go
&StakingResult{
    NFTTokenID:     nftTokenID,              // Input token ID
    ActualAmount0:  big.NewInt(0),           // Not populated by Stake
    ActualAmount1:  big.NewInt(0),           // Not populated by Stake
    FinalTickLower: 0,                       // Not populated by Stake
    FinalTickUpper: 0,                       // Not populated by Stake
    Transactions:   []TransactionRecord{...}, // 1-2 records (approval + deposit)
    TotalGasCost:   totalGas,                // Sum of all transaction gas costs
    Success:        true,
    ErrorMessage:   "",
}
```

**Transactions Array Contents**:
- If approval was needed: `[ApproveNFT, DepositNFT]` (2 records)
- If approval existed: `[DepositNFT]` (1 record)

Each `TransactionRecord` contains:
- `TxHash`: Transaction hash
- `GasUsed`: Gas consumed
- `GasPrice`: Effective gas price (wei)
- `GasCost`: Total cost = GasUsed * GasPrice
- `Timestamp`: Confirmation time
- `Operation`: "ApproveNFT" or "DepositNFT"

### Failure Cases - `(*StakingResult, error)`

Returns partial `StakingResult` with `Success=false` and descriptive error:

| Error Scenario | ErrorMessage Pattern | Transactions Array | NFT State |
|----------------|---------------------|-------------------|-----------|
| Invalid NFT token ID | "validation failed: invalid token ID" | Empty | N/A |
| NFT not owned by user | "validation failed: NFT not owned by wallet" | Empty | In user wallet (unchanged) |
| Invalid gauge address | "validation failed: invalid gauge address" | Empty | In user wallet |
| Insufficient gas balance | "failed to approve NFT: insufficient funds" | Empty | In user wallet |
| Approval tx failed | "NFT approval transaction failed: [reason]" | [ApproveNFT with failed status] | In user wallet |
| Deposit tx failed | "deposit transaction failed: [reason]" | [ApproveNFT?, DepositNFT with failed status] | In user wallet |
| Network timeout | "transaction timeout: [context]" | [Partial records] | In user wallet |

**Note**: In all failure cases, the NFT remains in the user's wallet. No intermediate state exists where ownership is unclear.

## Preconditions

Before calling `Stake()`, the following must be true:

1. **NFT Existence**: Token ID must have been minted via prior `Mint()` call
2. **NFT Ownership**: Caller's wallet (`b.myAddr`) must own the NFT
3. **Gas Balance**: Caller's wallet must have sufficient native token for gas fees (typically ~0.001-0.01 AVAX)
4. **Network Connection**: RPC connection to Avalanche C-Chain must be available
5. **Gauge Readiness**: Gauge contract must be active and accepting deposits

## Postconditions

### On Success

1. **NFT Ownership Transferred**: NFT now owned by gauge contract
2. **Gauge Balance Updated**: User's balance in gauge increased by 1 position
3. **Rewards Earning**: Position immediately starts earning gauge rewards
4. **Transactions Recorded**: All transactions logged with gas costs
5. **Approval Granted**: NFT approval set to gauge address (if it wasn't already)

### On Failure

1. **NFT Ownership Unchanged**: NFT remains in user's wallet
2. **Partial Transactions Logged**: Any sent transactions recorded in result
3. **Gas May Be Consumed**: If approval succeeded but deposit failed, approval gas was spent
4. **Error Context Available**: ErrorMessage explains what failed and why

## Side Effects

### Blockchain State Changes

- **NonfungiblePositionManager Contract**:
  - `getApproved(tokenId)` may be updated to gauge address (if approval sent)
  - `ownerOf(tokenId)` updated from user to gauge (on successful deposit)

- **Gauge Contract**:
  - User's staked position count incremented
  - Total staked value in gauge incremented
  - User becomes eligible for reward distributions

### Local State Changes

- **Transaction Log**: TransactionRecord entries created in memory
- **Gas Tracking**: TotalGasCost accumulated across operations
- **Time Tracking**: Timestamps recorded for each transaction

### No Side Effects On

- **Liquidity Amounts**: Staking doesn't modify the underlying liquidity (still in pool)
- **Tick Bounds**: Position's price range unchanged
- **Trading Fees**: Position continues earning trading fees regardless of staking
- **Other NFTs**: Only the specified token ID is affected

## Usage Example

```go
// Assume previous Mint operation returned:
// mintResult.NFTTokenID = 12345

gaugeAddr := common.HexToAddress("0x1234...") // WAVAX/USDC gauge

// Stake the minted position
stakeResult, err := blackhole.Stake(
    mintResult.NFTTokenID,
    gaugeAddr,
)

if err != nil {
    log.Printf("Stake failed: %v", err)
    if stakeResult != nil {
        log.Printf("Error message: %s", stakeResult.ErrorMessage)
        log.Printf("Partial gas cost: %s wei", stakeResult.TotalGasCost)
    }
    return err
}

if !stakeResult.Success {
    log.Printf("Stake unsuccessful: %s", stakeResult.ErrorMessage)
    return fmt.Errorf("stake failed")
}

// Success - log transaction details
fmt.Printf("✓ NFT staked successfully\n")
fmt.Printf("  Token ID: %s\n", stakeResult.NFTTokenID)
fmt.Printf("  Total Gas Cost: %s wei\n", stakeResult.TotalGasCost)
for _, tx := range stakeResult.Transactions {
    fmt.Printf("  - %s: %s (gas: %s wei)\n",
        tx.Operation, tx.TxHash.Hex(), tx.GasCost.String())
}
```

## Performance Characteristics

### Time Complexity

- **Best Case** (approval exists): ~2-5 seconds
  - 1 read call (getApproved)
  - 1 write transaction (deposit)
  - 1 confirmation wait (~2 seconds on Avalanche)

- **Typical Case** (approval needed): ~4-10 seconds
  - 1 read call (getApproved)
  - 2 write transactions (approve + deposit)
  - 2 confirmation waits (~4 seconds on Avalanche)

- **Worst Case** (network congestion): Up to 60 seconds
  - Same operations as typical case
  - Extended confirmation times

### Gas Costs

- **Approval Transaction**: ~45,000-55,000 gas
- **Deposit Transaction**: ~150,000-200,000 gas
- **Total Cost** (at 25 nAVAX/gas): ~0.002-0.006 AVAX (~$0.10-$0.30 USD at $50/AVAX)

### Network I/O

- **Read Calls**: 2-3 (ownerOf, getApproved, optional gauge validation)
- **Write Transactions**: 1-2 (conditional approve + required deposit)
- **Receipt Queries**: 1-2 (matching write transactions)

## Error Handling

### Validation Errors (No Transactions Sent)

```go
// Example: NFT not owned by user
return &StakingResult{
    NFTTokenID:   nftTokenID,
    Success:      false,
    ErrorMessage: "validation failed: NFT not owned by wallet",
}, fmt.Errorf("validation failed: NFT not owned by wallet")
```

### Transaction Errors (Partial State)

```go
// Example: Approval succeeded but deposit failed
return &StakingResult{
    NFTTokenID:   nftTokenID,
    Transactions: []TransactionRecord{approvalRecord},
    TotalGasCost: approvalRecord.GasCost,
    Success:      false,
    ErrorMessage: "deposit transaction failed: execution reverted",
}, fmt.Errorf("deposit transaction failed: %w", err)
```

## Thread Safety

**Not Thread-Safe**: This method is not safe for concurrent calls with the same wallet/NFT.

**Reasons**:
- Reads and modifies blockchain state (NFT ownership, approval)
- No internal locking mechanism
- Race conditions possible if multiple goroutines call with same NFT

**Recommendation**: Caller must ensure sequential execution per NFT token ID.

## Dependencies

### Internal Dependencies

- `b.Client(address)` - Get ContractClient for address
- `b.myAddr` - Caller's wallet address
- `b.privateKey` - Wallet's private key for signing
- `b.tl.WaitForTransaction()` - Transaction confirmation
- `util.ExtractGasCost()` - Gas cost calculation

### External Dependencies

- NonfungiblePositionManager contract at `0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146`
- GaugeV2 contract at specified `gaugeAddress`
- Avalanche C-Chain RPC endpoint
- Sufficient AVAX balance for gas

## Testing Considerations

### Unit Test Cases

1. **Valid stake (approval needed)**: Verify 2 transactions, correct gas tracking
2. **Valid stake (approval exists)**: Verify 1 transaction, skipped approval
3. **Invalid token ID**: Verify early return, no transactions sent
4. **NFT not owned**: Verify validation failure, no transactions sent
5. **Approval failure**: Verify error, NFT ownership unchanged
6. **Deposit failure**: Verify error, NFT still owned by user

### Integration Test Cases

1. **Full workflow**: Mint → Stake → Verify gauge balance
2. **Approve reuse**: Stake → Unstake → Stake again (approval persists)
3. **Network timeout**: Simulate RPC failure during deposit
4. **Gas tracking accuracy**: Compare tracked costs to on-chain receipts

### Edge Cases

1. **Zero token ID**: Should fail validation
2. **Zero gauge address**: Should fail validation
3. **NFT already staked**: ownerOf returns gauge address, validation fails
4. **Gauge paused**: Deposit transaction reverts
5. **Concurrent stake attempts**: Race condition (not handled - caller's responsibility)

## Constitutional Compliance

This API design adheres to project constitution:

- **Principle 1 (Scope)**: Only WAVAX/USDC pool NFTs, only Blackhole DEX gauges
- **Principle 3 (Transparency)**: Complete transaction and gas tracking
- **Principle 4 (Gas Optimization)**: Approval reuse, automatic estimation
- **Principle 5 (Safety)**: NFT never in intermediate state, comprehensive error handling

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-12-16 | Initial API design for 002-liquidity-staking |

## Related Methods

- `Mint()` - Creates the NFT position that will be staked
- `Unstake()` - (Future) Withdraws NFT from gauge back to wallet
- `ClaimRewards()` - (Future) Collects earned gauge rewards
