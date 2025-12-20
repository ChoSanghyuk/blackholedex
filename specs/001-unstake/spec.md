# Feature Specification: Unstake Liquidity from Blackhole DEX

**Feature Branch**: `001-unstake`
**Created**: 2025-12-18
**Status**: Draft
**Input**: User description: "implement unstake in the blackhole dex. It calls the FarmingCenter contract of multicall(bytes[] data)."

## User Scenarios & Testing *(mandatory)*


### User Story 2 -  Withdrawal (Priority: P1)

A user needs to quickly exit their position during market volatility or contract migration scenarios. They should be able to execute an unstake operation, claiming unclaimed rewards.

**Why this priority**: withdrawals provide users with a fail-safe option during critical situations, enhancing user trust and safety.

**Independent Test**: Can be tested independently by staking LP tokens, then executing an emergency unstake and verifying that LP tokens and rewards are returned.

**Acceptance Scenarios**:

1. **Given** a user has staked LP tokens with unclaimed rewards, **When** they execute an unstake, **Then** their LP tokens and reward are returned to their wallet.

---



### Edge Cases

- What happens when a user attempts to unstake while a reward distribution is in progress?
- How does the system handle unstake requests when the user's staked balance is exactly zero?
- What happens if the FarmingCenter contract is paused during an unstake operation?
- How does the system handle gas estimation failures for the multicall operation?
- What happens when unstaking the last remaining LP tokens from the contract (dust amounts)?
- How does the system handle reentrancy attack attempts during unstake operations?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-002**: System MUST interact with the FarmingCenter contract using the multicall(bytes[] data) function
- **FR-003**: System MUST validate that the user has sufficient staked balance before executing unstake operations
- **FR-004**: System MUST encode the unstake operation parameters correctly for the multicall bytes array
- **FR-005**: System MUST return LP tokens to the user's wallet address upon successful unstake
- **FR-006**: System MUST handle any accumulated rewards during the unstake process (either claim them or preserve them for later claim)
- **FR-007**: System MUST provide clear error messages when unstake operations fail (insufficient balance, contract paused, etc.)
- **FR-008**: System MUST support unstaking the full staked balance in a single operation
- **FR-009**: System MUST support partial unstake operations that leave remaining LP tokens staked
- **FR-010**: System MUST validate that unstake amount is greater than zero
- **FR-011**: System MUST update the user's staked balance after successful unstake
- **FR-012**: System MUST emit events or provide transaction receipts that can be parsed to confirm unstake success
- **FR-013**: System MUST handle the specific LP token pair (WAVAX/USDC) as defined in the project scope

### Constitutional Compliance

This feature MUST adhere to the following constitutional constraints:

- **Scope**: Operations limited to WAVAX/USDC pool on Blackhole DEX (Principle 1)
- **Safety**: All operations include fail-safe error handling and rollback mechanisms - failed unstake attempts must not modify user balances (Principle 5)
- **Transparency**: Financial impacts (gas costs, LP tokens withdrawn, rewards claimed) must be tracked and reportable (Principle 3)
- **Efficiency**: Gas costs must be minimized through proper multicall batching and avoiding unnecessary operations (Principle 4)

### Key Entities

- **Staked Position**: Represents a user's LP tokens deposited in the FarmingCenter contract, including the amount staked, timestamp, and accumulated rewards
- **Unstake Request**: Represents a request to withdraw LP tokens, including the amount to unstake, the user's address, and the target FarmingCenter contract
- **Multicall Data**: Encoded bytes array containing the unstake operation parameters formatted for the FarmingCenter's multicall function
- **LP Token**: The WAVAX/USDC liquidity pool token that users stake and unstake
- **Transaction Receipt**: Record of the unstake operation including gas used, success/failure status, and emitted events

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully unstake LP tokens and receive them in their wallet within one blockchain transaction confirmation time
- **SC-002**: System correctly handles 100% of valid unstake requests (amount <= staked balance) without data corruption
- **SC-003**: Unstake operations consume no more than 150% of the gas cost of a standard ERC20 transfer (ensuring cost efficiency)
- **SC-004**: All failed unstake attempts return clear error messages that allow users to understand and resolve the issue
- **SC-005**: System maintains accurate staked balance records with zero discrepancies across all unstake operations
- **SC-006**: Users can verify their unstake transaction success through transaction receipt parsing with 100% accuracy

## Assumptions

- The FarmingCenter contract address is known and accessible
- The FarmingCenter contract's multicall function accepts an array of encoded function calls
- Users have already approved the FarmingCenter contract to handle their LP tokens (approval was done during staking)
- The WAVAX/USDC LP token contract follows standard ERC20 interfaces
- Unstaking does not require a lock-up period or cooldown time (unless specified by the FarmingCenter contract)
- The system has access to read the user's current staked balance from the FarmingCenter contract
- Reward handling behavior during unstake follows the FarmingCenter contract's existing implementation
- The Go server has the necessary ABI definitions for the FarmingCenter contract to encode multicall data
