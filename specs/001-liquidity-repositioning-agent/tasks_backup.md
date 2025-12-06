# Tasks: Blackhole DEX Liquidity Repositioning Agent

**Input**: Design documents from `/specs/001-liquidity-repositioning-agent/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `blackhole_dex/` at repository root
- Paths shown below use absolute paths from repo root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create agent CLI directory structure at blackhole_dex/cmd/agent/
- [ ] T002 Create Blackhole package directory at blackhole_dex/pkg/blackhole/
- [ ] T003 [P] Create internal agent directory at blackhole_dex/internal/agent/
- [ ] T004 [P] Create internal util directory at blackhole_dex/internal/util/
- [ ] T005 [P] Create configs directory at blackhole_dex/configs/
- [ ] T006 [P] Create tests/integration directory at blackhole_dex/tests/integration/
- [ ] T007 [P] Create tests/testdata directory at blackhole_dex/tests/testdata/
- [ ] T008 Update go.mod to ensure go-ethereum v1.16.7 and testify v1.11.1 dependencies
- [ ] T009 Create .gitignore entries for agent config files and keystore
- [ ] T010 [P] Create example configuration file at blackhole_dex/configs/agent.yaml.example
- [ ] T011 [P] Create contracts configuration file at blackhole_dex/configs/contracts.yaml
- [ ] T012 Create ABI compilation script at blackhole_dex/scripts/compile-abis.sh

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T013 Create Blackhole-specific types file at blackhole_dex/pkg/blackhole/types.go with Position, PoolState, MintParams, SwapParams structs
- [ ] T014 Create logger utility at blackhole_dex/internal/util/logger.go with structured logging functions
- [ ] T015 Create rate limiter utility at blackhole_dex/internal/util/ratelimit.go for RPC call throttling
- [ ] T016 Create configuration loader at blackhole_dex/internal/agent/config.go to parse agent.yaml and contracts.yaml
- [ ] T017 Write table-driven tests for config loader in blackhole_dex/internal/agent/config_test.go
- [ ] T018 Create batch RPC helper utility at blackhole_dex/internal/util/batch_rpc.go for multi-call optimization
- [ ] T019 Write table-driven tests for batch RPC in blackhole_dex/internal/util/batch_rpc_test.go
- [ ] T020 Load contract ABIs for NonfungiblePositionManager, AlgebraPool, RouterV2, ERC20 using existing util.LoadABIFromHardhatArtifact
- [ ] T021 Create singleton manager pattern in blackhole_dex/pkg/blackhole/manager.go with sync.Once initialization and cached contract clients

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Position Monitoring & Zone Detection (Priority: P1) 🎯 MVP

**Goal**: Monitor liquidity positions and detect when they move out of active trading range

**Independent Test**: Deploy test position, simulate price movements, verify agent correctly identifies in-range vs out-of-range status

### Implementation for User Story 1

- [ ] T022 [P] [US1] Create Position entity struct in blackhole_dex/pkg/blackhole/types.go with PositionID, Owner, Pool, Ticks, Liquidity, Status fields
- [ ] T023 [P] [US1] Create PoolState entity struct in blackhole_dex/pkg/blackhole/types.go with PoolAddress, Tick, SqrtPriceX96, ActiveLiquidity fields
- [ ] T024 [P] [US1] Create PositionStatus enum in blackhole_dex/pkg/blackhole/types.go (InRange, BelowRange, AboveRange)
- [ ] T025 [US1] Implement GetPosition function in blackhole_dex/pkg/blackhole/position_manager.go to query single position via NonfungiblePositionManager.positions()
- [ ] T026 [US1] Implement GetPoolState function in blackhole_dex/pkg/blackhole/position_manager.go to query pool via safelyGetStateOfAMM()
- [ ] T027 [US1] Implement CalculatePositionStatus function in blackhole_dex/pkg/blackhole/position_manager.go to determine if position is in-range based on current tick vs tick bounds
- [ ] T028 [US1] Implement BatchGetPositions function in blackhole_dex/pkg/blackhole/position_manager.go using batch RPC calls to query multiple positions efficiently
- [ ] T029 [US1] Write table-driven tests for GetPosition in blackhole_dex/pkg/blackhole/position_manager_test.go using real mainnet transaction hash from README.md (0x9e2247a0...)
- [ ] T030 [US1] Write table-driven tests for CalculatePositionStatus in blackhole_dex/pkg/blackhole/position_manager_test.go covering all three status states
- [ ] T031 [US1] Write table-driven tests for BatchGetPositions in blackhole_dex/pkg/blackhole/position_manager_test.go with mocked RPC batch responses
- [ ] T032 [US1] Create position monitor service at blackhole_dex/internal/agent/monitor.go with polling loop to check position status every 5 minutes
- [ ] T033 [US1] Implement status change detection in monitor service to compare previous vs current status and trigger notifications
- [ ] T034 [US1] Add context.Context support to monitor service for graceful shutdown
- [ ] T035 [US1] Write integration test at blackhole_dex/tests/integration/position_test.go to verify end-to-end position monitoring workflow
- [ ] T036 [US1] Profile GetPosition and BatchGetPositions to ensure <2s and <5s performance targets respectively

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently (can monitor positions and detect range status)

---

## Phase 4: User Story 2 - Manual Liquidity Management (Priority: P2)

**Goal**: Provide manual controls to mint new positions and unstake existing positions

**Independent Test**: Execute mint and unstake operations, verify transactions succeed and balances update correctly

### Implementation for User Story 2

- [ ] T037 [P] [US2] Create MintParams struct in blackhole_dex/pkg/blackhole/types.go matching NonfungiblePositionManager.mint() parameters
- [ ] T038 [P] [US2] Create UnstakeParams struct in blackhole_dex/pkg/blackhole/types.go for decreaseLiquidity and collect parameters
- [ ] T039 [P] [US2] Create MintResult struct in blackhole_dex/pkg/blackhole/types.go to capture position ID, liquidity, amounts, tx hash, gas used
- [ ] T040 [P] [US2] Create UnstakeResult struct in blackhole_dex/pkg/blackhole/types.go to capture withdrawn amounts, fees, tx hashes
- [ ] T041 [US2] Implement ValidateMintParams function in blackhole_dex/pkg/blackhole/liquidity_manager.go to check tick bounds, token ordering, sufficient balance
- [ ] T042 [US2] Implement ApproveTokens function in blackhole_dex/pkg/blackhole/liquidity_manager.go to approve NonfungiblePositionManager for token0 and token1
- [ ] T043 [US2] Implement MintPosition function in blackhole_dex/pkg/blackhole/liquidity_manager.go to execute approve -> mint workflow with transaction tracking
- [ ] T044 [US2] Implement DecreaseL function in blackhole_dex/pkg/blackhole/liquidity_manager.go to call decreaseLiquidity() and wait for confirmation
- [ ] T045 [US2] Implement CollectTokens function in blackhole_dex/pkg/blackhole/liquidity_manager.go to call collect() and retrieve withdrawn tokens + fees
- [ ] T046 [US2] Implement UnstakePosition function in blackhole_dex/pkg/blackhole/liquidity_manager.go to orchestrate decreaseLiquidity -> collect workflow
- [ ] T047 [US2] Add error handling for approval failures with retry logic (max 3 attempts, exponential backoff)
- [ ] T048 [US2] Add slippage validation to ensure actual amounts meet minimum thresholds
- [ ] T049 [US2] Write table-driven tests for ValidateMintParams in blackhole_dex/pkg/blackhole/liquidity_manager_test.go covering all validation rules
- [ ] T050 [US2] Write table-driven tests for MintPosition in blackhole_dex/pkg/blackhole/liquidity_manager_test.go with mocked RPC calls
- [ ] T051 [US2] Write table-driven tests for UnstakePosition in blackhole_dex/pkg/blackhole/liquidity_manager_test.go with mocked RPC calls
- [ ] T052 [US2] Write integration test at blackhole_dex/tests/integration/liquidity_test.go to verify end-to-end mint workflow against testnet
- [ ] T053 [US2] Write integration test at blackhole_dex/tests/integration/liquidity_test.go to verify end-to-end unstake + collect workflow
- [ ] T054 [US2] Profile MintPosition to ensure <30s performance target including approvals
- [ ] T055 [US2] Profile UnstakePosition to ensure <30s performance target for decrease + collect

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently (can monitor positions AND manually manage liquidity)

---

## Phase 5: User Story 3 - Token Swapping (Priority: P3)

**Goal**: Enable token swaps for rebalancing before minting new positions

**Independent Test**: Execute swaps, verify transactions complete and token balances update with correct slippage protection

### Implementation for User Story 3

- [ ] T056 [P] [US3] Create SwapParams struct in blackhole_dex/pkg/blackhole/types.go matching RouterV2.swapExactTokensForTokens() parameters
- [ ] T057 [P] [US3] Create Route struct in blackhole_dex/pkg/blackhole/types.go with pair, from, to, stable, concentrated, receiver fields
- [ ] T058 [P] [US3] Create SwapResult struct in blackhole_dex/pkg/blackhole/types.go to capture amountIn, amountOut, slippageActual, tx hash, gas used
- [ ] T059 [US3] Implement ValidateSwapParams function in blackhole_dex/pkg/blackhole/swap_manager.go to check sufficient balance, valid route, reasonable slippage
- [ ] T060 [US3] Implement CalculateSlippage function in blackhole_dex/pkg/blackhole/swap_manager.go to compute amountOutMin from expected output and slippage percentage
- [ ] T061 [US3] Implement ApproveRouter function in blackhole_dex/pkg/blackhole/swap_manager.go to approve RouterV2 to spend input token
- [ ] T062 [US3] Implement ExecuteSwap function in blackhole_dex/pkg/blackhole/swap_manager.go to execute approve -> swap workflow with transaction tracking
- [ ] T063 [US3] Implement QuoteSwap function in blackhole_dex/pkg/blackhole/swap_manager.go to query expected output without executing transaction (dry-run via eth_call)
- [ ] T064 [US3] Add slippage exceeded error handling to revert if actual output < amountOutMin
- [ ] T065 [US3] Add deadline validation to prevent transaction execution after deadline timestamp
- [ ] T066 [US3] Write table-driven tests for ValidateSwapParams in blackhole_dex/pkg/blackhole/swap_manager_test.go covering balance checks and slippage validation
- [ ] T067 [US3] Write table-driven tests for CalculateSlippage in blackhole_dex/pkg/blackhole/swap_manager_test.go with various slippage percentages (0.1%, 0.5%, 1%, 5%)
- [ ] T068 [US3] Write table-driven tests for ExecuteSwap in blackhole_dex/pkg/blackhole/swap_manager_test.go with mocked RPC calls
- [ ] T069 [US3] Write table-driven tests for QuoteSwap in blackhole_dex/pkg/blackhole/swap_manager_test.go using real mainnet swap transaction hash from README.md (0x1600e68b...)
- [ ] T070 [US3] Write integration test at blackhole_dex/tests/integration/swap_test.go to verify end-to-end WAVAX -> USDC swap workflow
- [ ] T071 [US3] Write integration test at blackhole_dex/tests/integration/swap_test.go to verify slippage protection triggers revert when output too low
- [ ] T072 [US3] Profile ExecuteSwap to ensure <20s performance target including approval

**Checkpoint**: All user stories 1, 2, AND 3 should now be independently functional (monitor + manual liquidity + swaps)

---

## Phase 6: User Story 4 - Automated Repositioning (Priority: P4)

**Goal**: Automatically detect out-of-range positions and reposition to active zone

**Independent Test**: Configure repositioning rules, simulate out-of-range position, verify agent automatically executes unstake -> swap -> mint workflow

### Implementation for User Story 4

- [ ] T073 [P] [US4] Create RepositioningEvent struct in blackhole_dex/pkg/blackhole/types.go to track trigger time, reason, old/new ticks, tx hashes, outcome
- [ ] T074 [P] [US4] Create RepositioningTrigger struct in blackhole_dex/pkg/blackhole/types.go with outOfRangeDuration, tickDistanceThreshold, minFeeOpportunity fields
- [ ] T075 [P] [US4] Create RepositioningPlan struct in blackhole_dex/pkg/blackhole/types.go with current/new tick ranges, swap requirements, gas/fee estimates
- [ ] T076 [US4] Implement EvaluateRepositioning function in blackhole_dex/internal/agent/orchestrator.go to check if position meets trigger conditions
- [ ] T077 [US4] Implement CalculateNewTickRange function in blackhole_dex/internal/agent/orchestrator.go to determine optimal new range based on current pool tick
- [ ] T078 [US4] Implement EstimateGasCost function in blackhole_dex/internal/agent/orchestrator.go to calculate total gas for unstake -> swap -> mint workflow
- [ ] T079 [US4] Implement EstimateFeeOpportunity function in blackhole_dex/internal/agent/orchestrator.go to predict fees in new range based on pool volume
- [ ] T080 [US4] Implement ShouldReposition decision logic in blackhole_dex/internal/agent/orchestrator.go to compare gas cost vs fee opportunity
- [ ] T081 [US4] Implement GenerateRepositioningPlan function in blackhole_dex/internal/agent/orchestrator.go to create full plan with all parameters
- [ ] T082 [US4] Implement ExecuteRepositioning function in blackhole_dex/internal/agent/orchestrator.go to orchestrate multi-step workflow (unstake -> collect -> swap if needed -> mint)
- [ ] T083 [US4] Add state tracking to handle partial failures (e.g., unstake succeeds but mint fails)
- [ ] T084 [US4] Add price revalidation before each step to abort if conditions change mid-execution
- [ ] T085 [US4] Add gas price check to defer repositioning if gas > max configured limit
- [ ] T086 [US4] Implement event logging to write RepositioningEvent to jsonl file at ~/.blackhole-agent/events.jsonl
- [ ] T087 [US4] Write table-driven tests for EvaluateRepositioning in blackhole_dex/internal/agent/orchestrator_test.go covering all trigger conditions
- [ ] T088 [US4] Write table-driven tests for CalculateNewTickRange in blackhole_dex/internal/agent/orchestrator_test.go to verify tick spacing alignment (multiples of 200)
- [ ] T089 [US4] Write table-driven tests for ShouldReposition in blackhole_dex/internal/agent/orchestrator_test.go covering gas vs fees decision logic
- [ ] T090 [US4] Write table-driven tests for ExecuteRepositioning in blackhole_dex/internal/agent/orchestrator_test.go with mocked position/swap/liquidity managers
- [ ] T091 [US4] Write integration test at blackhole_dex/tests/integration/repositioning_test.go to verify end-to-end automated repositioning workflow
- [ ] T092 [US4] Write integration test for partial failure recovery (unstake succeeds, mint fails scenario)
- [ ] T093 [US4] Profile ExecuteRepositioning to ensure <2min performance target for full workflow
- [ ] T094 [US4] Add multi-position handling logic to process largest positions first (from config.multi_position_strategy)

**Checkpoint**: All four user stories should now be independently functional - agent can monitor, manage manually, swap, and auto-reposition

---

## Phase 7: CLI Interface & Main Entry Point

**Purpose**: Build user-facing command-line interface for all operations

- [ ] T095 [P] Create main CLI entry point at blackhole_dex/cmd/agent/main.go with cobra/viper for command parsing
- [ ] T096 [P] Implement "start" command to launch agent daemon with config file path flag
- [ ] T097 [P] Implement "stop" command to gracefully shutdown running daemon
- [ ] T098 [P] Implement "status" command to show current monitoring state and position statuses
- [ ] T099 [P] Implement "positions list" command to display all monitored positions with status
- [ ] T100 [P] Implement "positions mint" command with flags for token0, token1, amounts, tick-lower, tick-upper, slippage
- [ ] T101 [P] Implement "positions unstake" command with flags for position-id, collect-fees
- [ ] T102 [P] Implement "positions evaluate" command to show dry-run repositioning plan for given position
- [ ] T103 [P] Implement "positions reposition" command to manually trigger repositioning for specific position
- [ ] T104 [P] Implement "swap" command with flags for token-in, token-out, amount-in, slippage
- [ ] T105 [P] Implement "history" command to display repositioning events from events.jsonl log
- [ ] T106 [P] Implement "wallet balance" command to show token balances
- [ ] T107 [P] Implement "validate-config" command to check config file validity
- [ ] T108 Add colored output for status indicators (green=in-range, yellow=out-of-range, red=failed)
- [ ] T109 Add progress bars for multi-step operations (mint, unstake, reposition workflows)
- [ ] T110 Add --dry-run flag to all state-changing commands for simulation mode
- [ ] T111 Add --daemon flag to start command for background execution
- [ ] T112 Write integration test for CLI commands at blackhole_dex/cmd/agent/main_test.go

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T113 [P] Add Prometheus metrics exporter at blackhole_dex/internal/util/metrics.go for monitoring (positions tracked, repositionings executed, gas spent)
- [ ] T114 [P] Add Slack notification support at blackhole_dex/internal/util/notifications.go with webhook integration
- [ ] T115 [P] Add email notification support using SMTP
- [ ] T116 Implement notification triggers for repositioning start, complete, failure events
- [ ] T117 [P] Create comprehensive README.md at blackhole_dex/README.md with setup instructions, examples, architecture diagram
- [ ] T118 [P] Create CONTRIBUTING.md with development guidelines and testing requirements
- [ ] T119 Add CI/CD workflow file at .github/workflows/agent-ci.yml for automated testing
- [ ] T120 Create Dockerfile for containerized agent deployment at blackhole_dex/Dockerfile
- [ ] T121 Create docker-compose.yml for local testing environment
- [ ] T122 Add systemd service file at blackhole_dex/scripts/blackhole-agent.service for daemon management
- [ ] T123 Implement log rotation for events.jsonl to prevent unbounded growth
- [ ] T124 Add performance profiling mode with --profile flag to collect pprof data
- [ ] T125 Run security audit on private key handling and ensure no keys logged
- [ ] T126 Add rate limiting tests to verify RPC call throttling works correctly
- [ ] T127 Add end-to-end smoke test script at blackhole_dex/scripts/smoke-test.sh
- [ ] T128 Profile entire application to identify and optimize any operations >200ms
- [ ] T129 Update quickstart.md with real CLI examples from implemented commands
- [ ] T130 Create troubleshooting guide in docs/troubleshooting.md with common error solutions

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-6)**: All depend on Foundational phase completion
  - User Story 1 (P1): Can start after Foundational - No dependencies on other stories
  - User Story 2 (P2): Can start after Foundational - No dependencies on other stories
  - User Story 3 (P3): Can start after Foundational - No dependencies on other stories
  - User Story 4 (P4): Depends on US1 (monitoring), US2 (liquidity mgmt), US3 (swaps) being complete
- **CLI Interface (Phase 7)**: Depends on all user stories being functionally complete
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

```
Foundational (Phase 2)
        ↓
   ┌────┴────┬────────┬────────┐
   ▼         ▼        ▼        ▼
  US1       US2      US3    (wait)
  (P1)      (P2)     (P3)      │
   │         │        │        │
   └─────────┴────────┴────────┘
              ↓
             US4
             (P4)
