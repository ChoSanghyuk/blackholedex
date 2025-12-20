# Implementation Plan: Unstake Liquidity from Blackhole DEX

**Branch**: `001-unstake` | **Date**: 2025-12-19 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-unstake/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement unstake functionality for liquidity positions in Blackhole DEX by calling the FarmingCenter contract's multicall function. Users can withdraw their staked LP tokens and claim accumulated rewards through the exitFarming operation. This feature complements the existing Mint/Stake operations by providing a safe exit mechanism for liquidity providers.

## Technical Context

**Language/Version**: Go 1.24.10
**Primary Dependencies**:
- github.com/ethereum/go-ethereum v1.16.7 (Ethereum client library)
- Existing internal packages: internal/util, pkg/contractclient, pkg/types
- FarmingCenter contract (Algebra Integral v1.2.2 farming)

**Storage**: N/A (blockchain state only, no local persistence)
**Testing**: Go standard testing (go test), mainnet transaction validation
**Target Platform**: Server/CLI (Linux/macOS/Windows)
**Project Type**: Single project (Go library + CLI)
**Performance Goals**: Single transaction completion within 30 seconds, gas cost < 150% of direct contract call
**Constraints**:
- Must interact only with FarmingCenter.multicall(bytes[] data) function
- Must validate staked balance before unstake attempts
- Must handle reward collection during exitFarming
- NEEDS CLARIFICATION: FarmingCenter contract address on Avalanche mainnet
- NEEDS CLARIFICATION: IncentiveKey structure and how to obtain it for exitFarming

**Scale/Scope**: Single-pool (WAVAX/USDC) operations, ~5-10 new methods in blackhole.go

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

This feature MUST comply with all principles in `.specify/memory/constitution.md`. Check each principle:

### Principle 1: Pool-Specific Scope
- [x] Feature operates ONLY on WAVAX/USDC pool
  - Unstake operations target only NFT positions from WAVAX/USDC pool
  - Token validation against constitutional addresses (WAVAX: 0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7, USDC: 0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E)
- [x] No support for other tokens or pools introduced
  - FarmingCenter interaction scoped to WAVAX/USDC pair only
- [x] Contract addresses match constitution technical constraints
  - Will validate NFT positions against NonfungiblePositionManager: 0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146

### Principle 2: Autonomous Rebalancing
- [x] Monitoring capability included (if relevant)
  - N/A - Unstake is a manual operation triggered by operator, not autonomous
- [x] Rebalancing logic maintains atomicity/rollback safety
  - N/A - Unstake does not perform rebalancing, it exits positions
- [x] Decision logic documented and testable
  - N/A - Manual operation, no autonomous decision logic

### Principle 3: Financial Transparency
- [x] Gas tracking implemented for all transactions
  - Will track gas for approval (if needed), exitFarming, and collectRewards transactions
  - Use existing TransactionRecord structure and ExtractGasCost utility
- [x] Swap fees calculated and recorded
  - N/A - Unstake does not involve swaps
- [x] Incentives/rewards tracked
  - Will track rewards collected during exitFarming via collectRewards function
  - Return reward amounts in UnstakeResult
- [x] Profit/loss calculations accurate
  - Will provide reward amounts; P/L calculation remains responsibility of calling code
- [x] Exportable reporting available
  - Transaction records returned in structured UnstakeResult for export

### Principle 4: Gas Optimization
- [x] Rebalancing thresholds configured to prevent excessive transactions
  - N/A - Manual operation, not autonomous rebalancing
- [x] Gas estimation used before transaction submission
  - Will use automatic gas limit estimation via contractclient (nil gas limit parameter)
- [x] Contract calls optimized (batch operations, minimal storage)
  - Will use FarmingCenter.multicall to batch exitFarming operations if needed
  - Minimize separate transactions by using multicall pattern
- [x] Approval reuse implemented where safe
  - N/A - Unstake typically doesn't require token approvals (FarmingCenter already has NFT custody)

### Principle 5: Fail-Safe Operation
- [x] All external calls have timeout/retry logic
  - Will use TxListener.WaitForTransaction with existing timeout logic
- [x] Transaction failures logged with full context
  - Will capture transaction hash, error message, and context in UnstakeResult
- [x] Partial failures trigger rollback or safe termination
  - FarmingCenter.exitFarming is atomic; partial state changes revert on-chain
  - Will return error and transaction records for any failed operations
- [x] Slippage protection enforced
  - N/A - Unstake does not involve swaps or liquidity calculations with slippage
- [x] Circuit breaker pattern implemented
  - N/A - Manual operation; operator controls execution
