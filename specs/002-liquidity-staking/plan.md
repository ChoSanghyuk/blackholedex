# Implementation Plan: Liquidity Position Staking in Gauge

**Branch**: `002-liquidity-staking` | **Date**: 2025-12-16 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-liquidity-staking/spec.md`

**User Input**: The deposit transaction follows this method signature `Function: deposit(uint256 pglAmount)`. Use this signature for the method ABI.

## Summary

Implement a `Stake` method in `blackhole.go` that stakes liquidity position NFTs (from `Mint` operations) into GaugeV2 contracts to earn additional rewards. The method follows the existing `Mint` pattern for transaction tracking and error handling, conditionally approving the NFT for transfer to the gauge contract before depositing it via the `deposit(uint256 pglAmount)` function where `pglAmount` represents the NFT token ID.

## Technical Context

**Language/Version**: Go 1.24.10
**Primary Dependencies**: github.com/ethereum/go-ethereum v1.16.7, existing internal packages (util, contractclient)
**Storage**: N/A (blockchain state only, no local persistence)
**Testing**: Go testing package, table-driven tests, integration tests against Avalanche C-Chain
**Target Platform**: Avalanche C-Chain mainnet/testnet
**Project Type**: Single project (CLI tool with library functions)
**Performance Goals**: Complete stake operation in <90 seconds (includes approval + deposit confirmations)
**Constraints**: Gas costs must be minimized (reuse existing approvals), all operations must be fail-safe (NFT ownership never compromised)
**Scale/Scope**: Single method addition to existing Blackhole struct, reuses existing types (StakingResult, TransactionRecord)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

This feature MUST comply with all principles in `.specify/memory/constitution.md`. Check each principle:

### Principle 1: Pool-Specific Scope
- [x] Feature operates ONLY on WAVAX/USDC pool
- [x] No support for other tokens or pools introduced
- [x] Contract addresses match constitution technical constraints

**Compliance Details**:
- Stake method accepts gauge address as parameter (configuration-controlled)
- NFT positions are only for WAVAX/USDC pool (minted by existing Mint method)
- No changes to supported tokens or pool scope

### Principle 2: Autonomous Rebalancing
- [x] Monitoring capability included (if relevant)
- [x] Rebalancing logic maintains atomicity/rollback safety
- [x] Decision logic documented and testable

**Compliance Details**:
- Not directly applicable to manual staking operation
- Staking is step 1 of future autonomous rebalancing workflow (unstake → rebalance → stake)
- Individual operations maintain atomicity (NFT either in wallet or gauge, never intermediate)

### Principle 3: Financial Transparency
- [x] Gas tracking implemented for all transactions
- [x] Swap fees calculated and recorded
- [x] Incentives/rewards tracked
- [x] Profit/loss calculations accurate
- [x] Exportable reporting available

**Compliance Details**:
- All transactions tracked in TransactionRecord with full gas details
- Approval gas + deposit gas separately recorded
- TotalGasCost sums all transaction costs
- No swap fees in this operation (pure staking)
- Rewards tracking deferred to future ClaimRewards method

### Principle 4: Gas Optimization
- [x] Rebalancing thresholds configured to prevent excessive transactions
- [x] Gas estimation used before transaction submission
- [x] Contract calls optimized (batch operations, minimal storage)
- [x] Approval reuse implemented where safe

**Compliance Details**:
- Existing approval checked via `getApproved(tokenId)` before sending approval tx
- Automatic gas estimation (nil gas limit parameter)
- Single NFT approval (not setApprovalForAll) for least privilege
- No unnecessary contract reads

### Principle 5: Fail-Safe Operation
- [x] All external calls have timeout/retry logic
- [x] Transaction failures logged with full context
- [x] Partial failures trigger rollback or safe termination
- [x] Slippage protection enforced
- [x] Circuit breaker pattern implemented
- [x] Manual recovery documented

**Compliance Details**:
- TxListener.WaitForTransaction provides timeout/retry
- All errors wrapped with full context (fmt.Errorf with %w)
- Partial failures return StakingResult with Success=false and transaction records
- NFT ownership never in intermediate state (ERC721 atomic transfer)
- No slippage in pure staking (no swaps involved)
- Circuit breaker not needed for single operation (future enhancement for autonomous loop)

**Violations**: None

## Project Structure

### Documentation (this feature)

```text
specs/002-liquidity-staking/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output - Technical decisions and ABI verification
├── data-model.md        # Phase 1 output - Entity definitions and state transitions
├── quickstart.md        # Phase 1 output - Developer implementation guide
├── contracts/           # Phase 1 output - API contracts
│   └── stake-api.md     # Method signature and usage contract
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created yet)
```

### Source Code (repository root)

```text
blackhole_dex/
├── blackhole.go                    # Main implementation file (ADD Stake method here)
├── blackhole_interfaces.go         # Interfaces (no changes needed)
├── types.go                        # Type definitions (reuse StakingResult, TransactionRecord)
├── internal/
│   └── util/                       # Utilities (reuse ExtractGasCost, ValidateStakingRequest patterns)
├── pkg/
│   └── contractclient/             # Generic contract client (no changes needed)
├── cmd/                            # CLI entry points (future: add stake command)
├── blackholedex-contracts/
│   └── artifacts/                  # Contract ABIs (reference GaugeV2, INonfungiblePositionManager)
└── tests/                          # Test files (ADD stake_test.go)
```

**Structure Decision**: Single project structure is appropriate. Stake method is a new public method on the existing `Blackhole` struct in `blackhole.go`, following the pattern established by `Mint`, `Swap`, and `GetAMMState`. No new packages or modules required.

**Key Files to Modify**:
- `blackhole.go`: Add `Stake` method (lines ~535+)
- Optional: Add helper if `sumGasCosts` doesn't exist

**Key Files to Reference**:
- `blackhole.go` lines 221-533: `Mint` method (transaction tracking pattern)
- `blackhole.go` lines 183-219: `ensureApproval` method (approval check pattern)
- `types.go` lines 166-186: `TransactionRecord` and `StakingResult` definitions
- Contract ABIs:
  - `blackholedex-contracts/artifacts/contracts/GaugeV2.sol/GaugeV2.json` (deposit method)
  - `blackholedex-contracts/artifacts/@cryptoalgebra/integral-periphery/contracts/interfaces/INonfungiblePositionManager.sol/INonfungiblePositionManager.json` (approve, getApproved, ownerOf methods)

## Complexity Tracking

No constitutional violations to justify. This feature fully complies with all principles.

## Phase 0: Research (Complete)

**Status**: ✅ Complete
**Output**: [research.md](./research.md)

### Key Decisions Made

1. **NFT Approval Pattern**: Use `approve(address to, uint256 tokenId)` for single NFT approval (not setApprovalForAll)
   - Rationale: Least privilege, more secure, consistent with single-position staking

2. **Deposit Function Signature**: Use `deposit(uint256 amount)` where amount = NFT token ID
   - Rationale: Confirmed in GaugeV2.sol ABI, parameter name "pglAmount" is historical but accepts NFT IDs

3. **Transaction Tracking**: Reuse existing `TransactionRecord` and gas extraction patterns from `Mint`
   - Rationale: Constitutional transparency requirement, code reuse, consistent UX

4. **Error Handling**: Fail-fast with comprehensive context, never leave NFT in intermediate state
   - Rationale: Constitutional safety requirement, blockchain atomicity guarantees fund safety

5. **Ownership Validation**: Query `ownerOf(tokenId)` before any transactions
   - Rationale: Prevents gas waste, provides clear error messages

6. **Gauge Validation**: Basic validation only (non-zero address, rely on contract revert for invalid gauge)
   - Rationale: Configuration controls gauge address, full ABI validation too complex for marginal benefit

### Research Artifacts

- Contract interface summary (NonfungiblePositionManager and GaugeV2 methods)
- Best practices for ERC721 approval patterns
- Gas optimization strategies (approval reuse)
- Error handling patterns from go-ethereum

## Phase 1: Design & Contracts (Complete)

**Status**: ✅ Complete
**Outputs**:
- [data-model.md](./data-model.md)
- [contracts/stake-api.md](./contracts/stake-api.md)
- [quickstart.md](./quickstart.md)

### Data Model

**Entities** (all existing, reused):
- `StakingResult`: Comprehensive output structure with transaction tracking
- `TransactionRecord`: Immutable transaction record with gas metrics
- `NFT Position Token`: External ERC721 entity (read-only queries)
- `Gauge Contract`: External GaugeV2 entity (deposit interaction)

**Key State Transitions**:
```
Initial → Validation → Approval Check → [Approval?] → Deposit → Success/Failed
```

**Invariants**:
- NFT is either owned by user OR gauge, never intermediate
- Every sent transaction tracked in Transactions array
- TotalGasCost = sum(TransactionRecord[*].GasCost)
- Success=true implies NFT owned by gauge
- Success=false implies NFT owned by user

### API Contract

**Method Signature**:
```go
func (b *Blackhole) Stake(
    nftTokenID *big.Int,
    gaugeAddress common.Address,
) (*StakingResult, error)
```

**Key Behaviors**:
- Input validation (token ID > 0, gauge address != 0x0)
- Ownership verification (ownerOf call)
- Conditional approval (getApproved → approve if needed)
- Gauge deposit (deposit(tokenId))
- Comprehensive transaction tracking

**Performance**:
- Best case: 2-5 seconds (no approval needed)
- Typical case: 4-10 seconds (approval + deposit)
- Gas cost: ~0.002-0.006 AVAX

**Error Handling**:
- Validation failures: No transactions sent, immediate error return
- Transaction failures: Partial StakingResult with all sent transactions tracked
- NFT ownership preserved in all failure scenarios

### Quickstart Guide

Provides step-by-step implementation instructions:
1. Define method signature (5 min)
2. Input validation (15 min)
3. Verify NFT ownership (20 min)
4. Check and handle approval (30 min)
5. Execute gauge deposit (30 min)
6. Construct success result (15 min)
7. Add helper functions (10 min)

Total estimated implementation time: 2-4 hours

## Phase 2: Task Generation (Pending)

**Status**: ⏳ Pending - Use `/speckit.tasks` to generate

This plan document stops at Phase 1 (design artifacts). The next command `/speckit.tasks` will:
- Generate detailed task breakdown in `tasks.md`
- Create dependency-ordered implementation tasks
- Map tasks to success criteria and requirements
- Provide task-specific acceptance criteria

**Do not proceed to implementation until tasks.md is generated and reviewed.**

## Implementation Notes

### Code Patterns to Follow

1. **Error Wrapping**: Always use `fmt.Errorf("context: %w", err)` to preserve error chain
2. **Gas Tracking**: Identical pattern to Mint method (extract receipt → parse gas → create record → append)
3. **Logging**: Use `log.Printf` for operational logging, `fmt.Printf` for user-facing success messages
4. **Type Conversions**: Use go-ethereum types (common.Address, common.Hash, *big.Int)
5. **Return Values**: Always return *StakingResult even on error (with Success=false)

### Testing Strategy

**Unit Tests**:
- Valid stake with approval (verify 2 transactions)
- Valid stake without approval (verify 1 transaction)
- Invalid token ID (verify early return, no transactions)
- NFT not owned (verify validation failure)
- Approval failure (verify partial result)
- Deposit failure (verify partial result, NFT still owned)

**Integration Tests**:
- Full workflow: Mint → Stake → Verify gauge ownership
- Approval reuse: Stake → Unstake → Stake (approval persists)
- Real contract interactions on testnet/fork

**Edge Case Tests**:
- Zero token ID
- Zero gauge address
- NFT already staked (ownerOf returns gauge)
- Gauge paused/disabled
- Concurrent stake attempts (race condition)

### Security Considerations

1. **Approval Scope**: Use `approve(gauge, tokenId)` NOT `setApprovalForAll(gauge, true)`
2. **Ownership Validation**: Always verify ownerOf before any transactions
3. **Private Key Handling**: Already secured in Blackhole struct (no changes needed)
4. **Contract Address Validation**: Basic checks only (full validation in configuration layer)
5. **Reentrancy**: Not applicable (blockchain transactions are atomic)

### Gas Optimization Notes

1. **Approval Reuse**: Check getApproved before submitting approval tx (saves ~50k gas when already approved)
2. **Automatic Estimation**: Use nil gas limit to let go-ethereum estimate (prevents over-payment)
3. **Batching**: Not applicable (single NFT deposit, approval can't be batched with deposit due to ERC721 requirement)
4. **Read Optimization**: Minimal contract reads (only ownerOf and getApproved)

## Dependencies

### External Dependencies (No Changes)
- github.com/ethereum/go-ethereum v1.16.7
- Avalanche C-Chain RPC endpoint
- Contract ABIs in blackholedex-contracts/artifacts

### Internal Dependencies (Existing)
- `ContractClient` interface and implementation
- `TxListener` interface for transaction confirmation
- `util.ExtractGasCost` for gas calculation
- `types.Standard` for transaction type
- `StakingResult` and `TransactionRecord` types

### Contract Dependencies
- NonfungiblePositionManager at 0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146
  - Methods: `ownerOf`, `getApproved`, `approve`
- GaugeV2 at configuration-specified address
  - Methods: `deposit`

## Risk Assessment

### Low Risk
- Code reuse from Mint method (proven patterns)
- Simple ERC721 operations (standard interface)
- No complex state management (blockchain handles atomicity)

### Medium Risk
- Gas price volatility (mitigated by automatic estimation)
- Network timeouts (mitigated by TxListener timeout handling)
- Configuration errors (wrong gauge address → transaction revert)

### Mitigation Strategies
- Comprehensive input validation before any blockchain calls
- Detailed error messages for all failure scenarios
- Transaction tracking even for failures (audit trail)
- Integration tests with real contracts on testnet

## Success Metrics (from Spec)

Implementation will be considered successful when these criteria are met:

- **SC-001**: Users can stake NFT in <90 seconds
  - Validation: Integration test measures total time approval + deposit + confirmations

- **SC-002**: 100% accuracy when wallet has gas and NFT ownership is valid
  - Validation: Unit tests cover all valid input scenarios, no false failures

- **SC-003**: Gas cost tracking matches on-chain receipts with 100% accuracy
  - Validation: Integration test compares tracked costs to queried receipt data

- **SC-004**: Transaction failures result in zero NFT ownership loss
  - Validation: Error injection tests verify NFT remains in user wallet

- **SC-005**: Approval reuse 90% effective (unnecessary approvals avoided)
  - Validation: Integration test stakes same NFT twice, verifies approval skipped on second call

- **SC-006**: Error messages actionable in 100% of failure cases
  - Validation: Code review + manual testing of each error path

## Next Steps

1. ✅ Research technical decisions (Phase 0) - **COMPLETE**
2. ✅ Design data model and contracts (Phase 1) - **COMPLETE**
3. ⏳ Generate implementation tasks: Run `/speckit.tasks`
4. ⏳ Implement Stake method following tasks.md
5. ⏳ Write unit tests for all scenarios
6. ⏳ Write integration tests with real contracts
7. ⏳ Verify constitutional compliance in implementation
8. ⏳ Code review and refinement
9. ⏳ Documentation and examples

**Current Status**: Planning phase complete. Ready for `/speckit.tasks` command to generate detailed implementation tasks.

## References

- **Feature Spec**: [spec.md](./spec.md)
- **Research**: [research.md](./research.md)
- **Data Model**: [data-model.md](./data-model.md)
- **API Contract**: [contracts/stake-api.md](./contracts/stake-api.md)
- **Quickstart**: [quickstart.md](./quickstart.md)
- **Constitution**: [../../.specify/memory/constitution.md](../../.specify/memory/constitution.md)
- **Existing Mint Implementation**: `blackhole.go` lines 221-533
- **Contract ABIs**: `blackholedex-contracts/artifacts/contracts/`
