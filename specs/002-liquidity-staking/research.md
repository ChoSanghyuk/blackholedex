# Research: Liquidity Position Staking in Gauge

**Feature**: 002-liquidity-staking
**Date**: 2025-12-16
**Purpose**: Resolve technical unknowns and establish implementation patterns for NFT staking

## Overview

This research document addresses the technical decisions required to implement the `Stake` method that transfers NFT liquidity positions from the user's wallet to GaugeV2 contracts for additional reward earning.

## Key Technical Decisions

### Decision 1: NFT Approval Pattern

**Decision**: Use ERC721 `approve(address to, uint256 tokenId)` for single NFT approval

**Rationale**:
- ERC721 standard provides two approval methods:
  - `approve(address to, uint256 tokenId)` - Approves single token
  - `setApprovalForAll(address operator, bool approved)` - Approves all tokens
- Single token approval is more secure (least privilege principle)
- Each liquidity position is staked individually, not in batches
- Reduces risk of unauthorized transfers of other NFT positions
- Consistent with ERC721 best practices for non-bulk operations

**Alternatives Considered**:
- `setApprovalForAll`: Rejected due to over-permissioning (gauge would have access to ALL user's positions)
- No approval check: Rejected due to gas waste (always sending approval tx even when not needed)

**Implementation Details**:
- Use `getApproved(uint256 tokenId)` to check existing approval
- Only call `approve(gaugeAddress, tokenId)` if current approval doesn't match gauge
- ABI signature verified in NonfungiblePositionManager contract

---

### Decision 2: GaugeV2 Deposit Function Signature

**Decision**: Use `deposit(uint256 amount)` where amount represents the NFT token ID

**Rationale**:
- User input confirms the method signature: `deposit(uint256 pglAmount)`
- Verified in GaugeV2.sol ABI: single uint256 parameter
- The parameter name "pglAmount" is historical (from Pangolin LP tokens) but accepts NFT token IDs
- This is the standard Blackhole DEX pattern for staking concentrated liquidity positions

**Alternatives Considered**:
- `deposit(uint256 tokenId, uint256 amount)`: Not present in GaugeV2 ABI
- `stake(uint256 tokenId)`: Not present in GaugeV2 ABI
- Custom staking interface: Rejected to maintain compatibility with existing Blackhole DEX contracts

**Implementation Details**:
- Method name: `deposit`
- Parameter: Single `*big.Int` representing NFT token ID
- No return value (void function)
- Requires prior NFT approval to gauge address

---

### Decision 3: Transaction Tracking Pattern

**Decision**: Reuse existing `TransactionRecord` structure and gas extraction patterns from `Mint` method

**Rationale**:
- Constitutional Principle 3 (Financial Transparency) requires consistent tracking
- Existing pattern proven effective in Mint implementation
- Code reuse reduces bugs and maintenance burden
- User expects identical tracking format for all operations

**Implementation Pattern**:
```go
// For each transaction:
1. Execute transaction and get tx hash
2. Wait for receipt via TxListener.WaitForTransaction()
3. Extract gas cost using util.ExtractGasCost(receipt)
4. Parse gas price and gas used from receipt
5. Create TransactionRecord with all fields
6. Append to transactions slice
7. Sum all gas costs for TotalGasCost field
```

**Alternatives Considered**:
- Simplified tracking (tx hash only): Rejected per constitutional requirement for complete cost visibility
- Async tracking: Rejected due to complexity and potential for incomplete data
- Database persistence: Out of scope (tracking in-memory for this feature)

---

### Decision 4: Error Handling Strategy

**Decision**: Fail-fast with comprehensive error context, never leaving NFT in intermediate state

**Rationale**:
- Constitutional Principle 5 (Fail-Safe Operation) mandates fund safety
- NFT positions represent significant value (liquidity deposits)
- Blockchain transactions are atomic - NFT either transfers or stays with owner
- Clear error messages enable user recovery

**Error Handling Rules**:
1. **Pre-validation failures**: Return error immediately, no transactions sent
2. **Approval failures**: Return error, NFT remains in user wallet
3. **Deposit failures**: Return error with deposit tx hash, NFT still in user wallet (approval succeeded but deposit reverted)
4. **Network failures**: Return timeout error with partial state logged

**Alternatives Considered**:
- Retry logic: Deferred to future enhancement (requires idempotency guarantees)
- Automatic rollback: Not applicable (blockchain transactions don't support traditional rollback)
- Silent failures: Rejected per Principle 3 (transparency)

---

### Decision 5: NFT Ownership Validation

**Decision**: Query `ownerOf(tokenId)` on NonfungiblePositionManager before any transactions

**Rationale**:
- Prevents gas waste on transactions that will fail
- Provides clear error message if NFT doesn't exist or belongs to different wallet
- ERC721 standard method: `ownerOf(uint256 tokenId) returns (address)`
- Read-only call (no gas cost for execution)

**Implementation**:
- Call `ownerOf(tokenId)` via ContractClient.Call()
- Compare returned address with `b.myAddr`
- Error if mismatch or call fails (NFT doesn't exist)
- Proceed to approval check only if ownership confirmed

**Alternatives Considered**:
- Skip validation: Rejected due to poor UX (confusing revert errors)
- Validate after approval: Rejected due to potential gas waste
- Balance check: Insufficient (doesn't prove ownership of specific token ID)

---

### Decision 6: Gauge Address Validation

**Decision**: Basic validation only - check non-zero address and contract code exists

**Rationale**:
- Full ABI validation would require complex interface detection
- Gauge addresses are configuration (typically hardcoded for WAVAX/USDC pool)
- Contract call will revert if gauge is invalid, providing clear error
- Constitutional scope limits to single pool/gauge anyway

**Implementation**:
- Verify gauge address != 0x0
- Optional: Use ContractClient existence check (if available)
- Rely on deposit transaction revert for invalid gauge detection

**Alternatives Considered**:
- ABI interface verification: Too complex for marginal benefit
- Whitelist validation: Better suited for configuration layer (future enhancement)
- No validation: Rejected due to confusing error messages

---

## Best Practices Applied

### Go-Ethereum Integration
- Use `common.Address` for all addresses
- Use `*big.Int` for uint256 values
- Use `common.Hash` for transaction hashes
- Follow existing ContractClient patterns from blackhole.go

### Error Wrapping
- Always wrap errors with `fmt.Errorf("context: %w", err)` for full error chain
- Include operation name in error context
- Preserve original error for debugging

### Gas Optimization (Principle 4)
- Check approval before sending approval transaction
- Use automatic gas estimation (nil gas limit)
- No unnecessary contract reads

### Financial Transparency (Principle 3)
- Track every transaction with full gas details
- Include timestamps for audit trail
- Sum total costs across all operations
- Reuse proven tracking patterns

---

## Contract Interface Summary

### NonfungiblePositionManager (ERC721)
- **Address**: 0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146
- **Methods Used**:
  - `getApproved(uint256 tokenId) returns (address)` - Check current approval
  - `approve(address to, uint256 tokenId)` - Approve single NFT
  - `ownerOf(uint256 tokenId) returns (address)` - Verify ownership

### GaugeV2
- **Address**: Configuration-dependent (per pool)
- **Methods Used**:
  - `deposit(uint256 amount)` - Stake NFT position (amount = token ID)

---

## Implementation Checklist

- [x] NFT approval pattern decided (single token approval)
- [x] Deposit function signature confirmed (deposit(uint256))
- [x] Transaction tracking pattern established (reuse Mint pattern)
- [x] Error handling strategy defined (fail-fast, preserve ownership)
- [x] Ownership validation approach determined (ownerOf check)
- [x] Gauge validation scope defined (basic validation only)
- [x] All contract ABIs verified in artifacts

---

## References

- NonfungiblePositionManager ABI: `blackholedex-contracts/artifacts/@cryptoalgebra/integral-periphery/contracts/interfaces/INonfungiblePositionManager.sol/INonfungiblePositionManager.json`
- GaugeV2 ABI: `blackholedex-contracts/artifacts/contracts/GaugeV2.sol/GaugeV2.json`
- Existing Mint implementation: `blackhole.go` lines 221-533
- Existing approval pattern: `blackhole.go` lines 183-219 (ensureApproval)
- Project constitution: `.specify/memory/constitution.md`
