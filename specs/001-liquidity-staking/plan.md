# Implementation Plan: Liquidity Staking

**Branch**: `001-liquidity-staking` | **Date**: 2025-12-09 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-liquidity-staking/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Complete the existing `Mint` method in `blackhole.go` to enable single-step liquidity staking in the WAVAX-USDC concentrated liquidity pool on Blackhole DEX. The system accepts maximum token amounts and a configurable range width parameter, automatically calculates optimal tick bounds and token amounts using existing AMM utility functions, executes token approvals, and mints a liquidity position NFT with comprehensive financial tracking and fail-safe error handling.

**Technical Approach**: Leverage existing `ComputeAmounts` utility in `internal/util/amm.go` for optimal amount calculation, use the established `ContractClient` pattern for ERC20 approvals and NonfungiblePositionManager minting, and implement slippage protection through minimum amount calculations. The implementation extends the partial `Mint` method already present in `blackhole.go`, completing the approval workflow and mint transaction submission with comprehensive logging for constitutional compliance.

## Technical Context

**Language/Version**: Go 1.24.10
**Primary Dependencies**: github.com/ethereum/go-ethereum v1.16.7, existing internal packages (util, contractclient)
**Storage**: N/A (blockchain state only, no local persistence)
**Testing**: Go testing framework (existing pattern with testify)
**Target Platform**: Avalanche C-Chain (EVM-compatible blockchain)
**Project Type**: Single project (blockchain client library)
**Performance Goals**: Complete staking operation (2 approvals + 1 mint) in under 5 minutes under normal network conditions
**Constraints**:
- Must complete within 20-minute transaction deadline
- Slippage tolerance configurable (default 5%)
- Gas costs must be minimized through approval reuse
- All operations must be fail-safe with rollback capability
**Scale/Scope**: Single-operator liquidity management system, handles 2-3 transactions per staking operation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

This feature MUST comply with all principles in `.specify/memory/constitution.md`. Check each principle:

### Principle 1: Pool-Specific Scope
- [x] Feature operates ONLY on WAVAX/USDC pool
  - Hardcoded addresses: WAVAX (0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7), USDC (0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E)
  - Pool query hardcoded to wavaxUsdcPair (0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0)
- [x] No support for other tokens or pools introduced
  - Mint method signature restricts to token0/token1 addresses that must match WAVAX/USDC
- [x] Contract addresses match constitution technical constraints
  - NonfungiblePositionManager: requires contract address (not yet in constants, needs addition)
  - Deployer: 0x5d433a94a4a2aa8f9aa34d8d15692dc2e9960584 (already present)

### Principle 2: Autonomous Rebalancing
- [x] Monitoring capability included (if relevant)
  - N/A for this feature (foundation for future rebalancing)
- [x] Rebalancing logic maintains atomicity/rollback safety
  - N/A for this feature (initial stake only)
- [x] Decision logic documented and testable
  - Tick calculation logic uses existing ComputeAmounts utility (well-tested)

### Principle 3: Financial Transparency
- [x] Gas tracking implemented for all transactions
  - Must add: log gas costs for approval txs and mint tx
- [x] Swap fees calculated and recorded
  - N/A for this feature (no swaps in initial stake)
- [x] Incentives/rewards tracked
  - N/A for this feature (staking only, rewards collection is future work)
- [x] Profit/loss calculations accurate
  - N/A for this feature (initial capital deployment)
- [x] Exportable reporting available
  - Must implement: structured logging of transaction hashes, gas costs, amounts

### Principle 4: Gas Optimization
- [x] Rebalancing thresholds configured to prevent excessive transactions
  - N/A for this feature (single stake operation)
- [x] Gas estimation used before transaction submission
  - Existing ContractClient.Send() handles gas estimation
- [x] Contract calls optimized (batch operations, minimal storage)
  - Single pool state query (already implemented in GetAMMState)
  - Two sequential approvals + one mint (cannot batch due to confirmation requirement)
- [x] Approval reuse implemented where safe
  - Must implement: check existing allowance before approval

### Principle 5: Fail-Safe Operation
- [x] All external calls have timeout/retry logic
  - Existing ContractClient and TxListener handle timeouts
- [x] Transaction failures logged with full context
  - Must implement: comprehensive error logging with tx hashes
- [x] Partial failures trigger rollback or safe termination
  - Sequential execution: if approval fails, no mint attempted
  - If mint fails after approvals, funds remain in wallet (safe state)
- [x] Slippage protection enforced
  - Must implement: calculate amount0Min/amount1Min from slippage parameter
- [x] Circuit breaker pattern implemented
  - N/A for this feature (single operation, no recurring execution)
- [x] Manual recovery documented
  - Documentation in quickstart.md will cover failure scenarios

**Violations**: None. All constitutional principles are satisfied by this feature design.

## Project Structure

### Documentation (this feature)

```text
specs/001-liquidity-staking/
├── plan.md              # This file (/speckit.plan command output)
├── spec.md              # Feature specification
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── mint-api.md      # Mint operation API contract
├── checklists/
│   └── requirements.md  # Specification quality checklist (completed)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
blackhole_dex/
├── blackhole.go                          # Main DEX interaction logic (Mint method completion)
├── blackhole_interfaces.go               # Interface definitions
├── types.go                              # Parameter types (MintParams, etc.)
├── internal/
│   └── util/
│       ├── amm.go                        # AMM math utilities (ComputeAmounts - existing)
│       └── validation.go                 # Input validation helpers (NEW)
├── pkg/
│   ├── contractclient/
│   │   └── contractclient.go             # Generic contract client (existing)
│   ├── txlistener/
│   │   └── txlistener.go                 # Transaction listener (existing)
│   └── types/
│       ├── priority.go                   # Gas priority types (existing)
│       └── transaction.go                # Transaction types (existing)
├── blackhole_test.go                     # Integration tests for Mint (NEW)
└── cmd/
    └── stake/                            # CLI for manual staking operations (NEW)
        └── main.go
```

