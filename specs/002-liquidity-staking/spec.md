# Feature Specification: Liquidity Position Staking in Gauge

**Feature Branch**: `002-liquidity-staking`
**Created**: 2025-12-16
**Status**: Draft
**Input**: User description: "implement stake method in blackholedex.go. It performs staking the depositied liquidity which are the results of the Mint in blackhole.go. It needs to approve NFT transfer in advance and deposit NFT. Just like the Mint method, track the transactions and totalGasCost."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Stake Minted Liquidity Position (Priority: P1)

A liquidity provider has successfully minted a liquidity position (receiving an NFT representing their WAVAX/USDC liquidity) and now wants to stake that position in the gauge to earn additional rewards and incentives on top of trading fees.

**Why this priority**: This is the core functionality needed to complete the liquidity provision workflow. Without staking, users cannot earn gauge rewards, making the minted position less profitable. This represents the minimum viable implementation.

**Independent Test**: Can be fully tested by minting a position (using existing `Mint` method), then calling `Stake` with the returned NFT token ID, and verifying the gauge contract shows the staked position with all transaction records properly tracked.

**Acceptance Scenarios**:

1. **Given** a wallet holds an NFT position token from previous `Mint` operation, **When** user calls `Stake` with the NFT token ID and gauge address, **Then** the NFT is transferred to the gauge contract and the position becomes eligible for gauge rewards
2. **Given** the NFT is not yet approved for transfer, **When** `Stake` is called, **Then** the system first executes approval transaction, waits for confirmation, then executes deposit transaction
3. **Given** the NFT is already approved for the gauge, **When** `Stake` is called, **Then** the system skips approval and directly executes deposit transaction (gas optimization)
4. **Given** staking completes successfully, **When** returning result, **Then** all transaction hashes, gas costs, and total gas cost are tracked and returned in structured format
5. **Given** wallet has insufficient balance to pay gas fees, **When** `Stake` is called, **Then** system returns clear error before submitting any transactions

---

### User Story 2 - Handle Staking Failures Safely (Priority: P2)

When staking operations fail (network issues, insufficient gas, contract reverts), the system must preserve fund safety and provide clear error information without leaving the position in an intermediate state.

**Why this priority**: Fail-safe operation is constitutional (Principle 5) and critical for user trust, but it's lower priority than the happy path since most transactions succeed under normal conditions.

**Independent Test**: Can be tested by simulating various failure scenarios (network timeout, gas estimation failure, contract revert) and verifying no funds are lost and error messages are actionable.

**Acceptance Scenarios**:

1. **Given** approval transaction fails or times out, **When** system detects failure, **Then** stake operation is aborted and NFT remains in user wallet with detailed error message
2. **Given** deposit transaction reverts, **When** system receives revert, **Then** error is logged with transaction hash and revert reason, and user retains NFT ownership
3. **Given** network connection fails during operation, **When** timeout occurs, **Then** system logs partial state and provides recovery instructions
4. **Given** gauge contract is paused or disabled, **When** attempting to stake, **Then** pre-validation catches issue and returns clear error without submitting transactions

---

### User Story 3 - Track Financial Impact Transparently (Priority: P3)

Users need complete visibility into the cost of staking operations (approval gas + deposit gas) to understand net profitability and compare against earned rewards.

**Why this priority**: Financial transparency is constitutional (Principle 3) and important for informed decision-making, but the staking operation can function without advanced reporting. Basic transaction tracking satisfies this requirement.

**Independent Test**: Can be tested by executing a stake operation and verifying that all transaction records include accurate gas used, gas price, and total cost fields matching on-chain data.

**Acceptance Scenarios**:

1. **Given** staking completes with both approval and deposit, **When** result is returned, **Then** `Transactions` array contains two records with distinct operations "ApproveNFT" and "DepositNFT"
2. **Given** staking completes with only deposit (approval already existed), **When** result is returned, **Then** `Transactions` array contains one record with operation "DepositNFT"
3. **Given** transaction records are generated, **When** accessing gas cost fields, **Then** `GasCost` equals `GasUsed * GasPrice` for each record and `TotalGasCost` equals sum of all `GasCost` values
4. **Given** timestamp tracking is enabled, **When** transactions complete, **Then** each record includes actual confirmation time

