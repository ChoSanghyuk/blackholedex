# Implementation Tasks: Unstake Liquidity from Blackhole DEX

**Feature**: 001-unstake
**Branch**: `001-unstake`
**Generated**: 2025-12-19

## Overview

This document provides a task-by-task breakdown for implementing the unstake functionality. Tasks are organized by user story to enable independent implementation and testing. The feature implements withdrawal of staked liquidity positions from the FarmingCenter contract using multicall for gas efficiency.

---

## Task Summary

- **Total Tasks**: 16
- **Parallelizable Tasks**: 3
- **User Stories**: 1 (P1: Withdrawal)
- **Estimated MVP**: User Story 1 (all tasks)

---

## Phase 1: Setup & Prerequisites

**Goal**: Prepare development environment and add necessary contract constants

**Tasks**:

- [X] T001 Add FarmingCenter contract address constant to blackhole.go
  - File: `blackhole.go`
  - Action: Add `farmingCenter = "0x..."` constant (address TBD - needs mainnet deployment confirmation)
  - Dependencies: None
  - Notes: ✅ COMPLETED - farmingCenter constant already exists at line 28 with address 0xa47Ad2C95FaE476a73b85A355A5855aDb4b3A449

- [X] T002 Load and parse FarmingCenter ABI for multicall encoding
  - File: `blackhole.go` (or new `abis.go`)
  - Action: Extract FarmingCenter ABI from `blackholedex-contracts/node_modules/@cryptoalgebra/integral-farming/artifacts/contracts/FarmingCenter.sol/FarmingCenter.json`
  - Action: Parse ABI using `abi.JSON()` in init() function or package-level variable
  - Dependencies: None
  - Notes: ✅ COMPLETED - Added farmingCenterABIJson const and farmingCenterABI package var with init() function in blackhole.go

---

## Phase 2: Foundational Types

**Goal**: Add required type definitions that all user stories depend on

**Tasks**:

- [X] T003 [P] Define IncentiveKey struct in types.go
  - File: `types.go`
  - Action: Add IncentiveKey type matching Solidity struct (RewardToken, BonusRewardToken, Pool, Nonce)
  - Action: Add JSON tags for export
  - Dependencies: None
  - Notes: ✅ COMPLETED - Added IncentiveKey struct at types.go:192-197

- [X] T004 [P] Define UnstakeResult struct in types.go
  - File: `types.go`
  - Action: Add UnstakeResult type (NFTTokenID, Rewards, Transactions, TotalGasCost, Success, ErrorMessage)
  - Action: Add RewardAmounts nested type (Reward, BonusReward, RewardToken, BonusRewardToken)
  - Dependencies: T003 (IncentiveKey)
  - Notes: ✅ COMPLETED - Added RewardAmounts (types.go:200-205) and UnstakeResult (types.go:208-215)

- [X] T005 [P] Add UnstakeParams validation helper in types.go
  - File: `types.go`
  - Action: Add UnstakeParams struct with Validate() method
  - Action: Implement validation: tokenID > 0, incentiveKey non-nil, pool == wavaxUsdcPair
  - Dependencies: T003 (IncentiveKey)
  - Notes: ✅ COMPLETED - Added UnstakeParams with Validate() method at types.go:218-239

---

## Phase 3: User Story 1 - Withdrawal (Priority P1)

**User Story**: A user needs to quickly exit their position during market volatility or contract migration scenarios. They should be able to execute an unstake operation, claiming unclaimed rewards.

**Independent Test Criteria**: Can be tested independently by staking LP tokens, then executing an emergency unstake and verifying that LP tokens and rewards are returned.

**Acceptance**:
- Given a user has staked LP tokens with unclaimed rewards
- When they execute an unstake
- Then their LP tokens and reward are returned to their wallet

**Tasks**:

### 3.1 Input Validation (US1)

- [X] T006 [US1] Implement NFT token ID validation in Unstake function
  - File: `blackhole.go`
  - Action: Add Unstake() function signature: `func (b *Blackhole) Unstake(nftTokenID *big.Int, incentiveKey *IncentiveKey, collectRewards bool) (*UnstakeResult, error)`
  - Action: Validate tokenID != nil and tokenID > 0
  - Action: Return UnstakeResult with Success=false and error on validation failure
  - Dependencies: T004 (UnstakeResult type)
  - Notes: quickstart.md Step 3, contracts/unstake-api.md behavior section 1

- [X] T007 [US1] Implement IncentiveKey validation in Unstake function
  - File: `blackhole.go`
  - Action: Validate incentiveKey != nil
  - Action: Validate incentiveKey.Pool matches wavaxUsdcPair constant (constitutional Principle 1)
  - Action: Return error if pool address doesn't match WAVAX/USDC pair
  - Dependencies: T006 (Unstake function structure)
  - Notes: Constitutional compliance check, plan.md Principle 1

