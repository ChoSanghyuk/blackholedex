# API Contract: Withdraw Function

**Feature**: 003-liquidity-withdraw
**Date**: 2025-12-21
**Type**: Go Method Interface

## Overview

This document defines the public interface for the `Withdraw` method on the `Blackhole` struct. This is not a REST or GraphQL API, but rather a Go library method contract that external callers must adhere to.

## Method Signature

```go
func (b *Blackhole) Withdraw(nftTokenID *big.Int) (*WithdrawResult, error)
```

## Parameters

### Input: `nftTokenID`

**Type**: `*big.Int`

**Description**: The NFT token ID representing the liquidity position to withdraw. This ID was returned by a previous `Mint` operation.

**Constraints**:
- MUST be non-nil
- MUST be greater than 0
- MUST represent an NFT owned by the caller's wallet
- MUST be a valid NFT that exists (not already burned)

**Example Values**:
```go
// Valid
nftTokenID := big.NewInt(12345)

// Invalid - will fail validation
nftTokenID := big.NewInt(0)         // Zero
nftTokenID := big.NewInt(-1)        // Negative
nftTokenID := nil                   // Nil pointer
```

## Return Values

### Success Response: `WithdrawResult`

**Type**: `*WithdrawResult`

**Structure**:
```go
type WithdrawResult struct {
    NFTTokenID     *big.Int            // Withdrawn NFT token ID (matches input)
    Amount0        *big.Int            // WAVAX withdrawn (wei)
    Amount1        *big.Int            // USDC withdrawn (smallest unit)
    Transactions   []TransactionRecord // Transaction details
    TotalGasCost   *big.Int            // Total gas cost (wei)
    Success        bool                // Always true in success case
    ErrorMessage   string              // Always empty in success case
}
```

**Fields**:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `NFTTokenID` | `*big.Int` | Echoes the input NFT token ID | `12345` |
| `Amount0` | `*big.Int` | WAVAX withdrawn in wei | `1500000000000000000` (1.5 WAVAX) |
| `Amount1` | `*big.Int` | USDC withdrawn in smallest unit (6 decimals) | `2500000` (2.5 USDC) |
| `Transactions` | `[]TransactionRecord` | List of blockchain transactions (typically 1 multicall) | See TransactionRecord below |
| `TotalGasCost` | `*big.Int` | Sum of all gas costs in wei | `2500000000000000` (0.0025 AVAX) |
| `Success` | `bool` | Operation success indicator | `true` |
| `ErrorMessage` | `string` | Error description (empty on success) | `""` |

**TransactionRecord Structure**:
```go
type TransactionRecord struct {
    TxHash    common.Hash  // 0x1234...
    GasUsed   uint64       // 125000
    GasPrice  *big.Int     // 25000000000 (25 gwei)
    GasCost   *big.Int     // GasUsed * GasPrice
    Timestamp time.Time    // 2025-12-21 10:30:00
    Operation string       // "Withdraw"
}
```

### Error Response: `error`

**Type**: `error`

