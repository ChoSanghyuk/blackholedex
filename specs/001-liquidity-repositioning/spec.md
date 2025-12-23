# Feature Specification: Automated Liquidity Repositioning Strategy

**Feature Branch**: `001-liquidity-repositioning`
**Created**: 2025-12-23
**Status**: Draft
**Input**: User description: "implement RunStrategy1 method on Blackhole struct in blackhole.go. This strategy automatically repositions the liquidity in dex by using Swap, Mint, Stake, Unstake, Withdraw methods predefined on it. First, this rebalance the WAVAX and USDC to a 50:50 ratio. Second, it provides the liquidity to the wavaxUsdcPair pool with 10 width. Third, it keeps tracking the price. 4th, when the price goes out of active zone which avoids me from getting incentives, then pull out the liquidity. 5th, repeat from the first step but only provide the liquidity  when the price is stabilized (no dynamic price changes)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Initial Position Entry (Priority: P1)

A liquidity provider wants to deploy capital into the WAVAX/USDC pool with optimal token ratios and start earning incentives without manually calculating balances or price ranges.

**Why this priority**: This is the core entry point for the strategy. Without it, no liquidity can be provided and no subsequent operations are possible. It delivers immediate value by automating the complex process of ratio balancing and position creation.

**Independent Test**: Can be fully tested by starting with unbalanced WAVAX/USDC holdings, executing the strategy, and verifying that a staked liquidity position exists with correctly balanced tokens within the target price range.

**Acceptance Scenarios**:

1. **Given** user holds 1000 USDC and 10 WAVAX (unbalanced), **When** strategy initiates position entry, **Then** system swaps tokens to achieve 50:50 value ratio, creates liquidity position with 10 tick width around current price, and stakes the position
2. **Given** user holds only WAVAX (100% imbalance), **When** strategy initiates position entry, **Then** system swaps 50% of WAVAX to USDC and creates balanced liquidity position
3. **Given** user holds only USDC (100% imbalance), **When** strategy initiates position entry, **Then** system swaps 50% of USDC to WAVAX and creates balanced liquidity position

---

### User Story 2 - Continuous Price Monitoring (Priority: P2)

The strategy must continuously monitor pool price movements to detect when the active trading range moves outside the current liquidity position's bounds, ensuring the user doesn't miss out on incentives due to out-of-range positions.

**Why this priority**: Monitoring enables the autonomous behavior that makes this strategy valuable. Without it, the position would become inactive and stop earning incentives. This is second priority because it builds on the initial position entry.

**Independent Test**: Can be tested by creating a staked position, simulating price movements via mock RPC data, and verifying that the system correctly identifies when price exits the active range without executing any rebalancing.

**Acceptance Scenarios**:

1. **Given** liquidity position staked with range [100-110] USDC per WAVAX, **When** current price is 105 (within range), **Then** system continues monitoring without triggering rebalancing
2. **Given** liquidity position staked with range [100-110] USDC per WAVAX, **When** current price moves to 112 (out of range), **Then** system detects out-of-range condition and flags position for rebalancing
3. **Given** liquidity position staked with range [100-110] USDC per WAVAX, **When** current price moves to 98 (out of range), **Then** system detects out-of-range condition and flags position for rebalancing

---

### User Story 3 - Automated Position Rebalancing (Priority: P1)

When the pool price moves outside the active range, the strategy must automatically withdraw the inactive position, rebalance token ratios, and re-enter at the new price range to restore incentive earning.

**Why this priority**: This is equally critical as the initial entry (P1) because it directly addresses the core problem: maintaining active liquidity positions to capture incentives. Without automatic rebalancing, the strategy would require manual intervention, defeating its purpose.

**Independent Test**: Can be tested by creating an out-of-range position, executing the rebalancing workflow, and verifying that the old position is withdrawn, tokens are rebalanced to 50:50 at current price, and a new position is staked in the active range.

**Acceptance Scenarios**:

1. **Given** staked position is out of range and current price is stable, **When** rebalancing executes, **Then** system unstakes position, withdraws liquidity, swaps tokens to 50:50 ratio at current price, creates new position with 10 tick width, and stakes it
2. **Given** staked position is out of range with accumulated fees, **When** rebalancing executes, **Then** system collects all earned fees before withdrawing, includes fees in rebalancing calculation
3. **Given** staked position is out of range but tokens are already at 50:50 ratio for current price, **When** rebalancing executes, **Then** system withdraws position and re-stakes at new range without executing unnecessary swaps

