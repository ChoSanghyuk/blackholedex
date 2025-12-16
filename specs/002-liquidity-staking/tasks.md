# Tasks: Liquidity Position Staking in Gauge

**Input**: Design documents from `/specs/002-liquidity-staking/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: No test tasks included - tests not explicitly requested in specification

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Single project structure - all paths relative to repository root:
- Main implementation: `blackhole.go`
- Type definitions: `types.go` (no changes needed - reusing existing)
- Utilities: `internal/util/` (no changes needed - reusing existing)
- Contract ABIs: `blackholedex-contracts/artifacts/` (reference only)

---

## Phase 1: Setup (No New Infrastructure Needed)

**Purpose**: Verify prerequisites and references

**Status**: This feature adds a single method to existing code - no new project setup required

- [x] T001 Verify blackhole.go exists at repository root and contains Blackhole struct
- [x] T002 Verify types.go contains StakingResult and TransactionRecord definitions (lines 166-186)
- [x] T003 [P] Verify GaugeV2.sol ABI exists at blackholedex-contracts/artifacts/contracts/GaugeV2.sol/GaugeV2.json
- [x] T004 [P] Verify INonfungiblePositionManager ABI exists at blackholedex-contracts/artifacts/@cryptoalgebra/integral-periphery/contracts/interfaces/INonfungiblePositionManager.sol/INonfungiblePositionManager.json
- [x] T005 Verify util.ExtractGasCost function exists in internal/util/ (referenced by Mint method)

**Checkpoint**: All referenced files and functions exist - ready for implementation

---

## Phase 2: Foundational (Core Method Structure)

**Purpose**: Create the Stake method skeleton with validation framework

**⚠️ CRITICAL**: This phase establishes the method structure that all user stories build upon

- [x] T006 Add Stake method signature to blackhole.go after Mint method (around line 535)
- [x] T007 Implement input validation for nftTokenID parameter (nil check, positive value check)
- [x] T008 Implement input validation for gaugeAddress parameter (zero address check)
- [x] T009 Initialize transactions slice for tracking TransactionRecord entries
- [x] T010 Add error return structure with StakingResult containing Success=false and ErrorMessage

**Checkpoint**: Method exists with signature `func (b *Blackhole) Stake(nftTokenID *big.Int, gaugeAddress common.Address) (*StakingResult, error)` and validates inputs

---

## Phase 3: User Story 1 - Stake Minted Liquidity Position (Priority: P1) 🎯 MVP

**Goal**: Enable users to stake NFT positions in gauge contracts with approval handling, deposit execution, and complete transaction tracking

**Independent Test**: Mint a position using existing Mint method, call Stake with returned NFT token ID and gauge address, verify:
1. NFT ownership transferred from user wallet to gauge contract
2. StakingResult.Success = true
3. StakingResult.Transactions contains 1-2 records (ApproveNFT and/or DepositNFT)
4. StakingResult.TotalGasCost = sum of all transaction gas costs
5. Gauge contract shows user's staked position balance increased

### Implementation for User Story 1

**Subtask: NFT Ownership Verification (FR-002)**

- [x] T011 [US1] Get NonfungiblePositionManager ContractClient in blackhole.go Stake method using b.Client(nonfungiblePositionManager)
- [x] T012 [US1] Call ownerOf(nftTokenID) on NonfungiblePositionManager contract to query current owner
- [x] T013 [US1] Compare returned owner address with b.myAddr and return validation error if mismatch
- [x] T014 [US1] Add error handling for ownerOf call failure (NFT doesn't exist scenario)

**Subtask: NFT Approval Check and Execution (FR-003, FR-004, FR-005)**

- [x] T015 [US1] Call getApproved(nftTokenID) on NonfungiblePositionManager to check existing approval
- [x] T016 [US1] Compare current approval address with gaugeAddress parameter
- [x] T017 [US1] If approval needed: call approve(gaugeAddress, nftTokenID) using ContractClient.Send with types.Standard and nil gas limit
- [x] T018 [US1] If approval needed: wait for approval transaction confirmation using b.tl.WaitForTransaction(approveTxHash)
- [x] T019 [US1] If approval needed: extract gas cost from approval receipt using util.ExtractGasCost
- [x] T020 [US1] If approval needed: parse receipt.EffectiveGasPrice and receipt.GasUsed to *big.Int
- [x] T021 [US1] If approval needed: create TransactionRecord with Operation="ApproveNFT" and append to transactions slice
- [x] T022 [US1] If approval exists: log "NFT already approved for gauge, skipping approval" and proceed to deposit
- [x] T023 [US1] Add error handling for approval transaction failure and return StakingResult with partial transactions

**Subtask: Gauge Deposit Execution (FR-006, FR-007)**

- [x] T024 [US1] Get gauge ContractClient in blackhole.go Stake method using b.Client(gaugeAddress.Hex())
- [x] T025 [US1] Call deposit(nftTokenID) on gauge contract using ContractClient.Send with types.Standard, nil gas limit, and nftTokenID as parameter
- [x] T026 [US1] Wait for deposit transaction confirmation using b.tl.WaitForTransaction(depositTxHash)
- [x] T027 [US1] Extract gas cost from deposit receipt using util.ExtractGasCost
- [x] T028 [US1] Parse receipt.EffectiveGasPrice and receipt.GasUsed to *big.Int for deposit transaction
- [x] T029 [US1] Create TransactionRecord with Operation="DepositNFT" and append to transactions slice
- [x] T030 [US1] Add error handling for deposit transaction failure and return StakingResult with all sent transactions

**Subtask: Result Construction and Gas Tracking (FR-008, FR-009, FR-010)**

- [x] T031 [US1] Calculate totalGasCost by summing all TransactionRecord.GasCost values in transactions slice
- [x] T032 [US1] Create StakingResult with NFTTokenID set to input nftTokenID parameter
- [x] T033 [US1] Set StakingResult.Transactions to transactions slice (1-2 records)
- [x] T034 [US1] Set StakingResult.TotalGasCost to calculated totalGasCost value
- [x] T035 [US1] Set StakingResult.Success to true on successful deposit confirmation
- [x] T036 [US1] Set StakingResult.ActualAmount0, ActualAmount1, FinalTickLower, FinalTickUpper to zero values (not used by Stake)
- [x] T037 [US1] Leave StakingResult.ErrorMessage empty string on success

**Subtask: Logging and User Feedback (FR-014)**

- [x] T038 [US1] Add log.Printf for approval skipped scenario in blackhole.go Stake method
- [x] T039 [US1] Add log.Printf for approval transaction submitted scenario
- [x] T040 [US1] Add log.Printf for deposit transaction submitted scenario
- [x] T041 [US1] Add fmt.Printf success message showing NFT token ID, gauge address, and total gas cost
- [x] T042 [US1] Add fmt.Printf for each transaction in transactions slice with operation type and gas cost
- [x] T043 [US1] Follow existing Mint method logging pattern for consistency

**Checkpoint**: User Story 1 complete - Stake method functional with approval optimization, deposit execution, and comprehensive transaction tracking

---

## Phase 4: User Story 2 - Handle Staking Failures Safely (Priority: P2)

**Goal**: Ensure all failure scenarios preserve NFT ownership in user wallet and provide actionable error messages

**Independent Test**: Simulate failure scenarios (invalid token ID, network timeout, contract revert) and verify:
1. NFT remains in user wallet after each failure
2. StakingResult.Success = false
3. StakingResult.ErrorMessage contains descriptive error
4. StakingResult.Transactions contains all successfully sent transactions (may be empty, may have approval only)
5. No intermediate state where NFT ownership is unclear

### Implementation for User Story 2

**Subtask: Pre-Transaction Validation Errors (Edge Cases)**

- [x] T044 [US2] Add validation error for nftTokenID <= 0 with ErrorMessage="validation failed: invalid token ID"
- [x] T045 [US2] Add validation error for zero gauge address with ErrorMessage="validation failed: invalid gauge address"
- [x] T046 [US2] Return StakingResult with Success=false and empty Transactions array for validation failures
- [x] T047 [US2] Ensure validation failures do not execute any blockchain operations

**Subtask: NFT Ownership Verification Errors (FR-002)**

- [x] T048 [US2] Handle ownerOf call failure with ErrorMessage="failed to verify NFT ownership: [error details]"
- [x] T049 [US2] Handle owner mismatch with ErrorMessage="NFT not owned by wallet: owned by [owner address]"
- [x] T050 [US2] Return StakingResult with NFTTokenID set and empty Transactions for ownership failures
- [x] T051 [US2] Wrap errors with fmt.Errorf("context: %w", err) for full error chain

**Subtask: Approval Transaction Failures (FR-005, FR-011)**

- [x] T052 [US2] Handle approval Send failure with ErrorMessage="failed to approve NFT: [error details]"
- [x] T053 [US2] Handle approval WaitForTransaction timeout with ErrorMessage="NFT approval transaction failed: [error details]"
- [x] T054 [US2] Handle approval gas extraction failure with ErrorMessage="failed to extract approval gas cost: [error details]"
- [x] T055 [US2] Return StakingResult with partial transactions (empty or approval record) on approval failures
- [x] T056 [US2] Ensure NFT ownership remains with user on approval failure

**Subtask: Deposit Transaction Failures (FR-007, FR-011)**

- [x] T057 [US2] Handle deposit Send failure with ErrorMessage="failed to submit deposit transaction: [error details]"
- [x] T058 [US2] Handle deposit WaitForTransaction timeout with ErrorMessage="deposit transaction failed: [error details]"
- [x] T059 [US2] Handle deposit gas extraction failure with ErrorMessage="failed to extract deposit gas cost: [error details]"
- [x] T060 [US2] Return StakingResult with transactions array containing approval record (if sent) on deposit failures
- [x] T061 [US2] Calculate and set TotalGasCost from partial transactions on deposit failure
- [x] T062 [US2] Ensure NFT ownership remains with user even if approval succeeded but deposit failed

**Subtask: ContractClient Retrieval Failures**

- [x] T063 [US2] Handle b.Client(nonfungiblePositionManager) failure with ErrorMessage="failed to get NFT manager client: [error details]"
- [x] T064 [US2] Handle b.Client(gaugeAddress) failure with ErrorMessage="failed to get gauge client: [error details]"
- [x] T065 [US2] Return appropriate StakingResult with transactions recorded up to failure point

**Checkpoint**: User Story 2 complete - All failure paths tested, NFT ownership preserved in all scenarios, error messages actionable

---

## Phase 5: User Story 3 - Track Financial Impact Transparently (Priority: P3)

**Goal**: Provide complete gas cost visibility with accurate tracking matching on-chain receipts

**Independent Test**: Execute stake operation and verify:
1. Each TransactionRecord has non-zero GasUsed, GasPrice, and GasCost fields
2. TransactionRecord.GasCost = GasUsed * GasPrice for each record
3. StakingResult.TotalGasCost = sum of all TransactionRecord.GasCost values
4. Timestamp field populated with transaction confirmation time
5. Operation field correctly set to "ApproveNFT" or "DepositNFT"
6. Gas costs match values queryable from on-chain transaction receipts

### Implementation for User Story 3

**Subtask: Approval Transaction Tracking (FR-008, FR-015)**

- [x] T066 [US3] Verify approval TransactionRecord.TxHash equals approveTxHash from Send call
- [x] T067 [US3] Verify approval TransactionRecord.GasUsed parsed correctly from receipt.GasUsed string to uint64
- [x] T068 [US3] Verify approval TransactionRecord.GasPrice parsed correctly from receipt.EffectiveGasPrice string to *big.Int
- [x] T069 [US3] Verify approval TransactionRecord.GasCost equals util.ExtractGasCost(receipt) result
- [x] T070 [US3] Verify approval TransactionRecord.Timestamp set to time.Now() at confirmation
- [x] T071 [US3] Verify approval TransactionRecord.Operation set to string literal "ApproveNFT"

**Subtask: Deposit Transaction Tracking (FR-008, FR-015)**

- [x] T072 [US3] Verify deposit TransactionRecord.TxHash equals depositTxHash from Send call
- [x] T073 [US3] Verify deposit TransactionRecord.GasUsed parsed correctly from receipt.GasUsed string to uint64
- [x] T074 [US3] Verify deposit TransactionRecord.GasPrice parsed correctly from receipt.EffectiveGasPrice string to *big.Int
- [x] T075 [US3] Verify deposit TransactionRecord.GasCost equals util.ExtractGasCost(receipt) result
- [x] T076 [US3] Verify deposit TransactionRecord.Timestamp set to time.Now() at confirmation
- [x] T077 [US3] Verify deposit TransactionRecord.Operation set to string literal "DepositNFT"

**Subtask: Total Gas Cost Calculation (FR-009)**

- [x] T078 [US3] Initialize totalGasCost as big.NewInt(0) in blackhole.go Stake method
- [x] T079 [US3] Iterate over transactions slice and sum each TransactionRecord.GasCost
- [x] T080 [US3] Use big.Int.Add(totalGasCost, tx.GasCost) for each transaction in loop
- [x] T081 [US3] Set StakingResult.TotalGasCost to final totalGasCost value
- [x] T082 [US3] Verify TotalGasCost calculation matches pattern from Mint method (lines 505-508)

**Subtask: Transaction Array Management**

- [x] T083 [US3] Verify transactions slice contains exactly 2 records when approval was needed (ApproveNFT + DepositNFT)
- [x] T084 [US3] Verify transactions slice contains exactly 1 record when approval existed (DepositNFT only)
- [x] T085 [US3] Verify transactions array ordering: ApproveNFT always before DepositNFT when both present
- [x] T086 [US3] Verify StakingResult.Transactions is a copy of transactions slice (not modified by caller)

**Checkpoint**: User Story 3 complete - Financial transparency fully implemented, gas tracking matches Mint method patterns, all costs accurately calculated

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Code quality, documentation, and constitutional compliance verification

- [x] T087 [P] Verify all error messages are descriptive and actionable per SC-006
- [x] T088 [P] Verify code follows existing patterns from Mint method (error wrapping, logging, gas tracking)
- [x] T089 [P] Add inline comments for complex logic (approval check, gas calculation)
- [x] T090 [P] Verify method signature matches API contract in specs/002-liquidity-staking/contracts/stake-api.md
- [x] T091 Run go fmt on blackhole.go to ensure consistent formatting
- [x] T092 Run go vet on blackhole.go to check for common Go mistakes
- [x] T093 Verify constitutional compliance: Principle 1 (WAVAX/USDC scope maintained)
- [x] T094 Verify constitutional compliance: Principle 3 (all gas costs tracked)
- [x] T095 Verify constitutional compliance: Principle 4 (approval reuse implemented)
- [x] T096 Verify constitutional compliance: Principle 5 (NFT ownership never in intermediate state)
- [x] T097 Review quickstart.md implementation guide and verify all steps followed
- [x] T098 Verify method works with existing Blackhole struct initialization (no breaking changes)

**Checkpoint**: Implementation complete, code quality verified, constitutional compliance confirmed

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - verification tasks only
- **Foundational (Phase 2)**: Depends on Setup verification - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2) - Core functionality
- **User Story 2 (Phase 4)**: Depends on User Story 1 implementation - Tests error paths of US1 code
- **User Story 3 (Phase 5)**: Depends on User Story 1 implementation - Verifies tracking from US1 code
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Independent - can implement after Foundational
- **User Story 2 (P2)**: Depends on User Story 1 - adds error handling to US1 code paths
- **User Story 3 (P3)**: Depends on User Story 1 - verifies transaction tracking from US1

**Note**: US2 and US3 are technically modifications/enhancements to US1 code, not independent stories. They must execute sequentially: US1 → US2 → US3

### Within Each User Story

**User Story 1**:
1. T011-T014: NFT ownership verification (sequential)
2. T015-T023: Approval logic (sequential within, but after ownership)
3. T024-T030: Deposit logic (sequential, after approval)
4. T031-T037: Result construction (sequential, after deposit)
5. T038-T043: Logging (parallel, can add throughout)

**User Story 2**:
- All error handling tasks can be implemented in parallel (different error conditions)
- T044-T065 modify different error paths in the same method

**User Story 3**:
- All verification tasks (T066-T086) can be implemented in parallel (different TransactionRecord fields)
- Gas calculation tasks (T078-T082) must be sequential

### Parallel Opportunities

**Phase 1 (Setup)**:
- T001, T002, T003, T004, T005 can all verify in parallel (different files)

**Phase 2 (Foundational)**:
- T006 must complete first
- T007-T010 can run in parallel after T006 (different validation checks)

**User Story 1**:
- Within each subtask, steps are sequential
- Logging tasks (T038-T043) can be done throughout implementation

**User Story 2**:
- T044-T047: Validation errors (parallel)
- T048-T051: Ownership errors (parallel)
- T052-T056: Approval errors (parallel)
- T057-T062: Deposit errors (parallel)
- T063-T065: Client errors (parallel)

**User Story 3**:
- T066-T071: Approval tracking verification (parallel)
- T072-T077: Deposit tracking verification (parallel)
- T078-T082: Gas calculation (sequential)
- T083-T086: Array verification (parallel)

**Phase 6 (Polish)**:
- T087-T090: Documentation tasks (parallel)
- T091-T092: Code quality tools (sequential)
- T093-T096: Constitutional checks (parallel)
- T097-T098: Final verification (sequential)

---

## Parallel Example: User Story 2 Error Handling

```bash
# Launch all error handling subtasks together:
Task: "Add validation error for nftTokenID <= 0" (T044)
Task: "Add validation error for zero gauge address" (T045)
Task: "Handle ownerOf call failure" (T048)
Task: "Handle approval Send failure" (T052)
Task: "Handle deposit Send failure" (T057)
Task: "Handle b.Client(nonfungiblePositionManager) failure" (T063)