**Structure Decision**: This is a single Go project following the existing structure. The feature extends the existing `Blackhole` type in `blackhole.go` by completing the `Mint` method. New files include validation helpers, integration tests, and a CLI tool for manual operator interaction. The structure maintains separation between blockchain interaction logic (root), utilities (internal/util), and reusable contract clients (pkg/).

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations - this section is empty per constitution compliance above.

## Phase 0: Research Findings

See [research.md](./research.md) for detailed research findings.

### Key Decisions

**Decision 1: NonfungiblePositionManager Contract Address**
- **What**: Address for the NonfungiblePositionManager contract on Avalanche C-Chain
- **Why**: Required for minting liquidity positions (not yet in blackhole.go constants)
- **Solution**: Add constant `nonfungiblePositionManager = "0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146"` to blackhole.go

**Decision 2: Approval Optimization Strategy**
- **What**: How to handle ERC20 approvals efficiently
- **Why**: Constitutional Principle 4 requires approval reuse when safe
- **Solution**: Check existing allowance via `allowance(owner, spender)` before calling `approve()`. Only approve if current allowance < required amount.

**Decision 3: Slippage Calculation Method**
- **What**: How to calculate amount0Min and amount1Min from slippage tolerance
- **Why**: Constitutional Principle 5 requires slippage protection
- **Solution**: `amountMin = amountDesired * (100 - slippagePercent) / 100` using big.Int arithmetic to avoid precision loss

**Decision 4: Gas Cost Tracking Implementation**
- **What**: How to retrieve actual gas cost for completed transactions
- **Why**: Constitutional Principle 3 requires comprehensive financial tracking
- **Solution**: After WaitForTransaction, retrieve receipt.GasUsed and multiply by receipt.EffectiveGasPrice to get native token cost

**Decision 5: Range Width Parameter Semantics**
- **What**: Interpretation of "range width parameter"
- **Why**: Spec requires configurable range width but calculation method not specified
- **Solution**: Range width `N` means ±(N/2) tick ranges from current tick, e.g., width=6 means ±3 ranges. Tick bounds calculated as:
  - `tickLower = (currentTick / tickSpacing - rangeWidth/2) * tickSpacing`
  - `tickUpper = (currentTick / tickSpacing + rangeWidth/2) * tickSpacing`

**Decision 6: Input Validation Requirements**
- **What**: What validations are needed before staking
- **Why**: Constitutional Principle 5 requires fail-safe operation
- **Solution**: Validate:
  - Range width > 0 and <= 20
  - Slippage tolerance > 0 and <= 50% (prevent accidental 100% slippage)
  - Token amounts > 0
  - Wallet balances >= requested amounts
  - Tick bounds within valid range (±887272)

## Phase 1: Design Artifacts

### Data Model

See [data-model.md](./data-model.md) for complete entity definitions.

**Core Entities**:
- **StakingRequest**: Input parameters for Mint operation
- **CalculatedPosition**: Intermediate computation results (tick bounds, amounts, liquidity)
- **StakingResult**: Output data (tx hashes, gas costs, NFT token ID)

### API Contracts

See [contracts/mint-api.md](./contracts/mint-api.md) for complete API specification.

**Core Operation**:
```go
func (b *Blackhole) Mint(
    maxWAVAX *big.Int,      // Maximum WAVAX to stake (wei)
    maxUSDC *big.Int,       // Maximum USDC to stake (smallest unit)
    rangeWidth int,         // Range width parameter (e.g., 6 = ±3 tick ranges)
    slippagePct int,        // Slippage tolerance (e.g., 5 = 5%)
) (*StakingResult, error)
```

**Supporting Functions**:
```go
// Validation
func validateStakingRequest(maxWAVAX, maxUSDC *big.Int, rangeWidth, slippagePct int) error

// Tick calculation
func calculateTickBounds(currentTick int32, rangeWidth int, tickSpacing int) (tickLower, tickUpper int32, err error)

// Slippage protection
func calculateMinAmounts(amount0Desired, amount1Desired *big.Int, slippagePct int) (amount0Min, amount1Min *big.Int)

// Approval optimization
func ensureApproval(tokenClient ContractClient, spender common.Address, amount *big.Int, owner common.Address, privateKey *ecdsa.PrivateKey) (txHash common.Hash, err error)

// Gas tracking
func extractGasCost(receipt *types.TxReceipt) (*big.Int, error)
```

### Quickstart Guide

See [quickstart.md](./quickstart.md) for operator usage instructions.

## Phase 2: Next Steps

This implementation plan is complete. Next phase is task generation via `/speckit.tasks` command, which will:

1. Break down implementation into sequential tasks
2. Organize by user story priority (P1, P2, P2)
3. Identify parallel execution opportunities
4. Define clear acceptance criteria per task
5. Map tasks to specific files and line ranges

The tasks will follow this general structure:
- **Phase 1: Setup** - Add constants, create validation helpers
- **Phase 2: Core Implementation** - Complete Mint method with all substeps
- **Phase 3: Testing** - Integration tests for all user scenarios
- **Phase 4: CLI Tool** - Manual operator interface for testing
- **Phase 5: Documentation** - Usage guides and error handling docs
