# Implementation Plan: Automated Liquidity Repositioning Strategy

**Branch**: `001-liquidity-repositioning` | **Date**: 2025-12-23 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-liquidity-repositioning/spec.md`

**User Requirement**: The RunStrategy1 method should take a string channel as a parameter, and use it to report errors, profits, and costs (gas) continuously.

## Summary

Implement RunStrategy1 method on the Blackhole struct to autonomously manage concentrated liquidity positions in the WAVAX/USDC pool. The strategy continuously monitors pool price, automatically rebalances positions when price exits active range, and re-enters only after price stabilization. All operations include comprehensive financial tracking (gas costs, fees, incentives) with continuous reporting via a string channel parameter. The method orchestrates existing Swap, Mint, Stake, Unstake, and Withdraw methods to maintain capital efficiency while maximizing incentive earnings.

## Technical Context

**Language/Version**: Go 1.24.10
**Primary Dependencies**: github.com/ethereum/go-ethereum v1.16.7, existing internal packages (util, contractclient)
**Storage**: N/A (blockchain state only, no local persistence)
**Testing**: Go standard testing framework (testing package), table-driven tests with mock RPC responses
**Target Platform**: Linux/macOS server, Avalanche C-Chain mainnet
**Project Type**: Single (Go application, no separate frontend/backend)
**Performance Goals**:
  - Detect out-of-range conditions within 2 monitoring intervals (120 seconds default)
  - Complete rebalancing cycle in under 10 minutes during stable network conditions
  - Monitor pool state at 1-minute intervals minimum
**Constraints**:
  - Gas costs per rebalancing must not exceed 2% of position value
  - Token ratio deviation from 50:50 must be less than 1%
  - RPC call rate limits (configurable, default conservative)
  - 24-hour continuous operation without intervention
**Scale/Scope**:
  - Single pool (WAVAX/USDC) operation
  - Single strategy instance per wallet
  - Configurable position size (user-determined capital allocation)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

This feature MUST comply with all principles in `.specify/memory/constitution.md`. Check each principle:

### Principle 1: Pool-Specific Scope
- [x] Feature operates ONLY on WAVAX/USDC pool (0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0)
- [x] No support for other tokens or pools introduced (hardcoded to WAVAX/USDC)
- [x] Contract addresses match constitution technical constraints (RouterV2, NonfungiblePositionManager, Gauge per constitution)
- [x] All swap routes limited to WAVAX↔USDC pairs only

**Status**: ✅ PASS

### Principle 2: Autonomous Rebalancing
- [x] Monitoring capability included (continuous price monitoring loop)
- [x] Rebalancing logic maintains atomicity/rollback safety (Withdraw uses multicall for atomic operations)
- [x] Decision logic documented and testable (price out-of-range detection, stability window logic)
- [x] Manual intervention not required (fully autonomous until explicitly stopped)

**Status**: ✅ PASS

### Principle 3: Financial Transparency
- [x] Gas tracking implemented for all transactions (TransactionRecord tracking in all operations)
- [x] Swap fees calculated and recorded (swap operations track gas costs)
- [x] Incentives/rewards tracked (Unstake collects and records rewards)
- [x] Profit/loss calculations accurate (net P&L = incentives - gas - swap fees)
- [x] Exportable reporting available (continuous reporting via string channel parameter)

**Status**: ✅ PASS - Enhanced by user requirement for continuous channel reporting

### Principle 4: Gas Optimization
- [x] Rebalancing thresholds configured to prevent excessive transactions (out-of-range detection prevents rebalancing for minor price moves)
- [x] Gas estimation used before transaction submission (existing ContractClient.Send uses automatic gas estimation)
- [x] Contract calls optimized (Withdraw uses multicall batching, approval reuse in ensureApproval)
- [x] Approval reuse implemented where safe (ensureApproval checks existing allowances)
- [x] Price stability detection prevents wasteful rebalancing during volatility (0.5% threshold over 5 intervals)

**Status**: ✅ PASS

### Principle 5: Fail-Safe Operation
- [x] All external calls have timeout/retry logic (existing ContractClient handles RPC timeouts)
- [x] Transaction failures logged with full context (TransactionRecord tracking, TxListener.WaitForTransaction)
- [x] Partial failures trigger rollback or safe termination (multicall atomicity, explicit error handling)
- [x] Slippage protection enforced (Mint uses slippagePct parameter, Swap uses amountOutMin)
- [x] Circuit breaker pattern implemented (error detection with halt-on-failure behavior)
- [x] Manual recovery documented (error reporting via channel enables operator intervention)

**Status**: ✅ PASS

**Violations**: None

**Overall Gate Status**: ✅ APPROVED - All 5 constitutional principles satisfied

## Project Structure

### Documentation (this feature)

```text
specs/001-liquidity-repositioning/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── strategy_api.go  # Go interface/type definitions for strategy
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
# Single Go project structure (matches existing codebase)

blackhole_dex/
├── blackhole.go                    # Blackhole struct with existing methods (Swap, Mint, Stake, Unstake, Withdraw)
├── blackhole.go                    # NEW: RunStrategy1 method to be added here
├── types.go                        # Existing types (TransactionRecord, StakingResult, etc.)
├── blackhole_interfaces.go         # Existing interfaces (ContractClient, TxListener)
│
├── internal/
│   └── util/                       # Existing utilities
│       ├── calculations.go         # ComputeAmounts, CalculateTickBounds, etc.
│       └── validation.go           # ValidateStakingRequest, etc.
│
├── pkg/
│   ├── contractclient/             # Generic contract client (existing)
│   └── types/                      # Transaction types (existing)
│
├── cmd/
│   └── main.go                     # CLI entry point
│
└── tests/
    ├── unit/
    │   └── strategy_test.go        # NEW: Unit tests for RunStrategy1
    └── integration/
        └── strategy_integration_test.go  # NEW: Integration tests with mock RPC
```

**Structure Decision**: Single Go project (Option 1) matches existing blackholego codebase. RunStrategy1 is added as a method to the existing Blackhole struct in blackhole.go, leveraging existing methods (Swap, Mint, Stake, Unstake, Withdraw) and utilities (util package). No new top-level packages required; all strategy logic lives within the Blackhole type as an orchestration method.

## Complexity Tracking

> **Not Required**: No constitutional violations detected. All principles satisfied by design.