```

- **User Story 1 (P1)**: Independent - provides position monitoring
- **User Story 2 (P2)**: Independent - provides manual liquidity management
- **User Story 3 (P3)**: Independent - provides token swapping
- **User Story 4 (P4)**: Depends on US1, US2, US3 - combines all capabilities for automation

### Within Each User Story

- Tests (if using TDD) MUST be written and FAIL before implementation
- Types/structs before functions
- Validation functions before operations
- Core operations before integration
- Unit tests before integration tests
- Profile performance after functional completion
- Story complete before moving to next priority

### Parallel Opportunities

**Setup Phase (Phase 1):**
```bash
# Launch all directory creation and config file tasks together
T002, T003, T004, T005, T006, T007, T010, T011  # Can all run in parallel
```

**Foundational Phase (Phase 2):**
```bash
# After types are created (T013), these can run in parallel:
T014, T015, T016, T018, T020  # Different files, no dependencies
```

**User Story 1 (Phase 3):**
```bash
# Create all structs in parallel:
T022, T023, T024  # All define types in types.go (combine into single edit)

# After PositionManager functions exist, run tests in parallel:
T029, T030, T031  # Different test files for different functions
```

**User Story 2 (Phase 4):**
```bash
# Create all structs in parallel:
T037, T038, T039, T040  # All define types

