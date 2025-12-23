# Tasks: Automated Liquidity Repositioning Strategy

**Input**: Design documents from `/specs/001-liquidity-repositioning/`
**Prerequisites**: plan.md, spec.md (user stories with priorities), research.md, data-model.md, contracts/strategy_api.go

**Tests**: Tests are included based on testing strategy (unit tests for algorithms, integration tests for full workflow)

**Organization**: Tasks grouped by user story priority (P1, P2) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and type definitions

- [x] T001 Create strategy types file at types.go with StrategyConfig, StrategyState, StrategyPhase
- [x] T002 Add StrategyReport type to types.go with JSON serialization
- [x] T003 [P] Add PositionRange type with IsOutOfRange, Width, Center methods to types.go
- [x] T004 [P] Add StabilityWindow type with CheckStability, Reset, Progress methods to types.go
- [x] T005 [P] Add CircuitBreaker type with RecordError, Reset, ErrorRate methods to types.go
- [x] T006 [P] Add PositionSnapshot type to types.go for reporting
- [x] T007 [P] Add RebalanceWorkflow type to types.go for tracking complete rebalancing operations

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core utility functions and helper methods that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T008 Implement StrategyConfig.Validate() method in types.go with all constraint checks per data-model.md
- [x] T009 Implement StrategyReport.ToJSON() method in types.go
- [x] T010 [P] Implement PositionRange.IsOutOfRange() in types.go using direct tick comparison (research.md R4)
- [x] T011 [P] Implement PositionRange.Width() and Center() helper methods in types.go
- [x] T012 [P] Implement StabilityWindow.CheckStability() in types.go using sliding window algorithm (research.md R2)
- [x] T013 [P] Implement CircuitBreaker.RecordError() in types.go with threshold and critical error logic (research.md R6)
- [x] T014 [P] Add helper function sqrtPriceToPrice() in internal/util/calculations.go for price conversions
- [x] T015 [P] Add helper function isCriticalError() in internal/util/validation.go for error classification
- [x] T016 [P] Add sendReport() helper function in blackhole.go for non-blocking channel sends

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Initial Position Entry (Priority: P1) 🎯 MVP

**Goal**: Automatically rebalance tokens to 50:50 ratio and create initial staked liquidity position

**Independent Test**: Start with unbalanced WAVAX/USDC, execute initial entry logic, verify staked position exists with correct ratio

### Implementation for User Story 1

- [x] T017 [P] [US1] Implement calculateRebalanceAmounts() helper in internal/util/calculations.go using value-based ratio algorithm (research.md R3)
- [x] T018 [P] [US1] Implement validateInitialBalances() in blackhole.go to check sufficient WAVAX/USDC per config
- [x] T019 [US1] Implement initialPositionEntry() in blackhole.go to orchestrate: validate balances → calculate rebalance → swap if needed → mint → stake
- [x] T020 [US1] Add integration with existing Swap method in initialPositionEntry() for token rebalancing
- [x] T021 [US1] Add integration with existing Mint method in initialPositionEntry() with RangeWidth from config
- [x] T022 [US1] Add integration with existing Stake method in initialPositionEntry() to stake minted NFT
- [x] T023 [US1] Add error handling and report sending to initialPositionEntry() for each step
- [x] T024 [US1] Update StrategyState after successful position creation in initialPositionEntry()

**Checkpoint**: Can create initial balanced position - US1 complete and testable

---

## Phase 4: User Story 3 - Automated Position Rebalancing (Priority: P1)

**Goal**: Automatically withdraw out-of-range positions, rebalance tokens, and re-create positions

**Independent Test**: Create an out-of-range position, execute rebalancing, verify old position withdrawn and new position staked in active range

**Dependencies**: Depends on US1 (need position entry logic), but can be tested independently by manually creating out-of-range position first

### Implementation for User Story 3

- [x] T025 [P] [US3] Implement executeUnstake() in blackhole.go to call existing Unstake method with correct nonce
- [x] T026 [P] [US3] Implement executeWithdraw() in blackhole.go to call existing Withdraw method and track results
- [x] T027 [US3] Implement executeRebalancing() in blackhole.go to orchestrate: unstake → withdraw → calculate rebalance → swap → update state
- [x] T028 [US3] Add RebalanceWorkflow creation and tracking in executeRebalancing()
- [x] T029 [US3] Integrate calculateRebalanceAmounts() from T017 for post-withdrawal ratio calculation
- [x] T030 [US3] Add cumulative gas tracking in executeRebalancing() for all transactions
- [x] T031 [US3] Add cumulative rewards tracking in executeRebalancing() from Unstake results
- [x] T032 [US3] Calculate and report net P&L in executeRebalancing(): (rewards - gas - fees)
- [x] T033 [US3] Send "rebalance_start", "gas_cost", "profit" reports during executeRebalancing()
- [x] T034 [US3] Add context cancellation checks before each major operation in executeRebalancing()

