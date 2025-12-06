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

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately

- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories

  

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
