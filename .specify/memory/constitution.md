<!--
SYNC IMPACT REPORT
==================
Version: 1.0.0 (Initial Constitution)
Date: 2025-12-09

Established Principles:
- PRINCIPLE 1: Pool-Specific Scope (WAVAX/USDC only)
- PRINCIPLE 2: Autonomous Rebalancing
- PRINCIPLE 3: Financial Transparency
- PRINCIPLE 4: Gas Optimization
- PRINCIPLE 5: Fail-Safe Operation

Templates Sync Status:
- ✅ spec-template.md: Constitutional compliance section added, requirements comments updated
- ✅ plan-template.md: Constitution check section added with all 5 principles checklist
- ✅ tasks-template.md: No changes needed (principle-aligned task structure already present)
- ✅ checklist-template.md: Constitutional compliance checklist section added (15 items)
- ✅ agent-file-template.md: Constitutional principles section added

Follow-up Items:
- None (all templates synchronized)

Ratification Notes:
- Initial constitution created based on user requirements for automated liquidity repositioning
- Project scope limited to WAVAX/USDC pool on Blackhole DEX
- Focus on autonomous operation with comprehensive financial tracking
- All templates updated to reference and enforce constitutional principles
- Version 1.0.0 establishes baseline governance for the project
-->

# Project Constitution: Blackhole DEX Liquidity Agent

**Project Name**: blackholego
**Version**: 1.0.0
**Ratification Date**: 2025-12-09
**Last Amended**: 2025-12-09
**Status**: Active

---

## Preamble

This constitution establishes the foundational principles and governance rules for the Blackhole DEX Liquidity Agent (blackholego), an autonomous system designed to optimize liquidity provision on the Avalanche network's Blackhole DEX. The system MUST automatically reposition liquidity within the WAVAX/USDC pool to maintain capital efficiency while providing transparent financial tracking.

All development, operations, and modifications to this project MUST adhere to the principles defined herein.

---

## Core Principles

### PRINCIPLE 1: Pool-Specific Scope

**Statement**: The system MUST operate exclusively within the WAVAX/USDC concentrated liquidity pool on Blackhole DEX. No other pools, tokens, or protocols SHALL be supported.

**Rationale**: This constraint ensures focused optimization, reduces complexity, minimizes attack surface, and allows for deep specialization in a single pool's dynamics. By restricting scope, the system can implement pool-specific strategies without generic abstraction overhead.

**Implementation Requirements**:
- All swap operations MUST involve only WAVAX and USDC tokens
- Liquidity positions MUST be created only in the designated WAVAX/USDC pool
- Contract interactions MUST be restricted to verified Blackhole DEX contracts
- Configuration MUST reject attempts to add other pools or tokens
- Error messages MUST clearly indicate when operations violate pool scope

**Testing Criteria**:
- Attempts to interact with other pools MUST fail with explicit errors
- Token allowlists MUST contain only WAVAX and USDC addresses
- Integration tests MUST verify pool address validation

---

### PRINCIPLE 2: Autonomous Rebalancing

**Statement**: The system MUST monitor pool state continuously and execute rebalancing operations automatically when the active trading range moves outside the current liquidity position, without requiring manual intervention.

**Rationale**: Concentrated liquidity pools require active management to maintain capital efficiency. Manual monitoring is error-prone and slow. Autonomous operation ensures positions remain optimally positioned to capture trading fees and liquidity incentives.

**Implementation Requirements**:
- Pool state monitoring MUST occur at configurable intervals (default: every block or every N seconds)
- Position tracking MUST identify when active price range exits staked liquidity bounds
- Rebalancing workflow MUST execute sequentially:
  1. Unstake existing liquidity position
  2. Collect earned fees and incentives
  3. Swap tokens to rebalance ratio based on target range
  4. Stake new liquidity position in updated price range
- All operations MUST complete atomically where possible, or implement rollback on partial failure
- System MUST log decision rationale for each rebalancing event

**Testing Criteria**:
- Simulated price movements MUST trigger rebalancing when thresholds exceeded
- Rebalancing MUST NOT occur when position remains in-range
- Failed rebalancing MUST NOT leave funds in intermediate states
- Recovery mechanisms MUST restore operations after transient failures

---

### PRINCIPLE 3: Financial Transparency

**Statement**: The system MUST track and report ALL financial flows including gas fees, swap fees, liquidity provision incentives, and net profit/loss with complete accuracy and auditability.

**Rationale**: Users MUST have visibility into true profitability. Hidden costs (gas, slippage, swap fees) can erode returns. Transparent accounting enables informed decisions about strategy effectiveness and operational costs.

**Implementation Requirements**:
- Gas tracking MUST record actual gas consumed for every transaction with native token cost
- Swap fee tracking MUST calculate fees paid during token rebalancing operations
- Incentive tracking MUST record all rewards claimed from liquidity provision (trading fees, gauge rewards, bribes)
- Profit calculation MUST compute net returns: (incentives + fees earned) - (gas costs + swap fees paid)
- All financial data MUST be exportable in structured format (JSON, CSV)
- Historical records MUST be persisted with timestamps and transaction hashes
- Reporting MUST distinguish between realized and unrealized gains

**Testing Criteria**:
- Mock transactions MUST produce accurate gas cost records
- Fee calculations MUST match on-chain values within acceptable precision
- Profit reports MUST reconcile with wallet balance changes
- Historical data MUST be retrievable for arbitrary time ranges

