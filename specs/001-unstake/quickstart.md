# Quickstart: Unstake Feature Implementation

**Feature**: 001-unstake
**Created**: 2025-12-18
**Audience**: Developers implementing the unstake functionality

## Overview

This guide provides step-by-step instructions for implementing the unstake feature in the Blackhole DEX Go client. Follow this guide after reviewing the spec, research, and data model documents.

---

## Prerequisites

Before starting implementation:

1. **Read required documents** (in order):
   - [ ] `spec.md` - Feature requirements and user stories
   - [ ] `research.md` - Technical decisions and patterns
   - [ ] `data-model.md` - Data structures and validation rules
   - [ ] `contracts/unstake-api.md` - API contract specification

2. **Environment setup**:
   - [ ] Go 1.24.10 installed
   - [ ] Access to Avalanche C-Chain RPC endpoint
   - [ ] Wallet with AVAX for gas fees (testnet or mainnet)
   - [ ] NFT token ID from previous Mint operation
   - [ ] Knowledge of IncentiveKey for the farming program

3. **Codebase familiarity**:
   - [ ] Review `blackhole.go` Stake function (lines 543-772)
   - [ ] Understand ContractClient interface usage
   - [ ] Understand TransactionRecord and StakingResult patterns

---

## Implementation Steps

### Step 1: Add Type Definitions to types.go

Add the IncentiveKey struct and related types:

```go
// File: types.go

// IncentiveKey identifies a farming incentive program in FarmingCenter
type IncentiveKey struct {
    RewardToken      common.Address `json:"rewardToken"`
    BonusRewardToken common.Address `json:"bonusRewardToken"`
    Pool             common.Address `json:"pool"`
    Nonce            *big.Int       `json:"nonce"`
}

// UnstakeParams contains parameters for unstaking an NFT position
type UnstakeParams struct {
    NFTTokenID     *big.Int      `json:"nftTokenId"`
    IncentiveKey   *IncentiveKey `json:"incentiveKey"`
    CollectRewards bool          `json:"collectRewards"`
}

// Validate checks if UnstakeParams are valid
func (p *UnstakeParams) Validate() error {
    if p.NFTTokenID == nil || p.NFTTokenID.Sign() <= 0 {
        return errors.New("NFT token ID must be positive")
    }
    if p.IncentiveKey == nil {
        return errors.New("IncentiveKey is required")
    }
    if p.IncentiveKey.Pool == (common.Address{}) {
        return errors.New("Pool address cannot be zero")
    }
    // Constitutional constraint: must be WAVAX/USDC pool
    if p.IncentiveKey.Pool.Hex() != wavaxUsdcPair {
        return fmt.Errorf("pool must be WAVAX/USDC pair: %s", wavaxUsdcPair)
    }
    if p.IncentiveKey.Nonce == nil {
        return errors.New("Nonce cannot be nil")
    }
    return nil
}
```

**Testing checkpoint**: Compile the project to verify types are valid.

---

### Step 2: Load FarmingCenter ABI

The FarmingCenter ABI is needed for encoding multicall data. You'll need to extract it from the compiled contract.

**Option A: Use pre-compiled ABI** (recommended)
```bash
# Navigate to contracts directory
cd blackholedex-contracts

# Compile if not already done
npx hardhat compile

# Extract FarmingCenter ABI
cat node_modules/@cryptoalgebra/integral-farming/artifacts/contracts/FarmingCenter.sol/FarmingCenter.json | jq '.abi' > ../farmingcenter-abi.json
```

**Option B: Embed ABI in Go code**
```go
// File: blackhole.go or new file abis.go

const FarmingCenterABI = `[
  {
    "inputs": [
      {
        "components": [
          {"internalType": "address", "name": "rewardToken", "type": "address"},
          {"internalType": "address", "name": "bonusRewardToken", "type": "address"},
          {"internalType": "address", "name": "pool", "type": "address"},
          {"internalType": "uint256", "name": "nonce", "type": "uint256"}
        ],
        "internalType": "struct IFarmingCenter.IncentiveKey",
        "name": "key",
        "type": "tuple"
      },
      {"internalType": "uint256", "name": "tokenId", "type": "uint256"}
    ],
    "name": "exitFarming",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "components": [
          {"internalType": "address", "name": "rewardToken", "type": "address"},
          {"internalType": "address", "name": "bonusRewardToken", "type": "address"},
          {"internalType": "address", "name": "pool", "type": "address"},
          {"internalType": "uint256", "name": "nonce", "type": "uint256"}
        ],
        "internalType": "struct IFarmingCenter.IncentiveKey",
        "name": "key",
        "type": "tuple"
      },
      {"internalType": "uint256", "name": "tokenId", "type": "uint256"}
    ],
    "name": "collectRewards",
    "outputs": [
      {"internalType": "uint256", "name": "reward", "type": "uint256"},
      {"internalType": "uint256", "name": "bonusReward", "type": "uint256"}
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {"internalType": "bytes[]", "name": "data", "type": "bytes[]"}
    ],
    "name": "multicall",
    "outputs": [
      {"internalType": "bytes[]", "name": "results", "type": "bytes[]"}
    ],
    "stateMutability": "payable",
    "type": "function"
  }
]`

// Parse ABI once at package level or in init()
var farmingCenterABI abi.ABI

func init() {
    var err error
    farmingCenterABI, err = abi.JSON(strings.NewReader(FarmingCenterABI))
    if err != nil {
        log.Fatalf("Failed to parse FarmingCenter ABI: %v", err)
    }
}
```