# Run all tests in parallel after implementation:
T049, T050, T051, T052, T053  # Different test scenarios
```

**User Story 3 (Phase 5):**
```bash
# Create all structs in parallel:
T056, T057, T058  # All define types

# Run all tests in parallel:
T066, T067, T068, T069, T070, T071  # Different test scenarios
```

**User Story 4 (Phase 6):**
```bash
# Create all structs in parallel:
T073, T074, T075  # All define types

# Run all tests in parallel after implementation:
T087, T088, T089, T090, T091, T092  # Different test scenarios
```

**CLI Phase (Phase 7):**
```bash
# All command implementations can run in parallel:
T095, T096, T097, T098, T099, T100, T101, T102, T103, T104, T105, T106, T107
```

**Polish Phase (Phase 8):**
```bash
# Documentation and deployment files can run in parallel:
T113, T114, T115, T117, T118, T119, T120, T121, T122
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T012)
2. Complete Phase 2: Foundational (T013-T021) **← CRITICAL - blocks all stories**
3. Complete Phase 3: User Story 1 (T022-T036)
4. **STOP and VALIDATE**: Test position monitoring independently
5. Build minimal CLI for "positions list" command
6. Deploy/demo position monitoring capability (MVP!)

**MVP Deliverable**: Agent that monitors positions and reports in-range vs out-of-range status