# All operate on different error conditions in same method
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup verification (~15 minutes)
2. Complete Phase 2: Foundational method skeleton (~30 minutes)
3. Complete Phase 3: User Story 1 full implementation (~2-3 hours)
4. **STOP and VALIDATE**: Test stake with real NFT on testnet
5. Verify: Approval optimization works, deposit succeeds, gas tracked correctly

### Incremental Delivery

1. Setup + Foundational → Method exists and validates inputs
2. Add User Story 1 → Test with approval needed case → Test with approval exists case → **MVP Ready!**
3. Add User Story 2 → Test all error scenarios → Verify NFT safety → **Error handling complete**
4. Add User Story 3 → Verify gas accuracy against on-chain → **Full transparency achieved**
5. Polish → Code review, compliance check → **Production ready**

### Single Developer Timeline

- **Phase 1-2**: 45 minutes (setup + skeleton)
- **Phase 3 (US1)**: 2-3 hours (core implementation)
- **Phase 4 (US2)**: 1-2 hours (error handling)
- **Phase 5 (US3)**: 1 hour (gas tracking verification)
- **Phase 6**: 30 minutes (polish)

**Total Estimate**: 5-7 hours for complete feature

### Parallel Team Strategy

With 2 developers after US1 is complete:
- Developer A: User Story 2 (error handling)
- Developer B: User Story 3 (gas tracking verification)
- Both can work in parallel since they enhance different aspects of the same code

