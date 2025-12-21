# Implementation Plan: Liquidity Position Withdrawal

**Branch**: `003-liquidity-withdraw` | **Date**: 2025-12-21 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-liquidity-withdraw/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

This feature implements a `Withdraw` function to fully exit liquidity positions created by the `Mint` function. The withdrawal process uses a multicall transaction to the nonfungiblePositionManager contract containing three operations: decreaseLiquidity (remove all liquidity), collect (retrieve tokens and fees), and burn (destroy the NFT). The function accepts an NFT token ID as parameter, validates ownership, tracks all gas costs, and returns comprehensive results with error handling.

## Technical Context

**Language/Version**: Go 1.24.10
**Primary Dependencies**: github.com/ethereum/go-ethereum v1.16.7, existing internal packages (util, contractclient)
**Storage**: N/A (blockchain state only, no local persistence)
**Testing**: Go testing package, existing test infrastructure in pkg/contractclient
**Target Platform**: Linux/macOS server (Avalanche C-Chain interaction)
**Project Type**: Single project (Go library/CLI)
**Performance Goals**: Transaction completion within 2 minutes under normal network conditions, gas estimation accuracy >95%
**Constraints**: Must operate within Avalanche C-Chain gas limits, slippage protection required, all operations must be atomic or safely rollback-able
**Scale/Scope**: Single user operation (wallet-level), handles NFT positions from existing Mint function, integrates with existing Blackhole struct and ContractClient map

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

This feature MUST comply with all principles in `.specify/memory/constitution.md`. Check each principle:

### Principle 1: Pool-Specific Scope
- [x] Feature operates ONLY on WAVAX/USDC pool - Withdraw function operates on NFT positions created by Mint for WAVAX/USDC pool
- [x] No support for other tokens or pools introduced - Function only interacts with nonfungiblePositionManager for existing WAVAX/USDC positions
- [x] Contract addresses match constitution technical constraints - Uses existing nonfungiblePositionManager (0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146)

### Principle 2: Autonomous Rebalancing
- [x] Monitoring capability included (if relevant) - N/A for withdrawal (user-initiated operation, not autonomous)
- [x] Rebalancing logic maintains atomicity/rollback safety - Multicall ensures atomic execution of decreaseLiquidity, collect, and burn
- [x] Decision logic documented and testable - Withdrawal is explicit user action with clear validation and execution steps

### Principle 3: Financial Transparency
- [x] Gas tracking implemented for all transactions - TransactionRecord tracks gas used, gas price, and total cost
- [x] Swap fees calculated and recorded - N/A for withdrawal (no swaps involved, only liquidity removal and fee collection)
- [x] Incentives/rewards tracked - Collected fees are part of withdrawal, tracked in result structure
- [x] Profit/loss calculations accurate - Gas costs tracked, token amounts returned are recorded
- [x] Exportable reporting available - WithdrawalResult contains all transaction records with structured data

### Principle 4: Gas Optimization
- [x] Rebalancing thresholds configured to prevent excessive transactions - N/A (user-initiated, not automated rebalancing)
- [x] Gas estimation used before transaction submission - Uses automatic gas limit estimation via ContractClient
- [x] Contract calls optimized (batch operations, minimal storage) - Multicall batches decreaseLiquidity, collect, and burn into single transaction
- [x] Approval reuse implemented where safe - N/A for withdrawal (no token approvals needed, NFT already owned)

### Principle 5: Fail-Safe Operation
- [x] All external calls have timeout/retry logic - TxListener.WaitForTransaction handles transaction confirmation with timeout
- [x] Transaction failures logged with full context - All errors include context (tx hash if available, operation type, error details)
- [x] Partial failures trigger rollback or safe termination - Multicall ensures atomicity; if any operation fails, entire transaction reverts
- [x] Slippage protection enforced - Minimum token amounts (amount0Min, amount1Min) calculated for decreaseLiquidity
- [x] Circuit breaker pattern implemented - N/A for single-operation function (validation prevents execution if preconditions fail)
- [x] Manual recovery documented - Function returns detailed error messages and success status for manual intervention

**Violations**: None - all constitutional principles are satisfied

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
blackhole_dex/
├── blackhole.go                    # Add Withdraw method to Blackhole struct
├── blackhole_interfaces.go         # Existing interfaces (no changes)
├── types.go                        # Add WithdrawResult and related types
├── internal/
│   └── util/                       # Utility functions for calculations
│       ├── validation.go           # Input validation helpers
│       └── gas.go                  # Gas cost extraction
├── pkg/
│   └── contractclient/             # Generic contract interaction (existing)
│       ├── contractclient.go
│       └── contractclient_test.go  # Add withdrawal integration tests
└── blackholedex-contracts/
    └── artifacts/                  # ABI files (reference only)
        └── contracts/
            └── periphery/
                └── NonfungiblePositionManager.sol/
                    └── NonfungiblePositionManager.json
```

**Structure Decision**: Single project structure. The Withdraw function will be added as a method on the existing `Blackhole` struct in `blackhole.go`, following the same pattern as Mint, Stake, and Unstake. A new `WithdrawResult` type will be added to `types.go` to track withdrawal outcomes, transaction details, and gas costs. Integration tests will be added to `pkg/contractclient/contractclient_test.go` to verify multicall execution and NFT burn verification.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