**Testing checkpoint**: Verify ABI parses without errors.

---

### Step 3: Implement Unstake Function

Add the Unstake function to `blackhole.go`:

```go
// File: blackhole.go

// Unstake withdraws a staked NFT position from FarmingCenter
// nftTokenID: ERC721 token ID from previous Mint operation
// incentiveKey: Identifies the farming program to exit
// collectRewards: Whether to claim accumulated rewards during unstake
// Returns StakingResult with transaction tracking and gas costs
func (b *Blackhole) Unstake(
    nftTokenID *big.Int,
    incentiveKey *IncentiveKey,
    collectRewards bool,
) (*StakingResult, error) {
    // Step 1: Input validation
    if nftTokenID == nil || nftTokenID.Sign() <= 0 {
        return &StakingResult{
            Success:      false,
            ErrorMessage: "validation failed: invalid token ID",
        }, fmt.Errorf("validation failed: invalid token ID")
    }

    if incentiveKey == nil {
        return &StakingResult{
            Success:      false,
            ErrorMessage: "validation failed: incentiveKey is required",
        }, fmt.Errorf("validation failed: incentiveKey is required")
    }

    // Constitutional check: must be WAVAX/USDC pool
    if incentiveKey.Pool.Hex() != wavaxUsdcPair {
        return &StakingResult{
            Success:      false,
            ErrorMessage: fmt.Sprintf("pool must be WAVAX/USDC: %s", wavaxUsdcPair),
        }, fmt.Errorf("invalid pool address")
    }

    // Step 2: Initialize transaction tracking
    var transactions []TransactionRecord

    // Step 3: Verify NFT ownership
    nftManagerClient, err := b.Client(nonfungiblePositionManager)
    if err != nil {
        return &StakingResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to get NFT manager client: %v", err),
        }, fmt.Errorf("failed to get NFT manager client: %w", err)
    }

    ownerResult, err := nftManagerClient.Call(&b.myAddr, "ownerOf", nftTokenID)
    if err != nil {
        return &StakingResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to verify NFT ownership: %v", err),
        }, fmt.Errorf("failed to verify NFT ownership: %w", err)
    }

    owner := ownerResult[0].(common.Address)
    if owner != b.myAddr {
        return &StakingResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("NFT not owned by wallet: owned by %s", owner.Hex()),
        }, fmt.Errorf("NFT not owned by wallet")
    }

    // Step 4: Verify NFT is currently farmed
    farmingCenterClient, err := b.Client(farmingcenter)
    if err != nil {
        return &StakingResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to get FarmingCenter client: %v", err),
        }, fmt.Errorf("failed to get FarmingCenter client: %w", err)
    }

    depositsResult, err := farmingCenterClient.Call(&b.myAddr, "deposits", nftTokenID)
    if err != nil {
        return &StakingResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to check farming status: %v", err),
        }, fmt.Errorf("failed to check farming status: %w", err)
    }

    currentIncentiveId := depositsResult[0].([32]byte)
    if currentIncentiveId == [32]byte{} {
        return &StakingResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: "NFT is not currently staked in farming",
        }, fmt.Errorf("NFT is not currently staked")
    }

    // Step 5: Build multicall data
    var multicallData [][]byte

    // Encode exitFarming call
    exitFarmingData, err := farmingCenterABI.Pack("exitFarming", incentiveKey, nftTokenID)
    if err != nil {
        return &StakingResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to encode exitFarming: %v", err),
        }, fmt.Errorf("failed to encode exitFarming: %w", err)
    }
    multicallData = append(multicallData, exitFarmingData)

    // Optionally encode collectRewards call
    if collectRewards {
        collectRewardsData, err := farmingCenterABI.Pack("collectRewards", incentiveKey, nftTokenID)
        if err != nil {
            return &StakingResult{
                NFTTokenID:   nftTokenID,
                Success:      false,
                ErrorMessage: fmt.Sprintf("failed to encode collectRewards: %v", err),
            }, fmt.Errorf("failed to encode collectRewards: %w", err)
        }
        multicallData = append(multicallData, collectRewardsData)
    }

    // Step 6: Execute multicall transaction
    log.Printf("Unstaking NFT %s from FarmingCenter %s", nftTokenID.String(), farmingcenter)

    multicallTxHash, err := farmingCenterClient.Send(
        types.Standard,
        nil, // Use automatic gas limit estimation
        &b.myAddr,
        b.privateKey,
        "multicall",
        multicallData,
    )
    if err != nil {
        return &StakingResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to submit multicall transaction: %v", err),
        }, fmt.Errorf("failed to submit multicall transaction: %w", err)
    }

    // Step 7: Wait for transaction confirmation
    multicallReceipt, err := b.tl.WaitForTransaction(multicallTxHash)
    if err != nil {
        return &StakingResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("multicall transaction failed: %v", err),
        }, fmt.Errorf("multicall transaction failed: %w", err)
    }

    // Step 8: Extract gas cost
    gasCost, err := util.ExtractGasCost(multicallReceipt)
    if err != nil {
        return &StakingResult{
            NFTTokenID:   nftTokenID,
            Success:      false,
            ErrorMessage: fmt.Sprintf("failed to extract gas cost: %v", err),
        }, fmt.Errorf("failed to extract gas cost: %w", err)
    }

    gasPrice := new(big.Int)
    gasPrice.SetString(multicallReceipt.EffectiveGasPrice, 0)
    gasUsed := new(big.Int)
    gasUsed.SetString(multicallReceipt.GasUsed, 0)

    operationName := "ExitFarming"
    if collectRewards {
        operationName = "ExitFarmingWithRewards"
    }

    transactions = append(transactions, TransactionRecord{
        TxHash:    multicallTxHash,
        GasUsed:   gasUsed.Uint64(),
        GasPrice:  gasPrice,
        GasCost:   gasCost,
        Timestamp: time.Now(),
        Operation: operationName,
    })

    // Step 9: Parse rewards if collected
    rewardAmount := big.NewInt(0)
    bonusRewardAmount := big.NewInt(0)

    if collectRewards {
        // Parse multicall results to extract reward amounts
        // This requires parsing the receipt logs or return data
        // For now, we'll set to 0 - actual implementation would parse events
        // TODO: Parse collectRewards return values from multicall results
        log.Printf("Rewards collected (parsing not yet implemented)")
    }

    // Step 10: Construct result
    totalGasCost := big.NewInt(0)
    for _, tx := range transactions {
        totalGasCost.Add(totalGasCost, tx.GasCost)
    }

    result := &StakingResult{
        NFTTokenID:     nftTokenID,
        ActualAmount0:  rewardAmount,
        ActualAmount1:  bonusRewardAmount,
        FinalTickLower: 0, // Not applicable for unstake
        FinalTickUpper: 0, // Not applicable for unstake
        Transactions:   transactions,
        TotalGasCost:   totalGasCost,
        Success:        true,
        ErrorMessage:   "",
    }

    // Step 11: Logging
    fmt.Printf("✓ NFT unstaked successfully\n")
    fmt.Printf("  Token ID: %s\n", nftTokenID.String())
    fmt.Printf("  FarmingCenter: %s\n", farmingcenter)
    if collectRewards {
        fmt.Printf("  Rewards: %s / %s\n", rewardAmount.String(), bonusRewardAmount.String())
    }
    fmt.Printf("  Total Gas Cost: %s wei\n", totalGasCost.String())
    for _, tx := range transactions {
        fmt.Printf("  - %s: %s (gas: %s wei)\n", tx.Operation, tx.TxHash.Hex(), tx.GasCost.String())
    }

    return result, nil
}
```