### Incremental Delivery

1. **Foundation** (T001-T021) → All user stories can now start
2. **Add User Story 1** (T022-T036) → Can monitor positions → Deploy/Demo (MVP!)
3. **Add User Story 2** (T037-T055) → Can manually mint/unstake → Deploy/Demo
4. **Add User Story 3** (T056-T072) → Can swap for rebalancing → Deploy/Demo
5. **Add User Story 4** (T073-T094) → Full automation → Deploy/Demo
6. **Add CLI** (T095-T112) → User-friendly interface → Deploy/Demo
7. **Polish** (T113-T130) → Production-ready → Final Release

Each increment adds value without breaking previous capabilities.

### Parallel Team Strategy

With multiple developers:

1. **Team completes Setup + Foundational together** (T001-T021)
2. **Once Foundational is done, split into parallel workstreams:**
   - **Developer A**: User Story 1 (T022-T036) - Position monitoring
   - **Developer B**: User Story 2 (T037-T055) - Liquidity management
   - **Developer C**: User Story 3 (T056-T072) - Token swapping
3. **After US1, US2, US3 complete:**
   - **Developer A**: User Story 4 (T073-T094) - Automated repositioning
   - **Developer B**: CLI Interface (T095-T112)
   - **Developer C**: Polish & Documentation (T113-T130)
