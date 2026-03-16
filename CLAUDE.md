# CLAUDE.md - Blackhole DEX Go Client Development Guide

## Project Overview

**Project Name**: blackholego
**Purpose**: Go server for reading/writing transactions to Blackhole DEX on Avalanche



## Project Structure

```
blackhole_dex/
├── cmd/                                    # CLI entry point
│   └── main.go                            # Application entry: main(), config loading, strategy execution
│
├── pkg/
│   ├── contractclient/                    # Generic EVM contract interaction layer
│   │   ├── contractclient.go             # Call(), Send(), SendWithValue(), GetReceipt(), ParseReceipt(),
│   │   │                                  # DecodeTransaction(), DecodeByHash()
│   │   └── contractclient_test.go        # Contract client tests
│   │
│   ├── types/                             # Type definitions & enums
│   │   ├── contract_params.go            # MintParams, StakeParams, UnstakeParams, WithdrawParams
│   │   ├── enums.go                       # PoolType, StrategyPhase, StrategyStep
│   │   ├── operation_results.go          # MintResult, StakeResult, UnstakeResult, WithdrawResult
│   │   ├── pool_types.go                 # PoolType methods: PoolNonce(), TickSpacing()
│   │   ├── strategy_types.go             # StrategyConfig, StrategyReport, PositionRange, StabilityWindow,
│   │   │                                  # CircuitBreaker, AMMState, Position, CurrentAssetSnapshot
│   │   ├── transaction.go                # TxReceipt, DecodedTransaction, Priority
│   │   └── priority.go                   # Gas priority definitions
│   │
│   └── util/                              # Utility functions
│       ├── abi_loader.go                 # LoadABIFromHardhatArtifact(), LoadABI(), GetContractInfo()
│       ├── amm.go                         # TickToSqrtPriceX96(), ComputeAmounts(), CalculateTokenAmountsFromLiquidity()
│       ├── calculations.go               # SqrtPriceToPrice(), CalculateRebalanceAmounts()
│       ├── validation.go                 # ValidateStakingRequest(), CalculateTickBounds(), CalculateMinAmount(),
│       │                                  # ExtractGasCost(), IsCriticalError()
│       ├── crypt.go                       # Encrypt(), Decrypt()
│       └── hex.go                         # Hex2Bytes()
│
├── internal/
│   └── db/                                # Database persistence layer
│       ├── transaction_recorder.go       # NewMySQLRecorder(), RecordReport(), GetLatestSnapshot(),
│       │                                  # GetSnapshotsByTimeRange(), GetSnapshotsByPhase()
│       └── transaction_recorder_test.go  # DB tests
│
├── configs/
│   └── config.go                          # LoadConfig(), ToBlackholeConfigs(), ToStrategyConfig()
│
├── Root-level core files:
│   ├── blackhole.go                      # Main Blackhole struct: NewBlackhole(), RunAutoPositionStrategy()
│   ├── blackhole_interfaces.go           # Interfaces: ContractClient, TxListener, TransactionRecorder
│   ├── position.go                        # Position operations: Mint(), Stake(), Unstake(), Withdraw()
│   ├── query.go                           # Query operations: GetAMMState(), GetPositionDetails(), GetUserPositions(),
│   │                                      # TokenOfOwnerByIndex(), validateBalances()
│   ├── token.go                           # Token operations: Swap(), ensureApproval()
│   ├── portfolio.go                       # Portfolio tracking & alarming: RecordCurrentAssetSnapshot(), GetCurrentAssetSnapshot(), sendReport()
│   └── contract_registry.go              # Contract registry: NewContractRegistry(), Client(), ClientByAddress()
│
├── blackholedex-contracts/               # Solidity smart contracts (reference)
│   ├── contracts/                        # DEX contracts (Pair, Router, VotingEscrow, Gauge, etc.)
│   └── artifacts/                        # Generated ABIs from Hardhat compilation
│
└── specs/                                # Development specifications & strategy implementations
    └── 001-liquidity-repositioning/
        └── contracts/
            └── strategy_api.go           # Legacy strategy types (migrated to pkg/types/)
```



## Core Components

### Existing: ContractClient (`pkg/contractclient/`)
Generic EVM contract interaction:
- `Call()` - Read-only contract calls
- `Send()` / `SendWithValue()` - State-changing transactions
- `GetReceipt()` / `ParseReceipt()` - Transaction receipt and event parsing

### To Build: BlackholeManager (`blackholemanager.go`)
High-level interface for Blackhole DEX operations:
```go
type BlackholeManager struct {
    // Clients for each contract type
    wavax       *ContractClient
    usdc  			*ContractClient
    lp 					*ContractClient
    // ... other contracts
}
```

## Contract Reference

### Priority Contracts for Manager Implementation

| Contract | Address | Key Functions |
|----------|---------|---------------|
| RouterV2 | {todo: mainnet address} | `swapExactTokensForTokens`, `addLiquidity`, `removeLiquidity` |
| PairFactory | {todo: mainnet address} | `getPair`, `allPairs`, `createPair` |
| VotingEscrow | {todo: mainnet address} | `create_lock`, `increase_amount`, `withdraw` |
| VoterV3 | {todo: mainnet address} | `vote`, `claimBribes`, `claimFees` |
| GaugeV2 | {todo: varies by pool} | `deposit`, `withdraw`, `getReward` |

### Contract ABIs Location
After build: `blackholedex-contracts/artifacts/contracts/{ContractName}.sol/{ContractName}.json`

{todo: Run `npx hardhat compile` in blackholedex-contracts/ to generate ABIs}

## Coding Standards

### Error Handling
```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to decode %s: %w", methodName, err)
}
```

### Naming Conventions
- Decoder methods: `Decode{Action}` (e.g., `DecodeSwap`, `DecodeAddLiquidity`)
- Manager methods: `{Action}` (e.g., `Swap`, `AddLiquidity`, `CreateLock`)
- Builder methods: `Build{Action}Data` (e.g., `BuildSwapData`)

### Type Conversions
```go
// Always use go-ethereum types
amount := new(big.Int).SetUint64(1000000)
address := common.HexToAddress("0x...")
hash := common.HexToHash("0x...")
```

### Testing
- Use table-driven tests
- Include real mainnet transaction hashes for decoder tests
- Mock RPC for manager tests where possible



## Notes

- Always use `context.Context` for cancellation support in network calls
- Consider rate limiting for RPC calls
- Cache ABI parsing results for performance
- Use `sync.Once` for singleton initialization

## Contact & Resources

- Blackhole DEX Docs: {todo: documentation URL}
- Avalanche RPC Providers: {todo: recommended providers}
- Contract Verification: {todo: explorer links}


## Active Technologies
- Go 1.24.10 + github.com/ethereum/go-ethereum v1.16.7, existing internal packages (util, contractclient) (001-liquidity-staking)
- N/A (blockchain state only, no local persistence) (001-liquidity-staking)

## Recent Changes
- 003-liquidity-withdraw: Added Withdraw function with multicall for atomic position exit and NFT burn (Go 1.24.10, go-ethereum v1.16.7)
- 001-liquidity-staking: Added Go 1.24.10 + github.com/ethereum/go-ethereum v1.16.7, existing internal packages (util, contractclient)
