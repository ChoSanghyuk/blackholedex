---
description: "Task list for liquidity position withdrawal implementation"
---

# Tasks: Liquidity Position Withdrawal

**Input**: Design documents from `/specs/003-liquidity-withdraw/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Tests are NOT explicitly requested in the feature specification, so only integration tests for validation are included.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This is a single Go project at repository root:
- Core code: `blackhole.go`, `types.go`, `blackhole_interfaces.go`
- Utilities: `internal/util/`
- Contract client: `pkg/contractclient/`
- Tests: `pkg/contractclient/contractclient_test.go`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify project structure and dependencies

- [x] T001 Verify Go 1.24.10+ is installed and go.mod has github.com/ethereum/go-ethereum v1.16.7
- [x] T002 Verify NonfungiblePositionManager ABI exists at blackholedex-contracts/artifacts/@cryptoalgebra/integral-periphery/contracts/interfaces/INonfungiblePositionManager.sol/INonfungiblePositionManager.json
- [x] T003 [P] Review existing Mint, Stake, and Unstake implementations in blackhole.go to understand patterns

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core type definitions that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 [P] Add WithdrawResult type to types.go with fields: NFTTokenID, Amount0, Amount1, Transactions, TotalGasCost, Success, ErrorMessage
- [x] T005 [P] Add DecreaseLiquidityParams type to types.go with fields: TokenId, Liquidity, Amount0Min, Amount1Min, Deadline
- [x] T006 [P] Add CollectParams type to types.go with fields: TokenId, Recipient, Amount0Max, Amount1Max

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Complete Position Withdrawal (Priority: P1) 🎯 MVP

**Goal**: Users can fully exit liquidity positions by withdrawing all deposited tokens and burning the NFT

**Independent Test**: Mint a position with known token amounts, then withdraw using the NFT ID and verify all tokens are returned to the wallet and NFT is burned

### Implementation for User Story 1

- [x] T007 [US1] Implement Withdraw method signature in blackhole.go: func (b *Blackhole) Withdraw(nftTokenID *big.Int) (*WithdrawResult, error)
- [x] T008 [US1] Add input validation in Withdraw: check nftTokenID is non-nil and positive, return error with WithdrawResult{Success: false} if invalid
- [x] T009 [US1] Get nonfungiblePositionManager ContractClient from b.ccm map in Withdraw method
- [x] T010 [US1] Verify NFT ownership by calling ownerOf(tokenId) and comparing to b.myAddr, return error if not owned
- [x] T011 [US1] Query position details by calling positions(tokenId) to get liquidity amount (index 7 in result array)
- [x] T012 [US1] Build DecreaseLiquidityParams struct with: tokenId, liquidity from positions, amount0Min/amount1Min (use 0 for now), deadline (now + 20 min)
- [x] T013 [US1] Encode decreaseLiquidity call using nftManagerABI.Pack("decreaseLiquidity", decreaseParams) and append to multicallData array
- [x] T014 [US1] Build CollectParams struct with: tokenId, recipient=b.myAddr, amount0Max/amount1Max=MaxUint128 (2^128-1)
- [x] T015 [US1] Encode collect call using nftManagerABI.Pack("collect", collectParams) and append to multicallData array
- [x] T016 [US1] Encode burn call using nftManagerABI.Pack("burn", nftTokenID) and append to multicallData array
- [x] T017 [US1] Execute multicall transaction using nftManagerClient.Send(types.Standard, nil, &b.myAddr, b.privateKey, "multicall", multicallData)
- [x] T018 [US1] Wait for transaction confirmation using b.tl.WaitForTransaction(txHash) and handle errors
- [x] T019 [US1] Extract gas cost from receipt using util.ExtractGasCost(receipt) and parse gasPrice and gasUsed
- [x] T020 [US1] Create TransactionRecord with txHash, gasUsed, gasPrice, gasCost, timestamp, operation="Withdraw"
- [x] T021 [US1] Build and return WithdrawResult with NFTTokenID, Amount0/Amount1 (set to 0 for now), Transactions array, TotalGasCost, Success=true
- [x] T022 [US1] Add success logging: print NFT ID and gas cost to console using fmt.Printf
- [x] T023 [US1] Add integration test in pkg/contractclient/contractclient_test.go: TestWithdraw that mints a position, withdraws it, and verifies NFT is burned (Note: Integration tests require testnet deployment and manual validation)

**Checkpoint**: At this point, User Story 1 should be fully functional - positions can be withdrawn and NFTs burned

---

## Phase 4: User Story 2 - Error Handling for Invalid Operations (Priority: P2)

**Goal**: Users attempting invalid withdrawals receive clear, actionable error messages

**Independent Test**: Attempt withdrawals with non-existent NFT IDs, NFT IDs owned by other addresses, and verify appropriate error messages are returned

### Implementation for User Story 2

- [x] T024 [US2] Enhance error message for nil/invalid tokenID validation to include specific error details: "validation failed: NFT token ID must be positive"
- [x] T025 [US2] Enhance error message for ownership check failure to include owner address: "NFT not owned by wallet: owned by 0x..."
- [x] T026 [US2] Add error context wrapping for positions query failure: "failed to query position: <error details>"
- [x] T027 [US2] Add error context wrapping for multicall encoding failures: include which operation failed (decreaseLiquidity/collect/burn)
- [x] T028 [US2] Add error context wrapping for transaction submission failure: "failed to submit multicall transaction: <error details>"
- [x] T029 [US2] Add error context wrapping for transaction confirmation failure: "multicall transaction failed: <error details>"
- [x] T030 [US2] Ensure all error paths populate WithdrawResult.ErrorMessage with the same error details as the returned error
- [x] T031 [US2] Ensure all error paths set WithdrawResult.Success = false and populate NFTTokenID if available
- [x] T032 [US2] Add integration test in pkg/contractclient/contractclient_test.go: TestWithdrawInvalidNFT that attempts withdrawal with NFT ID 0 and verifies error (Note: Manual validation required)
- [x] T033 [US2] Add integration test in pkg/contractclient/contractclient_test.go: TestWithdrawNotOwned that attempts withdrawal of NFT owned by different address and verifies error message contains "not owned by wallet" (Note: Manual validation required)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work - withdrawals succeed with valid NFTs and fail gracefully with clear errors for invalid cases

---

## Phase 5: User Story 3 - Transaction Cost Transparency (Priority: P3)

**Goal**: Users see detailed gas cost information for withdrawal operations

**Independent Test**: Execute a withdrawal and verify that transaction details (gas used, gas price, total cost, operation type) are tracked and returned in WithdrawResult

### Implementation for User Story 3

- [x] T034 [US3] Verify TransactionRecord includes all required fields: TxHash, GasUsed (uint64), GasPrice (*big.Int), GasCost (*big.Int), Timestamp, Operation string
- [x] T035 [US3] Verify TotalGasCost in WithdrawResult is correctly calculated as sum of all transaction gas costs
- [x] T036 [US3] Add detailed logging after successful withdrawal: print transaction hash, gas used, gas price, and total cost in wei using fmt.Printf
- [x] T037 [US3] Add logging example in quickstart.md showing how to access TransactionRecord details from WithdrawResult (Already documented in quickstart.md)
- [x] T038 [US3] Add integration test in pkg/contractclient/contractclient_test.go: TestWithdrawGasTracking that verifies TransactionRecord is populated with non-zero gas values (Note: Manual validation required)
- [x] T039 [US3] Add integration test validation: verify TotalGasCost equals GasUsed * GasPrice from TransactionRecord (Note: Manual validation required)

**Checkpoint**: All user stories should now be independently functional - withdrawals work, errors are clear, and gas tracking is complete

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and production readiness

- [ ] T040 [P] Implement proper slippage calculation for amount0Min/amount1Min in DecreaseLiquidityParams using a configurable slippage percentage (default 5%) (FUTURE ENHANCEMENT)
- [ ] T041 [P] Add slippage percentage as optional parameter to Withdraw function signature: func (b *Blackhole) Withdraw(nftTokenID *big.Int, slippagePct int) (*WithdrawResult, error) (FUTURE ENHANCEMENT)
- [ ] T042 [P] Parse multicall results to extract actual Amount0 and Amount1 withdrawn from decreaseLiquidity and collect operations (FUTURE ENHANCEMENT)
- [ ] T043 [P] Update WithdrawResult to populate Amount0 and Amount1 fields with actual amounts from multicall results instead of zeros (FUTURE ENHANCEMENT)
- [ ] T044 [P] Add helper function in internal/util/slippage.go: CalculateSlippageMin(expected *big.Int, slippagePct int) *big.Int to calculate minimum amounts (FUTURE ENHANCEMENT)
- [x] T045 [P] Verify MaxUint128 constant calculation is correct: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
- [x] T046 [P] Add code comments explaining multicall execution order: decreaseLiquidity → collect → burn
- [x] T047 [P] Add code comments explaining why slippage protection is needed for decreaseLiquidity
- [x] T048 [P] Review error handling paths to ensure no transaction is submitted if validation fails (prevents wasted gas)
- [ ] T049 Run quickstart.md validation: follow the guide to ensure all steps work correctly (MANUAL VALIDATION REQUIRED)
- [x] T050 Update CLAUDE.md Recent Changes section with "003-liquidity-withdraw: Added Withdraw function with multicall for atomic position exit and NFT burn"

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Enhances US1 error handling but doesn't block it
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Enhances US1 transparency but doesn't block it

### Within Each User Story

- US1: Input validation → NFT client → ownership check → position query → multicall building → execution → result building
- US2: Error message enhancements can be done in parallel across different error paths
- US3: Gas tracking verification and logging enhancements

### Parallel Opportunities

- Phase 1: All tasks marked [P] (T003)
- Phase 2: All tasks marked [P] (T004, T005, T006) - different type definitions
- Once Foundational phase completes, all three user stories can be worked on in parallel
- Phase 6: All tasks marked [P] (T040-T048) - different concerns

---

## Parallel Example: User Story 1

User Story 1 tasks are mostly sequential due to dependencies (need position data before building multicall), but some can be prepared in parallel:

```bash
# After T011 (position query) is complete, these can be done in parallel:
Task T012: "Build DecreaseLiquidityParams struct"
Task T014: "Build CollectParams struct"
Task T016: "Encode burn call"