---

### PRINCIPLE 4: Gas Optimization

**Statement**: The system MUST minimize gas consumption through batching, efficient contract calls, and avoiding unnecessary operations, while maintaining safety and correctness.

**Rationale**: On Avalanche C-Chain, gas costs directly reduce profitability. Frequent rebalancing can incur substantial costs. Optimization MUST balance responsiveness with cost efficiency.

**Implementation Requirements**:
- Rebalancing triggers MUST use hysteresis/threshold logic to prevent excessive transactions from minor price movements
- Contract calls MUST use gas-efficient patterns (minimal storage reads, batch operations where possible)
- Token approvals MUST use appropriate allowances (not infinite unless intentionally chosen) and reuse existing approvals
- Simulation/estimation MUST occur before transaction submission to avoid failed transactions
- Monitoring MUST use view functions and event logs rather than state-changing calls
- Configuration MUST allow tuning of rebalancing frequency vs gas cost trade-offs

**Testing Criteria**:
- Gas usage benchmarks MUST be established for each operation type
- Regressions in gas consumption MUST be detected in CI/CD
- Simulated high-frequency price volatility MUST NOT trigger excessive rebalancing
- Gas estimation accuracy MUST exceed 95%

---

### PRINCIPLE 5: Fail-Safe Operation

**Statement**: The system MUST operate defensively with comprehensive error handling, validation, and safeguards to prevent loss of funds under any failure scenario.

**Rationale**: Autonomous systems managing real funds MUST prioritize safety over performance. Network failures, smart contract reverts, or unexpected state changes MUST NOT result in locked or lost funds.

**Implementation Requirements**:
- All external calls (RPC, smart contracts) MUST have timeout and retry logic
- Transaction failures MUST be logged with full context (tx hash, revert reason, state)
- Partial rebalancing failures MUST trigger rollback or safe termination (never leave funds in intermediate contracts)
- Slippage protection MUST be enforced on all swaps with configurable tolerance
- Position validation MUST verify expected state after each operation
- Circuit breaker MUST halt operations if error rate exceeds threshold
- Manual recovery mechanisms MUST be documented and tested
- System MUST operate read-only when RPC/network issues detected

**Testing Criteria**:
- Chaos testing MUST inject RPC failures, contract reverts, network partitions
- All failure paths MUST preserve fund safety
- Recovery procedures MUST successfully restore operations
- Alert mechanisms MUST notify operators of critical failures

---

## Governance

### Amendment Procedure

This constitution MAY be amended through the following process:

1. **Proposal**: Amendments MUST be proposed in writing with clear rationale
2. **Review**: Proposed changes MUST be reviewed against existing implementations
3. **Impact Analysis**: A Sync Impact Report MUST identify all affected templates, code, and documentation
4. **Approval**: Amendments require explicit approval from project maintainers
5. **Propagation**: All dependent artifacts MUST be updated before amendment is considered complete
6. **Documentation**: Amendment history MUST be preserved in this document

### Versioning Policy

Constitution versions MUST follow semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Backward-incompatible changes to core principles or governance structure (e.g., removing a principle, changing scope)
- **MINOR**: New principles added, material expansion of existing principles, or new governance sections
- **PATCH**: Clarifications, wording improvements, typo corrections, or non-semantic refinements

### Compliance Review

All feature specifications, implementation plans, and code changes MUST be reviewed for constitutional compliance:

- Features that violate any principle MUST be rejected or principles MUST be amended first
- Implementation plans MUST include a "Constitution Check" section explicitly mapping to relevant principles
- Code reviews MUST verify adherence to principle-driven requirements
- Violations discovered post-implementation MUST be remediated or granted exception via amendment

### Exception Handling

Temporary exceptions to constitutional principles MAY be granted under these conditions:

- Exception MUST be documented with specific justification and expiration
- Exception MUST NOT compromise Principle 5 (Fail-Safe Operation) or fund safety
- Exception MUST be reviewed for permanent amendment consideration before expiration
- Active exceptions MUST be listed in implementation plans and commit messages

---

## Technical Constraints

The following technical constraints are established by this constitution:

### Supported Networks
- Avalanche C-Chain ONLY (mainnet and designated testnets)

### Supported Protocols
- Blackhole DEX ONLY (no Uniswap, Trader Joe, or other DEX integrations)

### Supported Tokens
- WAVAX (0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7)
- USDC (0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E)

### Required Contracts
- RouterV2: 0x04E1dee021Cd12bBa022A72806441B43d8212Fec
- WAVAX/USDC Pair: 0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0
- NonfungiblePositionManager: 0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146

### Monitoring Requirements
- Pool state MUST be checked at least once per minute
- Position state MUST be validated after every operation
- Financial records MUST be persisted within 5 minutes of transaction confirmation

---

## Interpretation

In case of ambiguity or conflict:

1. Safety principles (Principle 5) MUST take precedence over optimization principles (Principle 4)
2. Scope constraints (Principle 1) MUST be interpreted strictly
3. Transparency requirements (Principle 3) MUST be interpreted broadly
4. When principles conflict, the order of precedence is: 5 > 3 > 1 > 2 > 4

---

## Acknowledgments

This constitution is maintained by the blackholego project maintainers and reflects the requirements and constraints necessary for safe, effective, autonomous liquidity management on Blackhole DEX.

---

**END OF CONSTITUTION**