---

### User Story 4 - Price Stability Detection (Priority: P2)

After withdrawing an out-of-range position, the strategy must wait for price stabilization before re-entering to avoid creating a new position that immediately becomes out of range due to continued price volatility.

**Why this priority**: This optimization prevents wasteful gas consumption from repeated rebalancing during volatile periods. It's P2 because the strategy can still function without it, but user profitability improves significantly with this feature.

**Independent Test**: Can be tested by simulating price volatility after position withdrawal and verifying that the system waits to re-enter until price movements fall below a stability threshold for a minimum duration.

**Acceptance Scenarios**:

1. **Given** position has been withdrawn due to out-of-range condition, **When** price continues changing by more than 0.5% per monitoring interval, **Then** system waits without creating new position
2. **Given** position has been withdrawn and is waiting for stability, **When** price remains within 0.5% range for 5 consecutive monitoring intervals, **Then** system proceeds with position re-entry
3. **Given** position has been withdrawn during high volatility, **When** price stabilizes but then becomes volatile again before re-entry, **Then** system resets stability timer and continues waiting

---

### Edge Cases

- What happens when available token balance is insufficient to create minimum viable liquidity position after rebalancing swaps?
- How does system handle swap failures due to insufficient liquidity in the pool or excessive slippage?
- What occurs when transaction gas costs exceed expected profitability from incentives, making rebalancing uneconomical?
- How does strategy respond when staking/unstaking contracts are paused or unavailable?
- What happens if price moves out of range again while waiting for stability detection?
- How does system handle network failures or RPC unavailability during critical operations (mid-withdrawal, mid-swap)?
- What occurs when the pool itself has no liquidity or active market price cannot be determined?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST maintain token balances at 50:50 value ratio (WAVAX:USDC) before creating or recreating liquidity positions
- **FR-002**: System MUST create concentrated liquidity positions with exactly 10 tick width centered around current pool price
- **FR-003**: System MUST continuously monitor pool price at configurable intervals (default: once per minute per constitution requirement)
- **FR-004**: System MUST detect when current price exits the active liquidity range bounds (lower tick to upper tick)
- **FR-005**: System MUST automatically withdraw staked liquidity positions when price moves out of active range
- **FR-006**: System MUST collect all accumulated fees and incentives during position withdrawal
- **FR-007**: System MUST rebalance token holdings to 50:50 value ratio after withdrawal using swap operations
- **FR-008**: System MUST detect price stability before creating new liquidity positions after rebalancing, defined as price movement less than 0.5% over 5 consecutive monitoring intervals
- **FR-009**: System MUST create and stake new liquidity position only after price stability is confirmed
- **FR-010**: System MUST repeat the monitoring-rebalancing cycle continuously until explicitly stopped
- **FR-011**: System MUST enforce slippage protection on all swap operations with configurable tolerance (default: 1% per Principle 5)
- **FR-012**: System MUST validate position state after each operation (swap, mint, stake, unstake, withdraw) to ensure expected outcomes
- **FR-013**: System MUST halt operations and log error if any critical operation fails (swap revert, insufficient balance, contract unavailability)
- **FR-014**: System MUST track gas costs for all transactions executed during the strategy lifecycle
- **FR-015**: System MUST track swap fees paid during token rebalancing operations
- **FR-016**: System MUST track incentives and fees earned from liquidity positions
- **FR-017**: System MUST calculate and report net profit/loss: (incentives + fees earned) - (gas costs + swap fees paid)
- **FR-018**: System MUST persist financial records with timestamps and transaction hashes

### Constitutional Compliance

This feature MUST adhere to the following constitutional constraints:

- **Scope**: Operations limited to WAVAX/USDC pool on Blackhole DEX (Principle 1)
  - All swaps involve only WAVAX and USDC tokens
  - Liquidity positions created only in designated WAVAX/USDC pair (0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0)