**Checkpoint**: Can rebalance out-of-range positions - US3 complete and testable

---

## Phase 5: User Story 2 - Continuous Price Monitoring (Priority: P2)

**Goal**: Continuously monitor pool price and detect out-of-range conditions

**Independent Test**: Create staked position, simulate price changes via mock GetAMMState, verify correct out-of-range detection

**Dependencies**: Logically depends on US1 (need position) and US3 (rebalancing action), but monitoring loop can be tested independently

### Implementation for User Story 2

- [ ] T035 [P] [US2] Implement monitoringLoop() in blackhole.go with ticker at MonitoringInterval
- [ ] T036 [US2] Add GetAMMState call in monitoringLoop() to fetch current pool state
- [ ] T037 [US2] Add PositionRange.IsOutOfRange() check in monitoringLoop() using current tick from AMMState
- [ ] T038 [US2] Add phase transition logic in monitoringLoop(): ActiveMonitoring → RebalancingRequired when out-of-range
- [ ] T039 [US2] Send "monitoring" and "out_of_range" reports in monitoringLoop()
- [ ] T040 [US2] Add context cancellation check in monitoringLoop() for graceful shutdown
- [ ] T041 [US2] Integrate monitoringLoop() with executeRebalancing() from US3 when out-of-range detected

**Checkpoint**: Can detect out-of-range and trigger rebalancing - US2 complete and testable

---

## Phase 6: User Story 4 - Price Stability Detection (Priority: P2)

**Goal**: Wait for price stabilization before re-entering position after withdrawal

**Independent Test**: Simulate price volatility after withdrawal, verify system waits for stability threshold before re-entry

**Dependencies**: Logically depends on US3 (rebalancing workflow), but stability algorithm can be tested independently

### Implementation for User Story 4

- [ ] T042 [P] [US4] Implement stabilityLoop() in blackhole.go with ticker at MonitoringInterval
- [ ] T043 [US4] Add GetAMMState call in stabilityLoop() to fetch current price
- [ ] T044 [US4] Add StabilityWindow.CheckStability() call in stabilityLoop() with sqrtPrice from AMMState
- [ ] T045 [US4] Add phase transition logic in stabilityLoop(): WaitingForStability → ExecutingRebalancing when stable
- [ ] T046 [US4] Add StabilityWindow.Reset() call if price exceeds threshold
- [ ] T047 [US4] Send "stability_check" reports with Progress() in stabilityLoop()
- [ ] T048 [US4] Add context cancellation check in stabilityLoop() for graceful shutdown
- [ ] T049 [US4] Integrate stabilityLoop() between executeRebalancing() and initialPositionEntry() re-entry

**Checkpoint**: Can wait for stability before re-entry - US4 complete and testable

---

## Phase 7: Main Strategy Integration

**Purpose**: Integrate all user stories into RunStrategy1 main method

**Dependencies**: All user stories (US1, US2, US3, US4) must be complete

- [ ] T050 Implement RunStrategy1(ctx, reportChan, config) method signature in blackhole.go
- [ ] T051 Add StrategyConfig.Validate() call at RunStrategy1 start
- [ ] T052 Initialize StrategyState with Initializing phase in RunStrategy1
- [ ] T053 Initialize CircuitBreaker with config parameters in RunStrategy1
- [ ] T054 Initialize StabilityWindow with config parameters in RunStrategy1
- [ ] T055 Send "strategy_start" report in RunStrategy1
- [ ] T056 Call initialPositionEntry() from US1 in Initializing phase
- [ ] T057 Transition to ActiveMonitoring phase after initial position created
- [ ] T058 Implement main loop in RunStrategy1 with select on ticker and context
- [ ] T059 Call monitoringLoop() from US2 in ActiveMonitoring phase
- [ ] T060 Call executeRebalancing() from US3 when RebalancingRequired phase
- [ ] T061 Call stabilityLoop() from US4 in WaitingForStability phase
- [ ] T062 Call initialPositionEntry() again after stability confirmed in ExecutingRebalancing phase
- [ ] T063 Transition back to ActiveMonitoring after re-entry complete
- [ ] T064 Add error handling with CircuitBreaker.RecordError() in main loop
- [ ] T065 Add "error" report sending for all error cases
- [ ] T066 Add "halt" report with final net P&L when circuit breaker triggers
- [ ] T067 Add "shutdown" report with final net P&L on context cancellation
- [ ] T068 Add cumulative gas, rewards, and net P&L calculations throughout RunStrategy1
- [ ] T069 Ensure all phase transitions update StrategyState.CurrentState
- [ ] T070 Add position state persistence to StrategyState (NFTTokenID, TickLower, TickUpper) after each position create