---

## Success Criteria Mapping

**SC-001**: Users can stake NFT in <90 seconds
- Validated by: T024-T030 (deposit execution)
- Performance tracked by: T072-T077 (timestamp tracking)

**SC-002**: 100% accuracy when inputs valid
- Validated by: T011-T043 (all User Story 1 tasks)
- Error prevention: T044-T065 (User Story 2 validation)

**SC-003**: Gas tracking 100% accurate
- Validated by: T066-T086 (all User Story 3 tasks)
- Implementation: T019-T021, T027-T029 (gas extraction)

**SC-004**: Zero NFT loss on failure
- Validated by: T044-T065 (all User Story 2 error paths)
- Design: ERC721 atomic transfers (no intermediate state)

**SC-005**: 90% approval reuse effectiveness
- Implemented by: T015-T022 (approval check and conditional execution)
- Optimization: T022 (skip when approved)

**SC-006**: 100% actionable error messages
- Implemented by: T044-T065 (descriptive ErrorMessage for each scenario)
- Verified by: T087 (error message review)

---

## Notes

- [P] tasks = different files or independent error conditions
- [US1/US2/US3] labels map tasks to user stories from spec.md
- User Story 1 is the MVP - complete, independent implementation
- User Story 2 enhances US1 with comprehensive error handling
- User Story 3 verifies US1's transaction tracking accuracy
- No new files created - all work in blackhole.go (single method addition)
- Reuses existing types (StakingResult, TransactionRecord) from types.go
- Follows existing patterns from Mint method for consistency
- Constitutional compliance built into requirements (approval reuse, gas tracking, error safety)
- Total: 98 tasks organized across 6 phases