- [x] Manual recovery documented
  - Will document failure scenarios and recovery procedures in quickstart.md

**Violations**: None. All constitutional principles are satisfied or not applicable (N/A) to this manual unstake operation.

## Project Structure

### Documentation (this feature)

```text
specs/001-unstake/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output - FarmingCenter contract details, IncentiveKey structure
├── data-model.md        # Phase 1 output - UnstakeResult, FarmingParams entities
├── quickstart.md        # Phase 1 output - Step-by-step unstake guide
├── contracts/           # Phase 1 output - FarmingCenter ABI and interface definitions
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
blackhole_dex/
├── blackhole.go                # Add Unstake() method
├── types.go                    # Add UnstakeResult, IncentiveKey, FarmingParams types
├── blackhole_interfaces.go     # Extend ContractClient interface if needed
├── internal/
│   └── util/
│       └── validation.go       # Add unstake-specific validation utilities
├── pkg/
│   └── contractclient/
│       └── client.go           # Use existing Call/Send methods for FarmingCenter
└── blackholedex-contracts/
    └── artifacts/
        └── contracts/
            └── FarmingCenter.sol/
                └── FarmingCenter.json  # ABI for multicall interaction

tests/
└── blackhole_test.go           # Add Unstake() integration tests
```

**Structure Decision**: Single project structure. This feature adds methods to existing `blackhole.go` for FarmingCenter interaction, new types in `types.go`, and leverages existing `pkg/contractclient` infrastructure. No new major directories needed - only extensions to current codebase.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations - section not applicable.

---

## Post-Design Constitution Review

**Date**: 2025-12-19
**Status**: APPROVED - All principles satisfied

After completing Phase 0 (Research) and Phase 1 (Design), re-evaluating constitutional compliance:

### Principle 1: Pool-Specific Scope ✅
- **Design**: IncentiveKey.Pool validated against WAVAX/USDC pool constant in Unstake() function
- **Implementation**: quickstart.md demonstrates validation: `incentiveKey.Pool.Hex() != wavaxUsdcPair` check
- **Contracts**: unstake-api.md specifies rejection of non-WAVAX/USDC pools with error message
- **Status**: COMPLIANT - No violations introduced during design

### Principle 2: Autonomous Rebalancing ✅
- **Design**: Unstake is a manual operation, no autonomous logic
- **Status**: N/A - Principle not applicable to manual unstake operations

### Principle 3: Financial Transparency ✅
- **Design**:
  - TransactionRecord tracks all gas costs (data-model.md entity #5)
  - RewardAmounts tracks collected incentives (data-model.md entity #4)
  - UnstakeResult provides complete financial picture (data-model.md entity #6)
- **Implementation**: quickstart.md Step 8 shows gas extraction via util.ExtractGasCost()
- **Contracts**: unstake-api.md Returns section specifies TotalGasCost and Rewards fields
- **Status**: COMPLIANT - Full gas tracking and reward reporting implemented

### Principle 4: Gas Optimization ✅
- **Design**:
  - Multicall batches exitFarming + collectRewards in single transaction (research.md Task 3)
  - Pre-flight validation prevents wasted gas on doomed transactions (quickstart.md Steps 3-4)
  - Automatic gas estimation used (no manual calculation)
- **Implementation**: quickstart.md Step 6 shows `nil` gas limit parameter for auto-estimation
- **Contracts**: unstake-api.md specifies multicall batching pattern
- **Status**: COMPLIANT - Gas optimizations implemented per research decisions

### Principle 5: Fail-Safe Operation ✅
- **Design**:
  - Pre-flight ownership and farming status checks (data-model.md FarmingStatus entity)
  - Comprehensive error handling with UnstakeResult.ErrorMessage (data-model.md entity #6)
  - Transaction confirmation via TxListener.WaitForTransaction
  - FarmingCenter.exitFarming is atomic (on-chain revert on failure)
- **Implementation**: quickstart.md Steps 3-4 show ownership/farming verification before transaction
- **Contracts**: unstake-api.md Error Conditions table lists all failure scenarios
- **Troubleshooting**: quickstart.md includes troubleshooting guide for common failures
- **Status**: COMPLIANT - Comprehensive fail-safe patterns implemented

### Summary

**No constitutional violations introduced during design phase.**

All design artifacts (research.md, data-model.md, contracts/unstake-api.md, quickstart.md) explicitly enforce constitutional constraints:
- Pool scope validation in function signature
- Financial transparency via structured result types
- Gas optimization via multicall pattern
- Fail-safe operation via comprehensive validation and error handling

**Approval**: Ready to proceed with /speckit.tasks to generate implementation tasks.
