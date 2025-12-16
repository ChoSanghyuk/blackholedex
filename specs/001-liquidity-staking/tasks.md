---

description: "Task list for Liquidity Staking feature implementation"
---

# Tasks: Liquidity Staking

**Input**: Design documents from `/specs/001-liquidity-staking/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are NOT explicitly requested in the feature specification. Test tasks are included as OPTIONAL for validation purposes.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Single Go project: root directory for main logic, `internal/` for utilities, `pkg/` for reusable packages
- Paths shown below follow existing blackhole_dex structure

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure needed for all user stories

- [x] T001 Add NonfungiblePositionManager constant to blackhole.go contract addresses section
- [x] T002 [P] Create internal/util/validation.go file with package declaration and imports
- [x] T003 [P] Add StakingResult type definition to types.go with all fields from data model
- [x] T004 [P] Add TransactionRecord type definition to types.go with all fields from data model

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Implement validateStakingRequest function in internal/util/validation.go (range width 1-20, slippage 1-50, amounts > 0)
- [x] T006 Implement calculateTickBounds function in internal/util/validation.go (tick calculation with spacing 200, validation within ±887272)
- [x] T007 [P] Implement calculateMinAmount helper function in internal/util/validation.go (slippage protection calculation using big.Int)
- [x] T008 [P] Implement extractGasCost function in internal/util/validation.go (gas cost calculation from receipt)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Single-Step Liquidity Staking (Priority: P1) 🎯 MVP

**Goal**: Enable operator to stake liquidity in WAVAX-USDC pool with automatic position calculation, approvals, and minting

**Independent Test**: Call Mint with specific token amounts and range parameters, verify resulting liquidity position exists on-chain with correct tick bounds and token amounts

### Implementation for User Story 1

- [x] T009 [US1] Implement validateBalances method in blackhole.go (query WAVAX and USDC balances, compare with required amounts)
- [x] T010 [US1] Implement ensureApproval method in blackhole.go (check allowance, approve only if needed, return tx hash)
- [x] T011 [US1] Update Mint method signature in blackhole.go to accept maxWAVAX, maxUSDC, rangeWidth, slippagePct parameters
- [x] T012 [US1] Add input validation to Mint method in blackhole.go (call validateStakingRequest from T005)
- [x] T013 [US1] Add pool state query to Mint method in blackhole.go (call existing GetAMMState, handle errors)
- [x] T014 [US1] Add tick bounds calculation to Mint method in blackhole.go (call calculateTickBounds from T006)
- [x] T015 [US1] Add optimal amounts calculation to Mint method in blackhole.go (call existing ComputeAmounts from internal/util/amm.go)
- [x] T016 [US1] Add balance validation to Mint method in blackhole.go (call validateBalances from T009)
- [x] T017 [US1] Add slippage protection calculation to Mint method in blackhole.go (call calculateMinAmount from T007 for both tokens)
- [x] T018 [US1] Add WAVAX approval logic to Mint method in blackhole.go (call ensureApproval from T010, wait for confirmation if tx sent)
- [x] T019 [US1] Add USDC approval logic to Mint method in blackhole.go (call ensureApproval from T010, wait for confirmation if tx sent)
- [x] T020 [US1] Construct MintParams in Mint method in blackhole.go (use existing type, populate all fields including deadline)
- [x] T021 [US1] Get NonfungiblePositionManager client in Mint method in blackhole.go (call b.Client with nonfungiblePositionManager constant)
- [x] T022 [US1] Submit mint transaction in Mint method in blackhole.go (call Send on NFT manager client with MintParams)
- [x] T023 [US1] Wait for mint confirmation in Mint method in blackhole.go (call b.tl.WaitForTransaction)
- [x] T024 [US1] Extract gas costs for all transactions in Mint method in blackhole.go (call extractGasCost from T008 for each tx)
- [x] T025 [US1] Parse mint receipt for NFT token ID in Mint method in blackhole.go (extract token ID from event logs)
- [x] T026 [US1] Construct StakingResult in Mint method in blackhole.go (populate all fields, calculate total gas cost)
- [x] T027 [US1] Add comprehensive error handling to Mint method in blackhole.go (wrap errors with context at each step)
- [x] T028 [US1] Add transaction logging to Mint method in blackhole.go (log tx hashes, gas costs, amounts for financial tracking)

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Configurable Range Width (Priority: P2)

**Goal**: Enable operator to adjust position concentration via range width parameter

**Independent Test**: Create multiple positions with different range width parameters (2, 4, 6, 8) and verify tick bounds match expected widths

### Implementation for User Story 2

- [x] T029 [US2] Add range width validation edge cases to validateStakingRequest in internal/util/validation.go (test boundary values 1 and 20)
- [x] T030 [US2] Add tick bounds calculation edge case handling to calculateTickBounds in internal/util/validation.go (handle extreme ticks near ±887272)
- [x] T031 [US2] Add range width examples to error messages in validateStakingRequest in internal/util/validation.go (show valid examples in error text)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Balanced Token Calculation (Priority: P2)

**Goal**: Automatically calculate optimal token ratio for maximum capital deployment

**Independent Test**: Provide various unbalanced token amounts, verify system stakes maximum possible while maintaining correct ratio

### Implementation for User Story 3

- [x] T032 [US3] Add capital efficiency warning to Mint method in blackhole.go (log warning when >10% of either token unused)
- [x] T033 [US3] Add actual vs desired amount comparison to Mint method in blackhole.go (compare ComputeAmounts result with max amounts)
- [x] T034 [US3] Add capital utilization logging to Mint method in blackhole.go (log percentage of each token used)

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Testing & Validation (OPTIONAL)

**Purpose**: Validation of all user stories for production readiness

- [ ] T035 [P] Create blackhole_test.go with test setup (initialize Blackhole client with test RPC)
- [ ] T036 [P] [US1] Add integration test for balanced staking in blackhole_test.go (test with 10 WAVAX, 500 USDC, range width 6)
- [ ] T037 [P] [US1] Add integration test for unbalanced staking in blackhole_test.go (test with 5 WAVAX, 1000 USDC, verify optimal usage)
- [ ] T038 [P] [US1] Add integration test for slippage protection in blackhole_test.go (mock price movement, verify tx failure)
- [ ] T039 [P] [US1] Add integration test for gas cost tracking in blackhole_test.go (verify all tx costs recorded correctly)
- [ ] T040 [P] [US1] Add integration test for approval reuse in blackhole_test.go (stake twice, verify second stake skips approvals if sufficient)
- [ ] T041 [P] [US2] Add integration test for range width validation in blackhole_test.go (test invalid range width 0, verify error)
- [ ] T042 [P] [US2] Add integration test for multiple range widths in blackhole_test.go (test range widths 2, 6, 10, verify tick bounds)
- [ ] T043 [P] [US3] Add integration test for capital utilization in blackhole_test.go (test extreme imbalance, verify warning logged)
- [ ] T044 Add integration test for insufficient balance error in blackhole_test.go (test with insufficient WAVAX, verify error before any tx)
- [ ] T045 Add integration test for network failure handling in blackhole_test.go (mock RPC timeout, verify safe failure)

---

## Phase 7: CLI Tool (OPTIONAL for manual operator testing)

**Purpose**: Command-line interface for manual staking operations

- [ ] T046 Create cmd/stake/main.go with package main and imports
- [ ] T047 Implement loadEnv function in cmd/stake/main.go (load PRIVATE_KEY and RPC_URL from environment)
- [ ] T048 Implement initializeBlackhole function in cmd/stake/main.go (create ContractClient map, TxListener, Blackhole instance)
- [ ] T049 Implement parseAmounts function in cmd/stake/main.go (parse command-line args for WAVAX, USDC amounts)
- [ ] T050 Implement main function in cmd/stake/main.go (load env, parse args, call Mint, display results)
- [ ] T051 [P] Add formatting helpers to cmd/stake/main.go (formatWAVAX, formatUSDC, formatAVAX for human-readable output)

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T052 [P] Add detailed code comments to Mint method in blackhole.go (explain each calculation step)
- [ ] T053 [P] Add detailed code comments to validation functions in internal/util/validation.go (explain validation rules)
- [ ] T054 Review and optimize error messages across all functions in blackhole.go and internal/util/validation.go (ensure clarity and actionability)
- [ ] T055 Add logging for constitutional compliance in Mint method in blackhole.go (verify all Principle 3 requirements logged)
- [ ] T056 Verify quickstart.md examples in specs/001-liquidity-staking/quickstart.md match actual implementation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User Story 1 (P1): Can start after Foundational (Phase 2) - No dependencies on other stories
  - User Story 2 (P2): Can start after Foundational (Phase 2) - Builds on US1 validation but independently testable
  - User Story 3 (P3): Can start after Foundational (Phase 2) - Enhances US1 but independently testable
- **Testing (Phase 6)**: Depends on all user stories being complete (OPTIONAL)
- **CLI Tool (Phase 7)**: Depends on User Story 1 being complete (OPTIONAL)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - Core staking functionality, no dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Adds validation enhancements to US1 functions, independently testable
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - Adds logging enhancements to US1, independently testable

### Within Each User Story

- **User Story 1 Tasks**: Must execute sequentially due to dependencies:
  - T009-T010 can run in parallel (different methods)
  - T011-T028 must run sequentially (building up Mint method)
- **User Story 2 Tasks**: T029-T031 can run in parallel (different validation functions)
- **User Story 3 Tasks**: T032-T034 can run in parallel (different logging additions to Mint)

### Parallel Opportunities

- **Setup phase (T001-T004)**: All tasks marked [P] can run in parallel (different files/sections)
- **Foundational phase (T007-T008)**: Tasks marked [P] can run in parallel (different helper functions)
- **User Story 1**: T009-T010 can run in parallel (different methods in blackhole.go)
- **User Story 2**: All tasks (T029-T031) can run in parallel (different validation functions)
- **User Story 3**: All tasks (T032-T034) can run in parallel (different logging additions)
- **Testing phase**: All test tasks marked [P] (T035-T043) can run in parallel (different test functions)
- **CLI Tool**: T051 can run in parallel with other CLI tasks
- **Polish phase**: T052-T053 can run in parallel (different files)

---

## Parallel Example: User Story 1 Foundation

```bash
# These can be launched together (different methods in blackhole.go):
Task T009: "Implement validateBalances method"
Task T010: "Implement ensureApproval method"