**Testing checkpoint**: Compile and verify function signature matches API contract.

---

### Step 4: Add Helper Functions (Optional)

For parsing reward amounts from multicall results:

```go
// parseMulticallRewards extracts reward amounts from multicall return data
func parseMulticallRewards(multicallResults [][]byte, farmingCenterABI abi.ABI) (*big.Int, *big.Int, error) {
    if len(multicallResults) < 2 {
        return big.NewInt(0), big.NewInt(0), nil // No rewards collected
    }

    // Second element contains collectRewards return data
    collectRewardsResult := multicallResults[1]

    // Unpack the result
    var results []interface{}
    err := farmingCenterABI.UnpackIntoInterface(&results, "collectRewards", collectRewardsResult)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to unpack collectRewards result: %w", err)
    }

    if len(results) != 2 {
        return nil, nil, fmt.Errorf("unexpected collectRewards result length: %d", len(results))
    }

    reward := results[0].(*big.Int)
    bonusReward := results[1].(*big.Int)

    return reward, bonusReward, nil
}
```

---

### Step 5: Write Unit Tests

Create test file `blackhole_unstake_test.go`:

```go
package blackholedex

import (
    "math/big"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestUnstake_InvalidTokenID(t *testing.T) {
    b := &Blackhole{} // Minimal setup

    incentiveKey := &IncentiveKey{
        RewardToken: common.HexToAddress("0xcd94a87696fac69edae3a70fe5725307ae1c43f6"),
        Pool:        common.HexToAddress(wavaxUsdcPair),
        Nonce:       big.NewInt(1),
    }

    // Test nil token ID
    result, err := b.Unstake(nil, incentiveKey, false)
    assert.Error(t, err)
    assert.False(t, result.Success)
    assert.Contains(t, result.ErrorMessage, "invalid token ID")

    // Test zero token ID
    result, err = b.Unstake(big.NewInt(0), incentiveKey, false)
    assert.Error(t, err)
    assert.False(t, result.Success)
}

func TestUnstake_InvalidPool(t *testing.T) {
    b := &Blackhole{}

    invalidIncentiveKey := &IncentiveKey{
        RewardToken: common.HexToAddress("0xcd94a87696fac69edae3a70fe5725307ae1c43f6"),
        Pool:        common.HexToAddress("0x0000000000000000000000000000000000000001"), // Wrong pool
        Nonce:       big.NewInt(1),
    }

    result, err := b.Unstake(big.NewInt(123), invalidIncentiveKey, false)
    assert.Error(t, err)
    assert.False(t, result.Success)
    assert.Contains(t, result.ErrorMessage, "WAVAX/USDC")
}

// TODO: Add integration tests with mock ContractClient
```