4. Stories integrate seamlessly because they're independently designed

---

## Testing Strategy

### Table-Driven Tests (Constitution Requirement)

All business logic MUST use table-driven test pattern:

```go
func TestCalculatePositionStatus(t *testing.T) {
    tests := []struct {
        name          string
        currentTick   int24
        tickLower     int24
        tickUpper     int24
        expectedStatus PositionStatus
    }{
        {"In range", -249000, -250000, -248000, InRange},
        {"Below range", -251000, -250000, -248000, BelowRange},
        {"Above range", -247000, -250000, -248000, AboveRange},
        {"At lower boundary", -250000, -250000, -248000, InRange},
        {"At upper boundary", -248000, -250000, -248000, AboveRange},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            status := CalculatePositionStatus(tt.currentTick, tt.tickLower, tt.tickUpper)
            assert.Equal(t, tt.expectedStatus, status)
        })
    }
}
```

### Real Mainnet Transaction Validation

Use real transaction hashes from README.md in tests:

- **Mint**: `0x9e2247a0210448cab301475eef741eba0ee9a9351188a92b8127fce27206b9d0`
- **Swap**: `0x1600e68bfd607a5e8452f7533b162eeb4afd4f0435f31639999aa46fbaef79b1`
- **Approve**: `0x17226fdd0f0df51d1fdd7a47a90de291766f4858a688cdc6c91833b9208bb13f`

