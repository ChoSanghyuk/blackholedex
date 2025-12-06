# CLAUDE.md - Blackhole DEX Go Client Development Guide

## Project Overview

**Project Name**: blackholego
**Purpose**: Go server for reading/writing transactions to Blackhole DEX on Avalanche



## Project Structure

```
blackhole_dex/
├── blackholedex-contracts/     # Solidity smart contracts (reference)
│   └── contracts/              # All DEX contracts (Pair, Router, VotingEscrow, Gauge, etc.)
│   └── artifacts/              # contracts build files (Generated ABIs)
├── internal/             			
│   └── util/         					# Generic contract client (ABI pack/unpack, tx send)
├── pkg/             						
│   └── contractclient/         # Generic contract client (ABI pack/unpack, tx send)
├── cmd/                        # CLI entry point
└── specs/                      # Development specifications
blackhole.go 	          				# High-level BlackholeDEX-specific operations
blackhole_interfaces.go 	      # Interfaces blockhole structure would use
types.go 	          						# Parameter types used in blackhole.go
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