# After both complete, proceed with sequential Mint method construction (T011-T028)
```

---

## Parallel Example: Foundational Phase

```bash
# These can be launched together (different helper functions):
Task T007: "Implement calculateMinAmount helper function in internal/util/validation.go"
Task T008: "Implement extractGasCost function in internal/util/validation.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only - Recommended)

1. Complete Phase 1: Setup (T001-T004)
2. Complete Phase 2: Foundational (T005-T008) - CRITICAL blocking phase
3. Complete Phase 3: User Story 1 (T009-T028)
4. **STOP and VALIDATE**: Test User Story 1 independently
   - Verify staking with balanced amounts works
   - Verify staking with unbalanced amounts works
   - Verify slippage protection functions
   - Verify gas costs tracked correctly
   - Verify approval reuse works
5. Deploy/demo if ready - this is a complete, usable MVP

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo (enhanced validation)
4. Add User Story 3 → Test independently → Deploy/Demo (capital efficiency warnings)
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T001-T008)
2. Once Foundational is done:
   - Developer A: User Story 1 (T009-T028) - Core functionality
   - Developer B: Testing infrastructure (T035) - Prepare test framework
   - Developer C: CLI Tool (T046-T051) - Manual testing interface
3. After User Story 1 complete:
   - Developer A: User Story 2 (T029-T031)
   - Developer B: User Story 1 tests (T036-T040)
   - Developer C: User Story 3 (T032-T034)