### Integration Test Pattern

```go
func TestEndToEndPositionMonitoring(t *testing.T) {
    // Setup: Load config, connect to testnet RPC
    cfg := loadTestConfig()
    client := setupTestClient(cfg)

    // Execute: Query position and check status
    position, err := GetPosition(context.Background(), client, testPositionID)
    require.NoError(t, err)

    // Verify: Check expected fields
    assert.Equal(t, expectedTickLower, position.TickLower)
    assert.NotZero(t, position.Liquidity)
    assert.Contains(t, []PositionStatus{InRange, BelowRange, AboveRange}, position.Status)
}
```

### Performance Profiling

After each user story implementation, profile key operations:

```bash
# Profile position monitoring (US1)
go test -bench=BenchmarkGetPosition -benchmem -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Profile liquidity operations (US2)
go test -bench=BenchmarkMintPosition -benchmem

# Profile swap operations (US3)
go test -bench=BenchmarkExecuteSwap -benchmem

# Profile full repositioning (US4)
go test -bench=BenchmarkExecuteRepositioning -benchmem
```

Target: All operations <200ms in hot paths

---

## Notes

- **[P] tasks**: Different files, can run in parallel
- **[Story] labels**: Map task to specific user story for traceability
- **Each user story is independently completable and testable**
- **Commit after each task or logical group**
- **Stop at any checkpoint to validate story independently**
- **Use real mainnet tx hashes from README.md in decoder tests**
- **All tests must pass before commits (Constitution requirement)**
- **Profile performance after each story to catch >200ms operations**