---

## Usage Guide

### Step 1: Initialize Blackhole Manager

```go
import (
    "blackholego"
    "github.com/ethereum/go-ethereum/crypto"
)

// Load private key
privateKey, err := crypto.HexToECDSA("your-private-key-hex")
if err != nil {
    log.Fatal(err)
}

// Create Blackhole manager
bh := blackholedex.NewBlackhole(privateKey, rpcURL)
```

### Step 2: Get NFT Token ID and IncentiveKey

```go
// From previous Stake operation
nftTokenID := big.NewInt(12345) // Your NFT token ID

// Query FarmingCenter to get current incentiveKey
// Or use known values:
incentiveKey := &blackholedex.IncentiveKey{
    RewardToken:      common.HexToAddress("0xcd94a87696fac69edae3a70fe5725307ae1c43f6"),
    BonusRewardToken: common.Address{}, // Zero if no bonus
    Pool:             common.HexToAddress("0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0"),
    Nonce:            big.NewInt(1),
}
```

### Step 3: Execute Unstake

```go
// Option A: Unstake without collecting rewards
result, err := bh.Unstake(nftTokenID, incentiveKey, false)
if err != nil {
    log.Fatalf("Unstake failed: %v", err)
}

// Option B: Unstake and collect rewards
result, err := bh.Unstake(nftTokenID, incentiveKey, true)
if err != nil {
    log.Fatalf("Unstake failed: %v", err)
}

// Check result
if !result.Success {
    log.Fatalf("Unstake unsuccessful: %s", result.ErrorMessage)
}

fmt.Printf("Successfully unstaked NFT %s\n", result.NFTTokenID.String())
fmt.Printf("Gas cost: %s wei\n", result.TotalGasCost.String())
```

---

## Troubleshooting

### Common Issues

**Issue**: "NFT not owned by wallet"
- **Cause**: Private key doesn't match NFT owner
- **Solution**: Verify you're using the correct wallet/private key

**Issue**: "NFT is not currently staked in farming"
- **Cause**: NFT was already unstaked or never staked
- **Solution**: Check FarmingCenter.deposits(tokenId) on-chain

**Issue**: "invalid incentiveId"
- **Cause**: IncentiveKey doesn't match the active farming program
- **Solution**: Query FarmingCenter.deposits(tokenId) and FarmingCenter.incentiveKeys(incentiveId) to get correct values

**Issue**: "failed to encode exitFarming"
- **Cause**: FarmingCenter ABI not loaded or malformed
- **Solution**: Verify ABI JSON is correct and complete

---

## Next Steps

After implementing the unstake feature:

1. Run `/speckit.tasks` to generate implementation tasks
2. Follow task-by-task development workflow
3. Write comprehensive tests (unit + integration)
4. Test on Avalanche testnet before mainnet deployment
5. Document any edge cases discovered during testing

---

## References

- [Spec](./spec.md)
- [Research](./research.md)
- [Data Model](./data-model.md)
- [API Contract](./contracts/unstake-api.md)
- [FarmingCenter Contract](../../blackholedex-contracts/node_modules/@cryptoalgebra/integral-farming/contracts/FarmingCenter.sol)
- [Stake Function](../../blackhole.go) (lines 543-772)