**Checkpoint**: Complete strategy functional - all user stories integrated

---

## Phase 8: Testing

**Purpose**: Comprehensive testing of all components and workflows

### Unit Tests

- [ ] T071 [P] Create tests/unit/strategy_config_test.go for StrategyConfig validation
- [ ] T072 [P] Create tests/unit/position_range_test.go for PositionRange methods with table-driven tests
- [ ] T073 [P] Create tests/unit/stability_window_test.go for StabilityWindow algorithm with various price scenarios
- [ ] T074 [P] Create tests/unit/circuit_breaker_test.go for CircuitBreaker logic (critical vs non-critical errors)
- [ ] T075 [P] Create tests/unit/strategy_report_test.go for StrategyReport JSON serialization
- [ ] T076 [P] Create tests/unit/calculations_test.go for calculateRebalanceAmounts() with various ratios
- [ ] T077 [P] Create tests/unit/error_classification_test.go for isCriticalError() function

### Integration Tests

- [ ] T078 Create tests/integration/strategy_integration_test.go test file
- [ ] T079 [P] Add mock RPC client implementation in tests/integration/mocks.go
- [ ] T080 [P] Add mock TxListener implementation in tests/integration/mocks.go
- [ ] T081 Test US1: Initial position entry with unbalanced tokens in strategy_integration_test.go
- [ ] T082 Test US1: Initial position entry with 100% WAVAX in strategy_integration_test.go
- [ ] T083 Test US1: Initial position entry with 100% USDC in strategy_integration_test.go
- [ ] T084 Test US2: Price monitoring detects in-range correctly in strategy_integration_test.go
- [ ] T085 Test US2: Price monitoring detects out-of-range (above) in strategy_integration_test.go
- [ ] T086 Test US2: Price monitoring detects out-of-range (below) in strategy_integration_test.go
- [ ] T087 Test US3: Full rebalancing workflow (unstake → withdraw → swap → mint → stake) in strategy_integration_test.go
- [ ] T088 Test US3: Rebalancing with accumulated fees in strategy_integration_test.go
- [ ] T089 Test US3: Rebalancing when tokens already 50:50 (no swap needed) in strategy_integration_test.go
- [ ] T090 Test US4: Stability detection during volatility (price > 0.5% change) in strategy_integration_test.go
- [ ] T091 Test US4: Stability detection after 5 consecutive stable intervals in strategy_integration_test.go
- [ ] T092 Test US4: Stability window reset when price becomes volatile again in strategy_integration_test.go
- [ ] T093 Test circuit breaker: Non-critical errors accumulate to threshold in strategy_integration_test.go
- [ ] T094 Test circuit breaker: Critical error triggers immediate halt in strategy_integration_test.go
- [ ] T095 Test graceful shutdown: Context cancellation stops strategy safely in strategy_integration_test.go
- [ ] T096 Test full 24-hour simulation: Continuous operation through multiple rebalances in strategy_integration_test.go
- [ ] T097 Test edge case: Insufficient balance after swap in strategy_integration_test.go
- [ ] T098 Test edge case: Swap failure due to slippage in strategy_integration_test.go
- [ ] T099 Test edge case: NFT not owned error in strategy_integration_test.go
- [ ] T100 Test edge case: Network failure during withdrawal in strategy_integration_test.go

**Checkpoint**: All tests passing - ready for polish phase

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, optimization, and final validation

