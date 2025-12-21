# Feature Specification: Liquidity Position Withdrawal

**Feature Branch**: `003-liquidity-withdraw`
**Created**: 2025-12-21
**Status**: Draft
**Input**: User description: "implement the withdraw of deposited liquidity in the blackhole.go. When the Mint function in blackhole.go deposits the liquidity, Withdraw retrieves it. It sends the transaction to nonfungiblePositionManager contract with multicall which contains decreaseLiquidity, collect and burn. The function should take NFTID for parameter."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Complete Position Withdrawal (Priority: P1)

A liquidity provider wants to fully exit their position by withdrawing all deposited tokens from an NFT position created by the Mint function.

**Why this priority**: This is the core functionality - users must be able to retrieve their capital. Without this, the Mint function would lock funds permanently, making the system unusable.

**Independent Test**: Can be fully tested by minting a position with known token amounts, then withdrawing using the NFT ID and verifying all tokens are returned to the wallet.

**Acceptance Scenarios**:

1. **Given** a user owns an NFT position with deposited liquidity, **When** they initiate withdrawal with the NFT ID, **Then** all liquidity is removed, fees are collected, the NFT is burned, and tokens are returned to their wallet
2. **Given** a position with accumulated fees, **When** withdrawal is executed, **Then** both the original liquidity tokens and accumulated fees are collected and returned
3. **Given** a successful withdrawal, **When** the transaction completes, **Then** the NFT no longer exists and the position is fully closed

---

### User Story 2 - Error Handling for Invalid Operations (Priority: P2)

A user attempts to withdraw with an invalid NFT ID or an NFT they don't own, and receives clear feedback about what went wrong.

**Why this priority**: Prevents user confusion and provides diagnostic information when operations fail. Essential for usability but not as critical as the core withdrawal functionality.

**Independent Test**: Can be tested by attempting withdrawals with non-existent NFT IDs, NFT IDs owned by other addresses, and validating appropriate error messages are returned.

**Acceptance Scenarios**:

1. **Given** a user provides an NFT ID they don't own, **When** they attempt withdrawal, **Then** the operation fails with a clear ownership error message
2. **Given** a user provides a non-existent NFT ID, **When** they attempt withdrawal, **Then** the operation fails with a clear error indicating the NFT doesn't exist
3. **Given** a user provides an invalid NFT ID (negative or zero), **When** they attempt withdrawal, **Then** the operation fails validation before any blockchain interaction

---

### User Story 3 - Transaction Cost Transparency (Priority: P3)

A user wants to see detailed information about gas costs and transaction outcomes from their withdrawal operation.

**Why this priority**: Enhances user experience and aids debugging, but the core withdrawal can function without detailed reporting.

**Independent Test**: Can be tested by executing a withdrawal and verifying that transaction details (gas used, gas price, total cost, operation type) are tracked and returned.

**Acceptance Scenarios**:

1. **Given** a withdrawal transaction completes successfully, **When** the user reviews the result, **Then** they see gas cost information for each operation in the multicall
2. **Given** a withdrawal involves multiple contract interactions, **When** the transaction completes, **Then** the total gas cost across all operations is calculated and reported

---

### Edge Cases

- What happens when the NFT position has already been burned or doesn't exist?
- How does the system handle positions with zero liquidity remaining?
- What occurs if the user has insufficient gas to complete the withdrawal transaction?
- How are slippage conditions handled during liquidity removal?
- What happens if the multicall partially succeeds (one operation succeeds but another fails)?
- How does the system respond if the user's NFT is currently staked in a farming contract?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept an NFT token ID as the withdrawal parameter
- **FR-002**: System MUST validate that the NFT token ID is owned by the calling user before proceeding
- **FR-003**: System MUST execute a multicall transaction to the nonfungiblePositionManager contract containing three operations: decreaseLiquidity, collect, and burn
- **FR-004**: System MUST remove 100% of the liquidity from the position via decreaseLiquidity
- **FR-005**: System MUST collect all accumulated fees and remaining tokens via the collect operation
- **FR-006**: System MUST burn the NFT after liquidity removal and collection via the burn operation
- **FR-007**: System MUST track transaction details including gas costs, transaction hashes, and operation types
- **FR-008**: System MUST return a result structure indicating success or failure with detailed error messages
- **FR-009**: System MUST validate input parameters (NFT ID must be positive and non-zero) before initiating blockchain transactions
- **FR-010**: System MUST calculate minimum token amounts for slippage protection during decreaseLiquidity
- **FR-011**: System MUST set an appropriate deadline for the withdrawal transaction
- **FR-012**: System MUST handle transaction failures gracefully and return informative error messages
- **FR-013**: System MUST retrieve position details (liquidity amount, tick bounds, tokens) before constructing the withdrawal transaction

### Constitutional Compliance

This feature MUST adhere to the following constitutional constraints:

- **Scope**: Operations limited to WAVAX/USDC pool on Blackhole DEX (Principle 1)
- **Safety**: All operations include fail-safe error handling and rollback mechanisms (Principle 5)
- **Transparency**: Financial impacts (gas, fees, profits) must be tracked and reportable (Principle 3)
- **Efficiency**: Gas costs must be minimized through batching via multicall (Principle 4)

### Key Entities

- **NFT Position**: Represents a liquidity position in the nonfungiblePositionManager, identified by a unique token ID, containing deposited token amounts, tick range, and accumulated fees
- **Withdrawal Result**: Contains transaction records, gas costs, amounts withdrawn, success status, and error messages
- **Transaction Record**: Tracks individual transaction details including hash, gas used, gas price, timestamp, and operation type
- **Multicall Parameters**: Encoded function calls for decreaseLiquidity, collect, and burn operations

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully withdraw 100% of their deposited liquidity and accumulated fees in a single transaction
- **SC-002**: Withdrawal operations complete within 2 minutes under normal network conditions
- **SC-003**: Invalid withdrawal attempts (wrong ownership, invalid ID) are rejected before spending gas on failed blockchain transactions
- **SC-004**: All withdrawal transactions include complete gas cost tracking with accuracy to the wei level
- **SC-005**: Users receive clear, actionable error messages for all failure scenarios (ownership, validation, transaction failures)
- **SC-006**: NFT positions are completely removed after successful withdrawal (burn verification)