**Constitution Compliance Checkpoints:**
- After each task: Verify go-ethereum types used, errors wrapped with context
- After each function: Verify naming convention (Decode*, Build*Data, etc.)
- After each test: Verify table-driven pattern used
- After each integration test: Verify context.Context used for cancellation
- After each profile: Verify <200ms hot path performance

**Avoid:**
- Vague tasks without file paths
- Tasks touching same file simultaneously (breaks parallelism)
- Cross-story dependencies that break independence
- Premature optimization before profiling
- Test failures blocking other parallel work

---

## Task Count Summary

- **Total Tasks**: 130
- **Phase 1 (Setup)**: 12 tasks
- **Phase 2 (Foundational)**: 9 tasks (BLOCKS all stories)
- **Phase 3 (US1 - Position Monitoring)**: 15 tasks
- **Phase 4 (US2 - Liquidity Management)**: 19 tasks
- **Phase 5 (US3 - Token Swapping)**: 17 tasks
- **Phase 6 (US4 - Automated Repositioning)**: 22 tasks
- **Phase 7 (CLI Interface)**: 18 tasks
- **Phase 8 (Polish)**: 18 tasks

**Parallel Opportunities**: 45+ tasks can run in parallel within phases

**MVP Scope**: Phase 1 + Phase 2 + Phase 3 = 36 tasks (position monitoring only)

**Full Feature**: All 130 tasks for complete automated repositioning agent
