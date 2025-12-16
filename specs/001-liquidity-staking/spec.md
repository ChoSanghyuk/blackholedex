# Feature Specification: Liquidity Staking

**Feature Branch**: `001-liquidity-staking`
**Created**: 2025-12-09
**Status**: Draft
**Input**: User description: "given max usdc, wavax amount, stake them in the wavax-usdc pool. How broad the target range would be can be determined by the parameter. complete Mint method pre-exisiting in the blackhole.go and make corresdonding methods to do staking."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Single-Step Liquidity Staking (Priority: P1)

An operator provides maximum WAVAX and USDC amounts along with a range width parameter, and the system stakes liquidity in a single operation within the active trading range of the WAVAX-USDC pool.

**Why this priority**: This is the core functionality that enables capital deployment. Without this, the entire liquidity management system cannot function. It provides immediate value by allowing operators to deploy capital efficiently into the DEX.

**Independent Test**: Can be fully tested by calling the staking operation with specific token amounts and range parameters, then verifying the resulting liquidity position exists on-chain with correct tick bounds and token amounts.

**Acceptance Scenarios**:

1. **Given** operator has 10 WAVAX and 500 USDC in wallet, **When** operator stakes with range width parameter of 6 (±3 tick ranges from current), **Then** system creates liquidity position with tick bounds centered around current pool price using all available capital within slippage tolerance
2. **Given** operator has 5 WAVAX and 1000 USDC (unbalanced amounts), **When** operator stakes with range width parameter of 4, **Then** system calculates optimal amounts for the target range, uses maximum possible capital without exceeding either token balance, and stakes position successfully
3. **Given** operator specifies 2% slippage tolerance, **When** staking operation encounters price movement during execution, **Then** system enforces minimum token amounts based on slippage protection and fails transaction if minimum requirements cannot be met
4. **Given** liquidity position is successfully staked, **When** operation completes, **Then** system returns the transaction hash and position details (NFT token ID, actual amounts staked, final tick bounds)

---

### User Story 2 - Configurable Range Width (Priority: P2)

An operator adjusts the range width parameter to control how concentrated or broad the liquidity position should be, balancing capital efficiency against rebalancing frequency.

**Why this priority**: Range width directly impacts strategy effectiveness and cost. Narrower ranges generate higher fees but require frequent rebalancing; wider ranges reduce rebalancing costs but earn less per unit capital. This configurability is essential for optimization but the system can function with a fixed default range.

**Independent Test**: Can be tested by creating multiple positions with different range width parameters (e.g., 2, 4, 6, 8) and verifying the resulting tick bounds match expected widths centered on current price.

**Acceptance Scenarios**:

1. **Given** current pool tick is 10000, **When** operator stakes with range width parameter of 2 (±1 tick range), **Then** system creates position with tickLower = 9800 and tickUpper = 10200 (assuming 200 tick spacing)
2. **Given** current pool tick is 10000, **When** operator stakes with range width parameter of 8 (±4 tick ranges), **Then** system creates position with tickLower = 9200 and tickUpper = 10800
3. **Given** operator provides invalid range width (0 or negative), **When** staking is attempted, **Then** system rejects operation with clear error message indicating valid range requirements

---

### User Story 3 - Balanced Token Calculation (Priority: P2)

The system automatically calculates the optimal ratio of WAVAX to USDC needed for the specified price range, ensuring maximum capital deployment without leaving unused tokens.

**Why this priority**: Manual ratio calculation is error-prone and leads to capital inefficiency. Automatic calculation ensures optimal capital utilization. While important for efficiency, the system could function with manual pre-calculation, making this P2.

**Independent Test**: Can be tested by providing various unbalanced token amounts and verifying the system stakes the maximum possible while maintaining correct ratio for the range and respecting both token balance limits.

**Acceptance Scenarios**:

1. **Given** operator has 100 WAVAX and 100 USDC but current range requires 1:50 ratio (WAVAX:USDC), **When** staking is initiated, **Then** system uses 100 WAVAX and 100 USDC by adjusting to the optimal amounts that fit within available balances
2. **Given** operator has insufficient balance of one token to match the required ratio, **When** staking is initiated, **Then** system calculates maximum stakeable amount using the constraining token and uses proportional amount of the other token
3. **Given** operator specified amounts that would leave significant unused capital (>10% of either token), **When** calculation completes, **Then** system logs a warning about capital inefficiency while still proceeding with the stake

---

### Edge Cases

- What happens when pool price moves significantly between calculation and transaction execution (slippage exceeds tolerance)?
- How does system handle when wallet balance is insufficient for even minimum liquidity stake?
- What happens if the current pool tick is at extreme boundaries and requested range extends beyond valid tick range?
- How does system behave when gas estimation suggests transaction will fail before submission?
- What happens when token approvals already exist vs. when new approvals are needed?
- How does system handle network failures or RPC timeouts during multi-step approval and mint operations?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept maximum WAVAX amount, maximum USDC amount, and range width parameter as inputs for staking operation
- **FR-002**: System MUST query current WAVAX-USDC pool state to determine current tick and price before calculating position bounds
- **FR-003**: System MUST calculate tick bounds (tickLower and tickUpper) based on current pool tick and range width parameter, respecting pool's tick spacing (200 ticks)
- **FR-004**: System MUST calculate optimal token amounts needed for the specified range using current pool price and tick bounds
- **FR-005**: System MUST verify wallet has sufficient balance of both WAVAX and USDC before proceeding with transaction
- **FR-006**: System MUST approve NonfungiblePositionManager contract to spend required amounts of both WAVAX and USDC tokens
- **FR-007**: System MUST call the mint function on NonfungiblePositionManager contract with calculated MintParams
- **FR-008**: System MUST enforce slippage protection by calculating minimum token amounts (amount0Min, amount1Min) based on configurable slippage tolerance percentage
- **FR-009**: System MUST set transaction deadline to prevent execution after excessive delay (default: 20 minutes from submission)
- **FR-010**: System MUST wait for approval transactions to confirm before submitting mint transaction
- **FR-011**: System MUST return transaction hash upon successful submission and position details (NFT token ID) after confirmation
- **FR-012**: System MUST log all transaction hashes, gas costs, and position parameters for financial tracking
- **FR-013**: System MUST handle transaction failures gracefully and return descriptive error messages indicating failure reason (insufficient balance, slippage exceeded, deadline passed, etc.)
- **FR-014**: System MUST validate range width parameter is positive and within reasonable bounds (e.g., 1-20 tick ranges)
- **FR-015**: System MUST support configurable slippage tolerance percentage with reasonable default (e.g., 2%)