### 3.2 Pre-flight Checks (US1)

- [X] T008 [US1] Verify NFT ownership before unstake
  - File: `blackhole.go`
  - Action: Get NonfungiblePositionManager client via `b.Client(nonfungiblePositionManager)`
  - Action: Call `ownerOf(tokenId)` to get current owner
  - Action: Compare owner with `b.myAddr`
  - Action: Return error if not owned by caller
  - Dependencies: T007 (validation complete)
  - Notes: quickstart.md Step 3, follows Stake() pattern from blackhole.go:567-584

- [X] T009 [US1] Verify NFT farming status before unstake
  - File: `blackhole.go`
  - Action: Get FarmingCenter client via `b.Client(farmingCenter)`
  - Action: Call `deposits(tokenId)` to get current incentiveId ([32]byte)
  - Action: Verify incentiveId != zero (NFT is staked)
  - Action: Return error if NFT not currently in farming
  - Dependencies: T008 (ownership verified)
  - Notes: data-model.md FarmingStatus entity, quickstart.md Step 3

### 3.3 Multicall Construction (US1)

- [X] T010 [US1] Encode exitFarming call for multicall
  - File: `blackhole.go`
  - Action: Use farmingCenterABI.Pack("exitFarming", incentiveKey, nftTokenID)
  - Action: Append encoded bytes to multicallData array
  - Action: Handle encoding errors with descriptive error messages
  - Dependencies: T002 (ABI loaded), T009 (pre-flight checks complete)
  - Notes: research.md Task 3, quickstart.md Step 5

- [X] T011 [US1] Conditionally encode collectRewards call for multicall
  - File: `blackhole.go`
  - Action: If collectRewards==true, encode collectRewards(incentiveKey, nftTokenID)
  - Action: Append to multicallData array after exitFarming
  - Action: Handle encoding errors
  - Dependencies: T010 (exitFarming encoded)
  - Notes: research.md Task 4, allows optional reward collection

### 3.4 Transaction Execution (US1)

- [X] T012 [US1] Execute multicall transaction to FarmingCenter
  - File: `blackhole.go`
  - Action: Call farmingCenterClient.Send(types.Standard, nil, &b.myAddr, b.privateKey, "multicall", multicallData)
  - Action: Use nil for gas limit (automatic estimation per Principle 4)
  - Action: Log transaction submission with NFT tokenID and FarmingCenter address
  - Dependencies: T011 (multicall data prepared)
  - Notes: quickstart.md Step 6, follows Stake() pattern

- [X] T013 [US1] Wait for transaction confirmation and extract gas cost
  - File: `blackhole.go`
  - Action: Call `b.tl.WaitForTransaction(multicallTxHash)` to get receipt
  - Action: Use `util.ExtractGasCost(receipt)` to get gas cost
  - Action: Parse EffectiveGasPrice and GasUsed from receipt
  - Action: Create TransactionRecord with operation name (e.g., "ExitFarmingWithRewards")
  - Dependencies: T012 (transaction submitted)
  - Notes: Follows existing pattern from Stake() blackhole.go:698-711, Principle 3 compliance

### 3.5 Result Construction (US1)

- [X] T014 [US1] Parse reward amounts from multicall results (if collected)
  - File: `blackhole.go`
  - Action: If collectRewards==true, parse multicall return data for reward amounts
  - Action: Extract reward and bonusReward values (both *big.Int)
  - Action: Populate RewardAmounts struct with reward token addresses from incentiveKey
  - Dependencies: T013 (transaction confirmed)
  - Notes: quickstart.md Step 4 helper function, research.md indicates parsing needed

- [X] T015 [US1] Construct and return UnstakeResult with all transaction details
  - File: `blackhole.go`
  - Action: Calculate totalGasCost by summing all transaction records
  - Action: Populate UnstakeResult: NFTTokenID, Rewards (if collected), Transactions, TotalGasCost, Success=true
  - Action: Log success message with NFT ID, gas cost, and reward amounts
  - Dependencies: T014 (rewards parsed)
  - Notes: Follows Stake() result pattern blackhole.go:744-771, Principle 3 compliance

---

## Phase 4: Polish & Error Handling

**Goal**: Add comprehensive error handling and logging

**Tasks**:

- [X] T016 Add comprehensive error handling for all failure scenarios
  - File: `blackhole.go`
  - Action: Ensure all error paths return UnstakeResult with Success=false and descriptive ErrorMessage
  - Action: Track partial gas costs even on failure (for transparency)
  - Action: Add troubleshooting context to error messages (e.g., "NFT not owned by wallet: owned by 0x...")
  - Dependencies: T015 (basic implementation complete)
  - Notes: contracts/unstake-api.md Error Conditions table, Principle 5 compliance