**When Returned**:
- Input validation fails (NFT ID invalid)
- Ownership verification fails (NFT not owned by caller)
- Position query fails (NFT doesn't exist)
- Transaction submission fails
- Transaction execution reverts
- Network/RPC errors

**Error Message Patterns**:
```go
// Validation errors
"validation failed: NFT token ID must be positive"

// Ownership errors
"NFT not owned by wallet: owned by 0xABC..."

// Position errors
"failed to verify NFT ownership: <error details>"
"failed to query position: <error details>"

// Transaction errors
"failed to encode multicall: <error details>"
"failed to submit multicall transaction: <error details>"
"multicall transaction failed: <error details>"

// Gas extraction errors
"failed to extract gas cost: <error details>"
```

**WithdrawResult on Error**:
When an error is returned, the `WithdrawResult` may contain partial information:
```go
type WithdrawResult struct {
    NFTTokenID     *big.Int            // Populated if provided
    Amount0        *big.Int            // Zero or nil
    Amount1        *big.Int            // Zero or nil
    Transactions   []TransactionRecord // May contain partial records
    TotalGasCost   *big.Int            // Partial gas if tx submitted
    Success        bool                // false
    ErrorMessage   string              // Populated with error details
}
```

## Behavioral Specifications

### Pre-Conditions

Before calling `Withdraw`, the following conditions MUST be met:

1. **Wallet Setup**: The `Blackhole` struct must be initialized with:
   - Valid private key
   - Wallet address (myAddr)
   - ContractClient map with nonfungiblePositionManager client
   - TxListener for transaction confirmation

2. **NFT Ownership**: The caller's wallet must own the NFT token ID

3. **Position State**: The NFT must represent a valid position with:
   - Liquidity > 0 (or position with only accumulated fees)
   - Not currently staked in a farming contract (must be unstaked first)

4. **Network**: Connection to Avalanche C-Chain RPC

### Post-Conditions

After successful execution:

1. **Liquidity Removed**: Position liquidity is reduced to 0
2. **Tokens Transferred**: All tokens (principal + fees) transferred to wallet
3. **NFT Burned**: NFT no longer exists (ownerOf will revert)
4. **Gas Paid**: Transaction gas deducted from wallet's AVAX balance
5. **Result Returned**: Complete WithdrawResult with all transaction details

After failed execution:

1. **State Unchanged**: NFT and position remain in original state (atomic revert)
2. **Error Returned**: Both error and WithdrawResult.ErrorMessage populated
3. **Partial Gas**: Gas may be consumed if transaction was submitted but reverted

## Usage Examples

### Example 1: Basic Withdrawal

```go
// Initialize Blackhole instance (assumed setup)
var blackhole *Blackhole

// NFT token ID from previous Mint operation
nftTokenID := big.NewInt(12345)

// Execute withdrawal
result, err := blackhole.Withdraw(nftTokenID)
if err != nil {
    log.Printf("Withdrawal failed: %v", err)
    if result != nil {
        log.Printf("Error message: %s", result.ErrorMessage)
        log.Printf("Gas spent: %s wei", result.TotalGasCost.String())
    }
    return err
}

// Success
log.Printf("✓ Withdrawal successful")
log.Printf("  NFT ID: %s", result.NFTTokenID.String())
log.Printf("  WAVAX withdrawn: %s wei", result.Amount0.String())
log.Printf("  USDC withdrawn: %s", result.Amount1.String())
log.Printf("  Gas cost: %s wei", result.TotalGasCost.String())

// Access transaction details
for _, tx := range result.Transactions {
    log.Printf("  - %s: %s (gas: %s wei)",
        tx.Operation, tx.TxHash.Hex(), tx.GasCost.String())
}
```

### Example 2: Error Handling

```go
nftTokenID := big.NewInt(99999)  // NFT we don't own

result, err := blackhole.Withdraw(nftTokenID)
if err != nil {
    // Check error type
    if strings.Contains(err.Error(), "not owned by wallet") {
        log.Printf("Ownership error: %v", err)
        // Handle ownership issue
    } else if strings.Contains(err.Error(), "validation failed") {
        log.Printf("Invalid input: %v", err)
        // Handle validation issue
    } else {
        log.Printf("Transaction error: %v", err)
        // Handle execution issue
    }

    // Result may contain partial information
    if result != nil && result.ErrorMessage != "" {
        log.Printf("Detailed error: %s", result.ErrorMessage)
    }

    return err
}
```

### Example 3: Gas Cost Analysis

```go
result, err := blackhole.Withdraw(nftTokenID)
if err != nil {
    return err
}

// Calculate gas cost in AVAX
gasCostWei := result.TotalGasCost
gasCostAVAX := new(big.Float).Quo(
    new(big.Float).SetInt(gasCostWei),
    big.NewFloat(1e18),
)

log.Printf("Gas cost: %s AVAX", gasCostAVAX.Text('f', 6))

// Calculate net proceeds (tokens - gas)
// Note: Gas is paid in AVAX, withdrawn in WAVAX/USDC
// This example assumes WAVAX = wrapped AVAX
netWAVAX := new(big.Int).Sub(result.Amount0, gasCostWei)
log.Printf("Net WAVAX (after gas): %s wei", netWAVAX.String())
```

## Integration Points

### Dependencies

**Required Clients**:
```go
// NonfungiblePositionManager client must be in ContractClient map
nftManagerClient := blackhole.ccm[nonfungiblePositionManager]
```

**Required Interfaces**:
```go
// ContractClient interface (existing)
type ContractClient interface {
    Call(from *common.Address, method string, args ...interface{}) ([]interface{}, error)
    Send(priority types.Priority, fixedGasLimit *big.Int, from *common.Address,
         privateKey *ecdsa.PrivateKey, method string, args ...interface{}) (common.Hash, error)
    Abi() *abi.ABI
}

// TxListener interface (existing)
type TxListener interface {
    WaitForTransaction(txHash common.Hash) (*types.TxReceipt, error)
}
```

### Contract Interactions

**NonfungiblePositionManager Methods Called**:

1. `ownerOf(uint256 tokenId) → address`
   - Purpose: Verify NFT ownership
   - Call type: View (read-only)

2. `positions(uint256 tokenId) → (uint88, address, address, address, address, int24, int24, uint128, ...)`
   - Purpose: Get position details (liquidity amount)
   - Call type: View (read-only)

3. `multicall(bytes[] data) → bytes[]`
   - Purpose: Execute decreaseLiquidity, collect, burn atomically
   - Call type: Transaction (state-changing)

**Multicall Operations** (in order):
1. `decreaseLiquidity(DecreaseLiquidityParams) → (uint256, uint256)`
2. `collect(CollectParams) → (uint256, uint256)`
3. `burn(uint256 tokenId)`

## Performance Characteristics

**Expected Duration**:
- Validation phase: < 1 second (2 view calls)
- Transaction submission: < 5 seconds
- Transaction confirmation: 5-120 seconds (depends on network)
- Total: Typically < 2 minutes under normal conditions

**Gas Consumption**:
- Estimated: 200,000 - 400,000 gas
- Actual varies based on position size and accumulated fees
- Higher if multicall includes first-time token transfers

**Network Calls**:
- 2 view calls (ownerOf, positions)
- 1 transaction (multicall)
- Transaction confirmation polling (TxListener)

## Security Considerations

### Slippage Protection

The function internally calculates minimum token amounts to protect against slippage:
```go
amount0Min = amount0Expected * (100 - slippagePct) / 100
amount1Min = amount1Expected * (100 - slippagePct) / 100
```

Default slippage tolerance: 5%

### Atomicity

The multicall ensures all-or-nothing execution:
- If decreaseLiquidity succeeds but collect fails → entire transaction reverts
- If collect succeeds but burn fails → entire transaction reverts
- No intermediate state where liquidity is removed but NFT still exists

### Re-entrancy

Not applicable - multicall is atomic and contract-side operation

### Front-Running

Position withdrawal is not susceptible to sandwich attacks as:
- Slippage protection limits price movement impact
- Withdrawal doesn't affect pool price significantly
- User is withdrawing their own position, not trading

## Testing Guidance

### Unit Test Cases

1. **Validation**: Test with invalid NFT IDs (0, negative, nil)
2. **Ownership**: Test with NFT owned by different address
3. **Position Query**: Test with non-existent NFT
4. **Slippage**: Verify amount0Min/amount1Min calculations

### Integration Test Cases

1. **Full Withdrawal**: Mint → Withdraw → Verify NFT burned
2. **Gas Tracking**: Verify TransactionRecord accuracy
3. **Error Recovery**: Test network failures during multicall
4. **Edge Cases**: Withdraw position with zero liquidity but accumulated fees

### Test NFT Setup

```go
// Setup: Create position via Mint
mintResult, _ := blackhole.Mint(
    big.NewInt(1e18),  // 1 WAVAX
    big.NewInt(1e6),   // 1 USDC
    6,                 // rangeWidth
    5,                 // slippagePct
)
nftTokenID := mintResult.NFTTokenID

// Test: Withdraw the position
withdrawResult, err := blackhole.Withdraw(nftTokenID)

// Verify: NFT no longer exists
_, err = nftManagerClient.Call(&blackhole.myAddr, "ownerOf", nftTokenID)
// Should fail with "ERC721: invalid token ID" or similar
```

## Version History

- **v1.0** (2025-12-21): Initial API specification