### Constitutional Compliance

This feature MUST adhere to the following constitutional constraints:

- **Scope**: Operations limited to WAVAX/USDC pool on Blackhole DEX (Principle 1)
  - Only WAVAX (0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7) and USDC (0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E) tokens accepted
  - Only WAVAX-USDC pool (0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0) for state queries
  - Only designated NonfungiblePositionManager (0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146) for minting

- **Safety**: All operations include fail-safe error handling and rollback mechanisms (Principle 5)
  - Balance validation before approval/mint
  - Slippage protection on all liquidity additions
  - Transaction deadline enforcement
  - Approval confirmation before mint submission
  - Comprehensive error logging with transaction context

- **Transparency**: Financial impacts (gas, fees, profits) must be tracked and reportable (Principle 3)
  - Log approval transaction hashes and gas costs
  - Log mint transaction hash and gas costs
  - Record actual amounts staked vs. requested
  - Record position NFT token ID for tracking
  - Track total gas expenditure for complete staking operation

- **Efficiency**: Gas costs must be minimized through batching and thresholds (Principle 4)
  - Reuse existing token approvals when possible
  - Use gas estimation before transaction submission
  - Avoid unnecessary state reads (single pool state query)
  - Batch approvals for both tokens where possible

### Key Entities

- **Staking Request**: Represents operator's intent to stake liquidity with maximum token amounts and range width preference
  - Maximum WAVAX amount (wei)
  - Maximum USDC amount (smallest unit)
  - Range width parameter (tick ranges from current price, e.g., 3 means ±3 tick ranges)
  - Slippage tolerance percentage (e.g., 2.0 for 2%)

- **Position Bounds**: Calculated tick range for liquidity position
  - Tick lower (calculated as: (currentTick / tickSpacing - rangeWidth) * tickSpacing)
  - Tick upper (calculated as: (currentTick / tickSpacing + rangeWidth) * tickSpacing)
  - Tick spacing (always 200 for this pool)

- **Optimal Amounts**: Calculated token amounts needed for the specified range
  - Amount0Desired (WAVAX amount in wei)
  - Amount1Desired (USDC amount in smallest unit)
  - Amount0Min (WAVAX minimum after slippage protection)
  - Amount1Min (USDC minimum after slippage protection)

- **Liquidity Position**: On-chain NFT representing staked liquidity
  - NFT Token ID (unique identifier for the position)
  - Actual amount0 staked (may differ from desired due to price movements)
  - Actual amount1 staked (may differ from desired due to price movements)
  - Final tick bounds (tickLower, tickUpper)
  - Position liquidity amount

- **Transaction Record**: Financial tracking for staking operation
  - Approval transaction hashes (WAVAX and USDC)
  - Mint transaction hash
  - Gas cost for each transaction
  - Total gas expenditure
  - Timestamp of operation

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Operator can stake liquidity in the WAVAX-USDC pool with a single function call, completing the operation (including approvals) in under 5 minutes under normal network conditions
- **SC-002**: System successfully stakes at least 95% of provided capital when token balances are reasonably balanced for the target range (within 2x of optimal ratio)
- **SC-003**: Slippage protection prevents position creation when actual amounts would deviate more than specified tolerance percentage from calculated amounts
- **SC-004**: Position tick bounds are correctly calculated to be symmetric around current pool tick with width matching the specified range parameter (±rangeWidth tick ranges)
- **SC-005**: All transaction hashes and gas costs are logged for every staking operation, enabling complete financial tracking
- **SC-006**: System handles network errors and transaction failures without leaving funds in intermediate states (approved but not staked), failing safely with descriptive error messages
- **SC-007**: Range width parameter allows operators to create positions ranging from highly concentrated (width=1, ±200 ticks) to broad (width=10, ±2000 ticks)
- **SC-008**: Gas estimation accurately predicts transaction success/failure before submission in at least 95% of cases, preventing unnecessary failed transaction costs

## Assumptions

- Pool uses 200 tick spacing (standard for concentrated liquidity pools on Blackhole DEX)
- Default slippage tolerance of 5% is appropriate for typical market conditions
- Default transaction deadline of 20 minutes provides sufficient time for normal network confirmation
- Wallet has sufficient native tokens (AVAX) to pay for gas fees
- RPC endpoint is reliable and responds within reasonable timeout (assumed handled by existing ContractClient implementation)
- MintParams structure and NonfungiblePositionManager ABI are correctly configured in the existing codebase
- Existing ContractClient implementation handles gas estimation, transaction signing, and submission correctly
- Token approvals use standard ERC20 approve function
- Range width parameter of 3 (±3 tick ranges, or ±600 ticks) is a reasonable default for balanced capital efficiency and rebalancing frequency