---

## Dependencies & Execution Order

### User Story Completion Order

```
User Story 1 (P1): Withdrawal
   No dependencies (can be implemented first)
```

### Task Dependencies Graph

```
T001 (FarmingCenter constant)  
T002 (ABI loading)             $
T003 (IncentiveKey type)       < � T006 (Input validation)   � T007 (Pool validation)   � T008 (Ownership check)
T004 (UnstakeResult type)                                                                     
T005 (Validation helper)                                                                       
                                                                                                �
                                                                                          T009 (Farming status)
                                                                                                
                                                                                                �
                                                                                          T010 (Encode exitFarming)
                                                                                                
                                                                                                �
                                                                                          T011 (Encode collectRewards)
                                                                                                
                                                                                                �
                                                                                          T012 (Execute multicall)
                                                                                                
                                                                                                �
                                                                                          T013 (Confirm & extract gas)
                                                                                                
                                                                                                �
                                                                                          T014 (Parse rewards)
                                                                                                
                                                                                                �
                                                                                          T015 (Return result)
                                                                                                
                                                                                                �
                                                                                          T016 (Error handling polish)
```

### Parallel Execution Opportunities

**Phase 2 (Foundational Types)** - Can work in parallel:
- T003 (IncentiveKey type) [P]
- T004 (UnstakeResult type) [P]
- T005 (Validation helper) [P]

All three tasks modify different sections of `types.go` and have no dependencies on each other.

---

## Implementation Strategy

### MVP Scope (Minimum Viable Product)

**Target**: User Story 1 (P1: Withdrawal) - Tasks T001-T016

This provides a complete, independently testable unstake feature:
- Add FarmingCenter contract integration
- Implement Unstake() function with full validation
- Support optional reward collection via multicall
- Track all gas costs and rewards (constitutional compliance)
- Comprehensive error handling

**MVP Deliverables**:
1. Working Unstake() function in blackhole.go
2. All required types in types.go
3. FarmingCenter ABI integration
4. Constitutional compliance (all 5 principles satisfied)

### Testing Approach

**Unit Testing** (per quickstart.md):
- Test invalid token ID rejection (nil, zero, negative)
- Test invalid pool address rejection
- Test NFT ownership validation
- Test farming status verification
- Mock ContractClient for isolated testing

**Integration Testing**:
- Stake LP tokens using existing Mint/Stake functions
- Execute Unstake with collectRewards=true
- Verify LP tokens returned to wallet
- Verify rewards claimed and received
- Validate gas tracking accuracy
- Test error scenarios (not owner, not staked, wrong pool)

### Incremental Delivery

1. **Phase 1-2 Complete**: Types and setup ready (T001-T005)
2. **Phase 3.1-3.2 Complete**: Validation and pre-flight checks working (T006-T009)
3. **Phase 3.3 Complete**: Multicall encoding functional (T010-T011)
4. **Phase 3.4 Complete**: Transaction execution and confirmation (T012-T013)
5. **Phase 3.5 Complete**: Result parsing and return (T014-T015)
6. **Phase 4 Complete**: Error handling polished (T016)

Each increment can be tested in isolation before proceeding to the next.

---

## Constitutional Compliance Mapping

All tasks enforce constitutional principles:

- **Principle 1 (Pool-Specific Scope)**: T007 validates pool address
- **Principle 3 (Financial Transparency)**: T013, T015 track gas costs; T014, T015 track rewards
- **Principle 4 (Gas Optimization)**: T010-T012 use multicall batching; T012 uses auto gas estimation
- **Principle 5 (Fail-Safe Operation)**: T008-T009 pre-flight validation; T016 comprehensive error handling

---

## Success Metrics

Aligned with spec.md Success Criteria:

- **SC-001**: Task T012-T013 ensures transaction completion within one confirmation
- **SC-002**: Task T008-T009 validates requests before execution
- **SC-003**: Task T010-T012 uses multicall for gas efficiency (target: <150% of standard ERC20 transfer)
- **SC-004**: Task T016 provides clear error messages for all failures
- **SC-005**: On-chain FarmingCenter contract maintains balance accuracy
- **SC-006**: Task T013, T015 provide transaction receipts for verification

---

## References

- [Specification](./spec.md) - User stories and requirements
- [Plan](./plan.md) - Technical context and constitutional compliance
- [Data Model](./data-model.md) - Entity definitions and relationships
- [API Contract](./contracts/unstake-api.md) - Function signature and behavior
- [Research](./research.md) - Technical decisions and patterns
- [Quickstart](./quickstart.md) - Implementation guide and examples
- [Existing Stake Implementation](../../blackhole.go:543-772) - Pattern reference

---

**Task Generation Complete** - Ready for implementation. Start with Phase 1-2 (setup and types), then proceed through User Story 1 tasks in order.