- [ ] T101 [P] Add comprehensive GoDoc comments to all public types and methods in types.go
- [ ] T102 [P] Add comprehensive GoDoc comments to RunStrategy1 method in blackhole.go
- [ ] T103 [P] Verify all reports sent match 12 event types from data-model.md
- [ ] T104 [P] Validate JSON report format matches StrategyReport schema
- [ ] T105 [P] Add logging for all phase transitions with context
- [ ] T106 [P] Optimize sqrtPrice to price conversions for gas efficiency
- [ ] T107 Validate constitutional compliance: Principle 1 (WAVAX/USDC only) in code review
- [ ] T108 Validate constitutional compliance: Principle 2 (autonomous operation) in code review
- [ ] T109 Validate constitutional compliance: Principle 3 (financial transparency) in code review
- [ ] T110 Validate constitutional compliance: Principle 4 (gas optimization) in code review
- [ ] T111 Validate constitutional compliance: Principle 5 (fail-safe operation) in code review
- [ ] T112 Run quickstart.md example code validation
- [ ] T113 Performance test: Verify rebalancing cycle completes in < 10 minutes
- [ ] T114 Performance test: Verify out-of-range detection within 2 intervals (SC-005)
- [ ] T115 Performance test: Verify 24-hour continuous operation without intervention (SC-008)
- [ ] T116 Verify token ratio deviation < 1% after rebalancing (SC-003)
- [ ] T117 Verify gas costs < 2% of position value per rebalance (SC-004)
- [ ] T118 Verify financial tracking accuracy within 0.1% (SC-007)
- [ ] T119 Code cleanup: Remove any debug logging or commented code
- [ ] T120 Final integration test: Run full strategy with real RPC on testnet

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup (Phase 1) completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational (Phase 2) completion
  - **US1 Initial Position Entry (Phase 3, P1)**: Can start after Foundational - No story dependencies
  - **US3 Automated Rebalancing (Phase 4, P1)**: Can start after Foundational - Logically depends on US1 but can be tested independently
  - **US2 Continuous Monitoring (Phase 5, P2)**: Can start after Foundational - Integrates with US1 and US3 but testable independently
  - **US4 Price Stability (Phase 6, P2)**: Can start after Foundational - Integrates with US3 but testable independently
- **Main Integration (Phase 7)**: Depends on all user stories (Phase 3-6) completion
- **Testing (Phase 8)**: Depends on Main Integration (Phase 7) completion
- **Polish (Phase 9)**: Depends on Testing (Phase 8) completion

### User Story Dependencies

- **US1 (P1 - Initial Entry)**: Independent - only depends on Foundational phase
- **US3 (P1 - Rebalancing)**: Logically uses US1 position entry, but can be tested with manually created positions
- **US2 (P2 - Monitoring)**: Integrates US1 (creates position) and US3 (rebalances), but monitoring logic is independent
- **US4 (P2 - Stability)**: Integrates with US3 (rebalancing workflow), but stability algorithm is independent

### Within Each User Story

- Tasks within a story follow logical order (helpers before usage)
- Tasks marked [P] can run in parallel within the same story phase
- All tasks for a story must complete before story checkpoint

### Parallel Opportunities

**Setup Phase (Phase 1)**:
```bash
# All type definitions (T003, T004, T005, T006, T007) can be created in parallel
```

**Foundational Phase (Phase 2)**:
```bash
# All implementation tasks marked [P] (T010-T016) can run in parallel after T008-T009
```

**User Story Phases**:
```bash
# After Foundational complete, all user stories can start in parallel:
- Developer A: US1 (T017-T024)
- Developer B: US3 (T025-T034)
- Developer C: US2 (T035-T041)
- Developer D: US4 (T042-T049)

# Within each story, tasks marked [P] can run in parallel
```

**Testing Phase**:
```bash
# All unit tests (T071-T077) can run in parallel
# Integration test setup (T079-T080) can run in parallel
```

---

## Parallel Example: User Story 1

```bash
# T017 and T018 can start in parallel (different functions):
Task T017: "Implement calculateRebalanceAmounts() in internal/util/calculations.go"
Task T018: "Implement validateInitialBalances() in blackhole.go"

# After both complete, continue with T019
```

---

## Parallel Example: Foundational Phase