4. Stories complete and integrate independently

---

## Task Count Summary

- **Total Tasks**: 56
- **Setup Phase**: 4 tasks
- **Foundational Phase**: 4 tasks (BLOCKING)
- **User Story 1 (P1)**: 20 tasks (Core MVP)
- **User Story 2 (P2)**: 3 tasks (Validation enhancements)
- **User Story 3 (P2)**: 3 tasks (Capital efficiency)
- **Testing Phase**: 11 tasks (OPTIONAL)
- **CLI Tool Phase**: 6 tasks (OPTIONAL)
- **Polish Phase**: 5 tasks

**MVP Scope** (Recommended first delivery):
- Phase 1: Setup (4 tasks)
- Phase 2: Foundational (4 tasks)
- Phase 3: User Story 1 (20 tasks)
- **Total MVP**: 28 tasks

**Parallel Opportunities**: 15 tasks can run in parallel (marked with [P])

**Independent Testing**:
- User Story 1: Verify by calling Mint with test amounts, check on-chain position
- User Story 2: Verify by testing multiple range widths, check tick bounds calculation
- User Story 3: Verify by testing unbalanced amounts, check capital utilization warnings

---

## Notes

- [P] tasks = different files/functions, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Testing tasks (Phase 6) are OPTIONAL but recommended for production readiness
- CLI Tool (Phase 7) is OPTIONAL but useful for manual operator testing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence

---

## File Change Summary

**New Files**:
- `internal/util/validation.go` - Validation and helper functions
- `blackhole_test.go` - Integration tests (OPTIONAL)
- `cmd/stake/main.go` - CLI tool (OPTIONAL)

**Modified Files**:
- `blackhole.go` - Complete Mint method, add validateBalances and ensureApproval methods, add nonfungiblePositionManager constant
- `types.go` - Add StakingResult and TransactionRecord types

**Existing Files Used**:
- `internal/util/amm.go` - ComputeAmounts function (no changes needed)
- `pkg/contractclient/contractclient.go` - ContractClient interface (no changes needed)
- `pkg/txlistener/txlistener.go` - TxListener interface (no changes needed)
- `types.go` - MintParams type (already exists, no changes needed)

**Total Files**: 2 new files (required), 2 new files (optional), 2 modified files