- **Safety**: All operations include fail-safe error handling and rollback mechanisms (Principle 5)
  - All external calls (RPC, smart contracts) include timeout and retry logic
  - Transaction failures trigger safe termination without leaving funds in intermediate states
  - Slippage protection enforced on all swaps
  - Position validation occurs after each operation
  - Circuit breaker halts operations if error rate exceeds threshold

- **Transparency**: Financial impacts (gas, fees, profits) must be tracked and reportable (Principle 3)
  - Gas tracking records actual consumption for every transaction
  - Swap fee tracking calculates fees paid during rebalancing
  - Incentive tracking records all rewards claimed
  - Profit calculation computes net returns
  - All financial data exportable in structured format

- **Efficiency**: Gas costs must be minimized through batching and thresholds (Principle 4)
  - Rebalancing triggers use threshold logic (out-of-range detection, not minor price movements)
  - Stability detection prevents excessive rebalancing during volatility
  - Token approvals reuse existing allowances where possible
  - Gas estimation occurs before transaction submission

- **Autonomy**: Automated rebalancing without manual intervention (Principle 2)
  - Continuous pool state monitoring at configurable intervals
  - Automatic detection of out-of-range conditions
  - Automatic execution of full rebalancing workflow (unstake, collect, swap, stake)
  - Decision rationale logged for each rebalancing event

### Key Entities

- **Liquidity Position**: Represents a staked concentrated liquidity position in WAVAX/USDC pool
  - Attributes: position NFT ID, lower tick bound, upper tick bound, liquidity amount, accumulated fees, stake status
  - Lifecycle: Created → Staked → Monitored → Unstaked (if out of range) → Withdrawn → Recreated

- **Price Stability Window**: Represents the monitoring period for price stabilization detection
  - Attributes: stability threshold percentage (default 0.5%), required consecutive stable intervals (default 5), current consecutive stable count
  - Used to prevent rebalancing during volatile market conditions

- **Rebalancing Transaction**: Represents a complete rebalancing operation from detection to new position creation
  - Attributes: trigger timestamp, old position details, tokens withdrawn, swap operations performed, new position details, total gas cost, net profit/loss
  - Used for financial transparency and audit trail

- **Strategy State**: Represents the current operational state of the RunStrategy1 method
  - States: Initializing → Active Monitoring → Rebalancing Required → Waiting for Stability → Executing Rebalancing → Active Monitoring
  - Used to track strategy lifecycle and prevent concurrent operations

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Liquidity positions remain within active trading range at least 80% of the time during normal market conditions (price volatility < 10% per hour)
- **SC-002**: Strategy completes full rebalancing cycle (detect out-of-range → withdraw → rebalance → wait for stability → re-stake) in under 10 minutes during stable network conditions
- **SC-003**: Token ratio after rebalancing operations deviates from 50:50 value ratio by less than 1%
- **SC-004**: Gas costs for rebalancing operations do not exceed 2% of total position value per rebalancing event
- **SC-005**: System detects out-of-range conditions within 2 monitoring intervals (default: 2 minutes) of price exiting active range
- **SC-006**: Price stability detection prevents rebalancing during volatile periods (price movement > 0.5% per minute) at least 90% of the time
- **SC-007**: All financial records (gas, fees, incentives) are accurate within 0.1% of on-chain values
- **SC-008**: Strategy executes continuously for at least 24 hours without manual intervention or critical failures during testing
- **SC-009**: Failed operations (swap reverts, insufficient balance, network errors) trigger safe halt without fund loss in 100% of test scenarios
- **SC-010**: Net profitability (incentives earned minus all costs) exceeds passive holding strategy by at least 15% over 30-day backtesting period

## Assumptions

- Pool liquidity is sufficient for swap operations at reasonable slippage (< 1%)
- RPC endpoint provides reliable and timely access to blockchain state
- Contract interfaces (Router, PositionManager, Gauge) remain stable and available
- User maintains sufficient WAVAX balance to cover gas costs for operations
- Price stability threshold of 0.5% over 5 intervals (5 minutes default) is appropriate for most market conditions
- 10 tick width provides adequate range for capturing trading activity while maintaining capital efficiency
- Monitoring interval of 1 minute balances responsiveness with RPC cost and rate limits
- Slippage tolerance of 1% is acceptable for rebalancing swaps
- Transaction confirmations occur within 3 seconds on Avalanche C-Chain under normal conditions