```bash
# After T008-T009 complete, all algorithm implementations can start:
Task T010: "Implement PositionRange.IsOutOfRange()"
Task T011: "Implement PositionRange.Width() and Center()"
Task T012: "Implement StabilityWindow.CheckStability()"
Task T013: "Implement CircuitBreaker.RecordError()"
Task T014: "Add sqrtPriceToPrice() helper"
Task T015: "Add isCriticalError() helper"
Task T016: "Add sendReport() helper"
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 3 Only - Both P1)

1. Complete Phase 1: Setup (T001-T007)
2. Complete Phase 2: Foundational (T008-T016) - **CRITICAL CHECKPOINT**
3. Complete Phase 3: US1 Initial Position Entry (T017-T024)
4. Complete Phase 4: US3 Automated Rebalancing (T025-T034)
5. **STOP and VALIDATE**: Test US1 and US3 together as minimal viable strategy
6. This gives: initial position creation + rebalancing when out-of-range (core value!)

### Full Feature (Add Monitoring and Stability)

1. Complete MVP (Phases 1-4)
2. Add Phase 5: US2 Continuous Monitoring (T035-T041) - enables autonomous operation
3. Add Phase 6: US4 Price Stability (T042-T049) - optimizes gas costs
4. Complete Phase 7: Main Integration (T050-T070) - brings everything together
5. Complete Phase 8: Testing (T071-T100)
6. Complete Phase 9: Polish (T101-T120)

### Parallel Team Strategy

With 4 developers after Foundational complete:
- **Developer A**: Phase 3 (US1 - T017-T024) → Then help with Integration
- **Developer B**: Phase 4 (US3 - T025-T034) → Then help with Integration
- **Developer C**: Phase 5 (US2 - T035-T041) → Then help with Testing
- **Developer D**: Phase 6 (US4 - T042-T049) → Then help with Testing

All developers converge on Phase 7 (Integration) once their stories complete.

---

## Notes

- **[P] tasks**: Different files or independent functions, can run in parallel
- **[Story] label**: Maps task to specific user story (US1, US2, US3, US4) for traceability
- **Constitutional compliance**: Tasks T107-T111 verify all 5 principles are satisfied in implementation
- **Success criteria**: Tasks T113-T118 verify measurable outcomes from spec.md
- **Test coverage**: Unit tests for algorithms (T071-T077), integration tests for workflows (T078-T100)
- **MVP scope**: Phases 1-4 deliver minimum viable strategy (initial entry + rebalancing)
- **Full feature**: All phases deliver complete autonomous strategy with monitoring and stability detection
- **Each user story checkpoint**: Story should be independently completable and demonstrable
- **Commit strategy**: Commit after each task or logical group of [P] tasks
- **Context checks**: All loops (monitoring, stability, main) must check ctx.Done() for graceful shutdown
- **Report channel**: Use non-blocking sends (T016 helper) to prevent strategy deadlock
- **Financial tracking**: Cumulative gas, rewards, fees tracked throughout (T030-T032, T068)

---

## Quick Reference: File Locations

| Component | File Path |
|-----------|-----------|
| Type definitions | `types.go` |
| Main strategy method | `blackhole.go` (RunStrategy1) |
| Helper calculations | `internal/util/calculations.go` |
| Helper validation | `internal/util/validation.go` |
| Unit tests | `tests/unit/strategy_*_test.go` |
| Integration tests | `tests/integration/strategy_integration_test.go` |
| Test mocks | `tests/integration/mocks.go` |

---

## Summary

- **Total Tasks**: 120
- **Setup**: 7 tasks
- **Foundational**: 9 tasks (blocks all stories)
- **US1 (P1 - Initial Entry)**: 8 tasks
- **US3 (P1 - Rebalancing)**: 10 tasks
- **US2 (P2 - Monitoring)**: 7 tasks
- **US4 (P2 - Stability)**: 8 tasks
- **Main Integration**: 21 tasks
- **Testing**: 30 tasks (7 unit + 23 integration)
- **Polish**: 20 tasks

**MVP Scope** (Phases 1-4): 34 tasks → Delivers initial entry + rebalancing
**Full Feature** (All Phases): 120 tasks → Complete autonomous strategy

**Parallel Opportunities Identified**:
- Setup: 5 parallel tasks
- Foundational: 7 parallel tasks
- User Stories: All 4 can start in parallel after Foundational (33 tasks parallelizable)
- Testing: 9 parallel tasks

**Independent Test Criteria**:
- US1: Start with unbalanced tokens → verify balanced staked position
- US2: Create position, simulate price change → verify out-of-range detection
- US3: Create out-of-range position → verify full rebalancing workflow
- US4: Simulate volatility → verify stability wait before re-entry