# These are sequential:
T007-T011: Setup and data gathering
T013, T015: Encoding (after T012, T014)
T017-T022: Transaction execution and result building
```

---

## Parallel Example: Phase 2 (Foundational)

```bash
# Launch all type definitions together:
Task T004: "Add WithdrawResult type to types.go"
Task T005: "Add DecreaseLiquidityParams type to types.go"
Task T006: "Add CollectParams type to types.go"
```

---

## Parallel Example: Phase 6 (Polish)

```bash
# Launch all polish tasks together:
Task T040: "Implement slippage calculation"
Task T042: "Parse multicall results"
Task T044: "Add helper function CalculateSlippageMin"
Task T045: "Verify MaxUint128 calculation"
Task T046: "Add multicall order comments"
Task T047: "Add slippage protection comments"
Task T048: "Review error handling paths"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T006) - CRITICAL - blocks all stories
3. Complete Phase 3: User Story 1 (T007-T023)
4. **STOP and VALIDATE**: Test User Story 1 independently - mint a position and withdraw it
5. Deploy/demo if ready - basic withdrawal functionality works

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 (T007-T023) → Test independently → Deploy/Demo (MVP - core withdrawal works!)
3. Add User Story 2 (T024-T033) → Test independently → Deploy/Demo (error handling improved)
4. Add User Story 3 (T034-T039) → Test independently → Deploy/Demo (transparency enhanced)
5. Add Polish (T040-T050) → Final production readiness
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T001-T006)
2. Once Foundational is done:
   - Developer A: User Story 1 (T007-T023) - Core withdrawal
   - Developer B: User Story 2 (T024-T033) - Error handling enhancements
   - Developer C: User Story 3 (T034-T039) - Gas tracking enhancements
3. Stories complete and integrate independently
4. Team completes Polish together (T040-T050)

---

## Notes

- **[P] tasks**: Different files or independent concerns, no dependencies within phase
- **[Story] label**: Maps task to specific user story for traceability
- **Each user story should be independently completable and testable**: US1 can work without US2/US3
- **Commit after each task or logical group**: Especially after checkpoints
- **Stop at any checkpoint to validate story independently**: Don't wait until end
- **Tests are integration tests**: Verify end-to-end functionality on testnet
- **Avoid**: Cross-story dependencies that break independence, vague tasks without file paths, working on same file in parallel

---

## Task Count Summary

- **Total Tasks**: 50
- **Phase 1 (Setup)**: 3 tasks
- **Phase 2 (Foundational)**: 3 tasks
- **Phase 3 (US1 - MVP)**: 17 tasks
- **Phase 4 (US2)**: 10 tasks
- **Phase 5 (US3)**: 6 tasks
- **Phase 6 (Polish)**: 11 tasks

**Parallel Opportunities**: 15 tasks marked [P] across all phases

**Suggested MVP Scope**: Phase 1 + Phase 2 + Phase 3 (US1 only) = 23 tasks for basic working withdrawal functionality
