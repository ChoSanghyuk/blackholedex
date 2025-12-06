<!--
Sync Impact Report:
- Version change: (new constitution) → 1.0.0
- Modified principles: N/A (new constitution)
- Added sections:
  * Core Principles (5 principles: Code Quality, Testing Discipline, User Experience Consistency, Performance Requirements, Security & Error Handling)
  * Development Standards
  * Quality Gates
  * Governance
- Removed sections: N/A
- Templates requiring updates:
  ✅ plan-template.md: Constitution Check section aligns with principles
  ✅ spec-template.md: Requirements align with UX consistency and testing principles
  ✅ tasks-template.md: Task organization reflects testing discipline and quality gates
- Follow-up TODOs: None - all placeholders filled
-->

# Blackholego Project Constitution

## Core Principles

### I. Code Quality & Maintainability

**Rule**: All code MUST adhere to Go best practices and be self-documenting through clear naming and structure.

- Type conversions MUST use go-ethereum types (`*big.Int`, `common.Address`, `common.Hash`)
- Naming conventions are MANDATORY:
  - Decoder methods: `Decode{Action}` (e.g., `DecodeSwap`)
  - Manager methods: `{Action}` (e.g., `Swap`, `AddLiquidity`)
  - Builder methods: `Build{Action}Data` (e.g., `BuildSwapData`)
- Error handling MUST wrap errors with context: `fmt.Errorf("failed to decode %s: %w", methodName, err)`
- ABI parsing results MUST be cached for performance
- Singleton initialization MUST use `sync.Once`
- Comments are ONLY required where logic is not self-evident
- No premature abstractions—three similar lines of code is better than unnecessary complexity

**Rationale**: Clear, idiomatic Go code reduces cognitive load, enables rapid onboarding, and prevents bugs through consistency. Explicit naming conventions ensure codebase-wide uniformity for DEX operations.

### II. Testing Discipline (NON-NEGOTIABLE)

**Rule**: Testing strategy MUST match the component type with appropriate coverage.

- Table-driven tests are MANDATORY for all business logic
- Contract decoder tests MUST include real mainnet transaction hashes from Avalanche Blackhole DEX
- Manager tests MUST mock RPC calls where possible to avoid network dependencies
- Integration tests MUST cover:
  - Contract interactions (swap, liquidity, staking)
  - Transaction receipt parsing
  - Event decoding
- Unit tests MUST validate:
  - ABI pack/unpack correctness
  - Parameter type conversions
  - Error handling paths
- Tests MUST be written BEFORE implementation when using TDD workflow (optional but encouraged)
- All tests MUST pass before commits

**Rationale**: Blockchain interactions are irreversible and expensive. Comprehensive testing with real transaction data ensures reliability. Table-driven tests enable exhaustive scenario coverage with minimal code duplication.

### III. User Experience Consistency

**Rule**: All user-facing interfaces (CLI, API, logs) MUST provide clear, actionable feedback.

- Network operations MUST use `context.Context` for cancellation support
- RPC calls MUST implement rate limiting to prevent provider throttling
- Error messages MUST include:
  - Clear description of what failed
  - Context (transaction hash, contract address, method name)
  - Actionable guidance where possible
- Transaction status MUST be tracked and reported (pending, confirmed, failed)
- Gas estimation failures MUST be surfaced with current network conditions
- All external calls (RPC, contract reads) MUST have timeout protection

**Rationale**: DEX operations involve user funds. Clear feedback prevents user confusion and enables rapid debugging. Timeout and cancellation support ensure responsive UX even during network issues.

### IV. Performance Requirements

**Rule**: The system MUST handle DEX operations efficiently without unnecessary resource consumption.

- ABI parsing MUST be cached using `sync.Once` or equivalent
- Contract clients MUST be reused across operations (no redundant instantiation)
- RPC call batching MUST be used when querying multiple pools or positions
- Memory allocations MUST be minimized in hot paths (use pre-allocated slices for known-size operations)
- Transaction signing MUST NOT block—use async patterns where appropriate
- Performance targets:
  - Contract call decoding: <10ms per transaction
  - Transaction building: <50ms per operation
  - RPC response handling: <100ms excluding network latency