---

### Edge Cases

- What happens when NFT token ID doesn't exist or doesn't belong to user's wallet?
- What happens when gauge address is invalid or doesn't match expected GaugeV2 interface?
- How does system handle when approval exists but has been revoked between check and deposit?
- What happens when user attempts to stake an NFT that is already staked in a gauge?
- How does system behave when gas price spikes dramatically between estimation and submission?
- What happens when NFT position has zero liquidity (withdrawn but not burned)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST implement a `Stake` method that accepts NFT token ID and gauge address as parameters and returns a structured result with transaction tracking
- **FR-002**: System MUST verify NFT ownership before attempting any blockchain operations (validate token exists and belongs to user's address)
- **FR-003**: System MUST check existing NFT approval status for the gauge contract before submitting approval transaction
- **FR-004**: System MUST execute NFT approval transaction (calling `approve` or `setApprovalForAll` on NonfungiblePositionManager) only when current approval is insufficient
- **FR-005**: System MUST wait for approval transaction confirmation before proceeding to deposit
- **FR-006**: System MUST execute gauge deposit transaction (calling `deposit` on GaugeV2 contract) with the NFT token ID
- **FR-007**: System MUST wait for deposit transaction confirmation before returning success
- **FR-008**: System MUST track all submitted transactions (approval and/or deposit) in structured `TransactionRecord` format including tx hash, gas used, gas price, gas cost, timestamp, and operation type
- **FR-009**: System MUST calculate and return total gas cost as sum of all transaction gas costs
- **FR-010**: System MUST return comprehensive `StakingResult` structure (reusing existing type) with success status, error messages, and transaction history
- **FR-011**: System MUST handle transaction failures gracefully, returning error context without leaving NFT in intermediate state
- **FR-012**: System MUST validate gauge address corresponds to a valid GaugeV2 contract (basic validation, not exhaustive ABI checking)
- **FR-013**: System MUST use automatic gas estimation for all transactions (nil gas limit parameter)
- **FR-014**: System MUST log key operation steps (approval skipped/executed, deposit submitted, confirmation received) to aid debugging
- **FR-015**: System MUST reuse `TransactionRecord` structure and gas tracking patterns identical to `Mint` method for consistency

### Constitutional Compliance

This feature MUST adhere to the following constitutional constraints:

- **Scope**: Operations limited to WAVAX/USDC pool on Blackhole DEX (Principle 1) - gauge address must correspond to WAVAX/USDC pool gauge
- **Safety**: All operations include fail-safe error handling and rollback mechanisms (Principle 5) - NFT ownership is never compromised by partial failures
- **Transparency**: Financial impacts (gas, fees, profits) must be tracked and reportable (Principle 3) - all transaction costs are recorded and summed
- **Efficiency**: Gas costs must be minimized through batching and thresholds (Principle 4) - approval is skipped when existing approval is sufficient

### Key Entities *(include if feature involves data)*

- **NFT Position Token**: Represents user's liquidity position in WAVAX/USDC pool, issued by NonfungiblePositionManager contract, identified by token ID (uint256)
- **Gauge Contract**: GaugeV2 contract that accepts staked NFT positions and distributes rewards, requires NFT approval before deposit
- **Transaction Record**: Immutable record of on-chain transaction with hash, gas metrics, timestamp, and operation classification
- **Staking Result**: Comprehensive output structure tracking staking outcome, gas costs, and all executed transactions

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can stake a minted liquidity position NFT in under 90 seconds (including approval + deposit + confirmations on Avalanche C-Chain)
- **SC-002**: System successfully stakes positions with 100% accuracy when wallet has sufficient gas and NFT ownership is valid
- **SC-003**: Gas cost tracking matches on-chain transaction receipts with 100% accuracy (zero discrepancy between calculated and actual costs)
- **SC-004**: Transaction failures result in zero NFT ownership loss (100% fund safety across all error scenarios)
- **SC-005**: Unnecessary approval transactions are avoided 90% of the time through existing approval reuse (gas optimization effectiveness)
- **SC-006**: Error messages provide actionable information in 100% of failure cases (user knows what failed and why)