- Profiling data MUST be collected for any operation taking >200ms

**Rationale**: Liquidity repositioning agents need real-time responsiveness. Cached ABIs and reused clients prevent redundant work. Performance targets ensure sub-second operation times critical for competitive DEX interactions.

### V. Security & Error Handling

**Rule**: All blockchain interactions MUST be validated and secured against common vulnerabilities.

- Input validation MUST occur at system boundaries (user input, external APIs)
- Contract addresses MUST be validated before any transaction
- Transaction parameters (amounts, slippage, deadlines) MUST be range-checked
- Private keys MUST NEVER be logged or exposed in error messages
- Nonce management MUST prevent transaction replacement attacks
- Gas price calculations MUST include safety margins
- Failed transactions MUST be logged with full context (excluding sensitive data)
- No secrets (.env, credentials.json) MUST be committed to version control

**Rationale**: Blockchain operations are permanent and involve value transfer. Validation at boundaries prevents injection attacks. Secure key handling is non-negotiable for custody safety.

## Development Standards

### Project Structure

- `pkg/contractclient/`: Generic EVM contract interaction (ABI pack/unpack, tx send)
- `internal/util/`: Internal utilities and helpers
- `cmd/`: CLI entry points
- `blackhole.go`: High-level Blackhole DEX operations
- `blackhole_interfaces.go`: Interface definitions
- `types.go`: Parameter types and structures
- `specs/`: Feature specifications and design documents
- `tests/`: All test files mirroring source structure

### Dependency Management

- MUST use go.mod for all dependency tracking
- Primary dependencies: `github.com/ethereum/go-ethereum`, `github.com/stretchr/testify`
- New dependencies MUST be justified and reviewed for security
- Dependency updates MUST include compatibility testing

### Code Organization

- One contract client per DEX contract type (Router, Pair, VotingEscrow, Gauge)
- Clear separation: models → services → clients
- Interfaces MUST be defined before implementations
- No circular dependencies between packages

## Quality Gates

### Pre-Commit Requirements

- [ ] All tests pass (`go test ./...`)
- [ ] Code builds without warnings (`go build ./...`)
- [ ] Error handling covers all failure paths
- [ ] No hardcoded secrets or private keys
- [ ] Logging includes appropriate context (no sensitive data)

### Pre-PR Requirements

- [ ] Table-driven tests added for new business logic
- [ ] Integration tests added for new contract interactions
- [ ] Error messages are clear and actionable
- [ ] Documentation updated (README, CLAUDE.md if architectural changes)
- [ ] Performance targets met (profiling data if new hot paths)
- [ ] Contract addresses verified against Avalanche mainnet

### Architecture Review Triggers

Require architecture review if:
- Adding new external dependencies beyond go-ethereum ecosystem
- Changing transaction signing or nonce management
- Modifying core ContractClient interface
- Introducing new storage or caching mechanisms
- Altering error handling patterns

## Governance

### Amendment Process

1. Propose change with rationale and impact analysis
2. Update constitution.md with version bump:
   - MAJOR: Backward incompatible governance/principle removals or redefinitions
   - MINOR: New principle/section added or materially expanded guidance
   - PATCH: Clarifications, wording, typo fixes, non-semantic refinements
3. Update dependent templates (plan-template.md, spec-template.md, tasks-template.md)
4. Document Sync Impact Report in HTML comment at top of constitution
5. Review and approve changes
6. Commit with message: `docs: amend constitution to vX.Y.Z (description)`

### Compliance Verification

- All PRs MUST verify compliance with Core Principles
- Code reviews MUST reference specific principles when requesting changes
- Complexity violations MUST be justified in plan.md Complexity Tracking section
- Quality Gates MUST be checked before merge

### Runtime Development Guidance

For day-to-day development guidance beyond governance rules, refer to `CLAUDE.md`.

**Version**: 1.0.0 | **Ratified**: 2025-12-06 | **Last Amended**: 2025-12-06
