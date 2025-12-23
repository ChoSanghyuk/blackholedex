# Phase 0: Research & Decision Log

**Feature**: Automated Liquidity Repositioning Strategy
**Date**: 2025-12-23

## Purpose

This document consolidates research findings and technical decisions made during the planning phase for RunStrategy1. All unknowns from the Technical Context section are resolved here.

---

## Research Topics

### R1: Price Monitoring Strategy

**Question**: What is the optimal approach for continuous price monitoring without excessive RPC calls?

**Decision**: **Polling with adaptive intervals**

**Rationale**:
- Avalanche C-Chain has 2-second block time, but concentrated liquidity positions typically don't need sub-second monitoring
- WebSocket subscriptions for newHead events would be ideal but add complexity and connection management overhead
- Polling at 60-second intervals (constitutional minimum) balances responsiveness with RPC cost
- Adaptive intervals could increase frequency during high volatility (detected by consecutive price changes)

**Alternatives Considered**:
- **WebSocket subscriptions**: More real-time but requires persistent connection management, reconnection logic, and increases complexity
- **Event-based monitoring**: Listening to pool Swap events would be reactive but misses external price changes and requires event filtering
- **Block-by-block polling**: Too frequent given 2-second blocks, would hit rate limits quickly

**Implementation Approach**:
```go
// Polling loop with configurable interval
ticker := time.NewTicker(config.MonitoringInterval) // default 60s
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        currentPrice, err := b.GetAMMState(poolAddress)
        // ... check if out of range
    case <-ctx.Done():
        return // graceful shutdown
    }
}
```

---

### R2: Price Stability Detection Algorithm

**Question**: How to reliably detect price stabilization while avoiding false positives during genuine trends?

**Decision**: **Sliding window with percentage threshold**

**Rationale**:
- Absolute price thresholds fail during different market regimes (volatile vs calm)
- Percentage-based threshold (0.5% from spec) adapts to any price level
- Consecutive intervals requirement (5 from spec) filters noise while allowing quick detection
- Window resets on ANY interval exceeding threshold, preventing premature re-entry

**Alternatives Considered**:
- **Standard deviation-based**: More statistically rigorous but adds computation and requires larger sample sizes
- **Exponential moving average**: Smooth but lags during actual stabilization, delaying profitable re-entry
- **Fixed price bands**: Fails to adapt to different price levels (10 USDC/WAVAX vs 100 USDC/WAVAX)

**Implementation Approach**:
```go
type StabilityWindow struct {
    threshold          float64  // 0.5% = 0.005
    requiredIntervals  int      // 5 consecutive intervals
    lastPrice          *big.Int
    stableCount        int
}

func (sw *StabilityWindow) CheckStability(currentPrice *big.Int) bool {
    if sw.lastPrice == nil {
        sw.lastPrice = currentPrice
        sw.stableCount = 1
        return false
    }

    // Calculate percentage change
    diff := new(big.Int).Sub(currentPrice, sw.lastPrice)
    pctChange := new(big.Float).Quo(
        new(big.Float).SetInt(diff),
        new(big.Float).SetInt(sw.lastPrice),
    )
    pctChangeFloat, _ := pctChange.Float64()

    if math.Abs(pctChangeFloat) <= sw.threshold {
        sw.stableCount++
        if sw.stableCount >= sw.requiredIntervals {
            return true // Stable!
        }
    } else {
        sw.stableCount = 0 // Reset on volatility
    }

    sw.lastPrice = currentPrice
    return false
}
```

---

### R3: Token Ratio Rebalancing Strategy

**Question**: How to calculate optimal swap amounts to achieve 50:50 value ratio?

**Decision**: **Value-based proportional rebalancing with current pool price**

**Rationale**:
- Target is 50:50 **value** ratio, not quantity ratio (quantities differ based on price)
- Current pool price (from GetAMMState) provides accurate valuation
- Swap exactly the amount needed to reach 50:50 split minimizes slippage and fees
- Existing ComputeAmounts utility already handles tick-based optimal quantities

**Alternatives Considered**:
- **Fixed percentage swaps**: Doesn't guarantee 50:50 outcome, requires iteration
- **Oracle-based pricing**: Adds dependency and potential manipulation risk; pool price is canonical
- **Iterative rebalancing**: More accurate but requires multiple swaps, increasing gas costs

**Implementation Approach**:
```go
// Get current token balances
wavaxBalance := getBalance(WAVAX)
usdcBalance := getBalance(USDC)

// Get current pool price (USDC per WAVAX)
poolState, _ := b.GetAMMState(wavaxUsdcPair)
priceUSDCperWAVAX := sqrtPriceToPrice(poolState.SqrtPrice)

// Calculate current values in USDC terms
wavaxValueInUSDC := wavaxBalance * priceUSDCperWAVAX
totalValue := wavaxValueInUSDC + usdcBalance

// Target 50% of total value in each token
targetUSDC := totalValue / 2
targetWAVAXValue := totalValue / 2

if usdcBalance > targetUSDC {
    // Swap excess USDC to WAVAX
    amountToSwap := usdcBalance - targetUSDC
    b.Swap(USDC -> WAVAX, amountToSwap)
} else {
    // Swap excess WAVAX to USDC
    excessWAVAXValue := wavaxValueInUSDC - targetWAVAXValue
    amountWAVAXToSwap := excessWAVAXValue / priceUSDCperWAVAX
    b.Swap(WAVAX -> USDC, amountWAVAXToSwap)
}
```

---

### R4: Out-of-Range Detection

**Question**: How to accurately determine when current price is outside the active liquidity range?

**Decision**: **Direct tick comparison using position bounds**

**Rationale**:
- Algebra/Uniswap V3 concentrated liquidity uses tick-based ranges
- Position has tickLower and tickUpper (stored from Mint operation)
- Current price tick is available from GetAMMState
- Simple integer comparison: `currentTick < tickLower || currentTick > tickUpper`
- No floating-point arithmetic needed, eliminates rounding errors

**Alternatives Considered**:
- **SqrtPrice-based comparison**: Requires conversion from tick to sqrtPrice, introduces precision issues
- **Token ratio-based**: Indirect and requires knowing ideal ratio for current price
- **Liquidity amount checking**: Position can have zero active liquidity but still be in range technically

**Implementation Approach**:
```go
type PositionRange struct {
    tickLower int32
    tickUpper int32
}

func (pr *PositionRange) IsOutOfRange(currentTick int32) bool {
    return currentTick < pr.tickLower || currentTick > pr.tickUpper
}

// Usage in monitoring loop
state, _ := b.GetAMMState(wavaxUsdcPair)
if position.IsOutOfRange(state.Tick) {
    // Trigger rebalancing workflow
    sendToChannel(reportChan, "Position out of range, initiating rebalancing")
    triggerRebalance()
}
```

---

### R5: Financial Tracking and Reporting

**Question**: How to implement continuous reporting of errors, profits, and gas costs via string channel?

**Decision**: **Structured JSON messages sent to channel for all significant events**

**Rationale**:
- String channel allows flexible reporting format (JSON for structured data)
- Non-blocking sends prevent strategy deadlock if channel consumer is slow
- Include timestamp, event type, and relevant financial data in each message
- Cumulative tracking of gas, fees, and incentives enables real-time P&L reporting
- Channel consumers can parse JSON and store/display as needed

**Alternatives Considered**:
- **Callback functions**: Tighter coupling, harder to test independently
- **Logging to files**: Less real-time, requires file I/O overhead
- **Metrics library (Prometheus)**: Overkill for initial version, can add later

**Implementation Approach**:
```go
type StrategyReport struct {
    Timestamp    time.Time `json:"timestamp"`
    EventType    string    `json:"event_type"` // "error", "gas_cost", "profit", "rebalance_start", etc.
    Message      string    `json:"message"`
    GasCost      *big.Int  `json:"gas_cost,omitempty"`
    Profit       *big.Int  `json:"profit,omitempty"`
    CumulativeGas *big.Int `json:"cumulative_gas,omitempty"`
    Error        string    `json:"error,omitempty"`
}

func sendReport(ch chan<- string, report StrategyReport) {
    jsonBytes, err := json.Marshal(report)
    if err != nil {
        log.Printf("Failed to marshal report: %v", err)
        return
    }

    select {
    case ch <- string(jsonBytes):
        // Sent successfully
    default:
        // Channel full, drop message to prevent blocking
        log.Printf("Report channel full, dropping message")
    }
}

// Usage examples:
sendReport(reportChan, StrategyReport{
    Timestamp: time.Now(),
    EventType: "rebalance_start",
    Message:   "Position out of range detected, starting rebalance workflow",
})

sendReport(reportChan, StrategyReport{
    Timestamp:     time.Now(),
    EventType:     "gas_cost",
    Message:       "Swap transaction completed",
    GasCost:       swapGasCost,
    CumulativeGas: totalGasSpent,
})

sendReport(reportChan, StrategyReport{
    Timestamp: time.Now(),
    EventType: "profit",
    Message:   "Rewards collected",
    Profit:    rewardAmounts.Reward,
})
```

---

### R6: Error Handling and Circuit Breaker

**Question**: How to implement fail-safe error handling to prevent fund loss?

**Decision**: **Error accumulation with threshold-based halt + immediate halt on critical errors**

**Rationale**:
- Transient errors (RPC timeout, gas spike) shouldn't halt strategy permanently
- Critical errors (insufficient balance, contract revert, invalid state) require immediate halt
- Error counter with sliding time window allows recovery from transient issues
- Explicit error classification enables appropriate response strategies
- All errors reported via channel for operator visibility

**Alternatives Considered**:
- **Immediate halt on any error**: Too conservative, would stop on every RPC hiccup
- **Retry indefinitely**: Could drain gas or lock position during genuine failures
- **Manual intervention only**: Delays response to critical issues, increases risk

**Implementation Approach**:
```go
type CircuitBreaker struct {
    errorCount        int
    errorWindow       time.Duration // e.g., 5 minutes
    errorThreshold    int           // e.g., 5 errors in window
    criticalErrorOccurred bool
    lastErrors        []time.Time
}

func (cb *CircuitBreaker) RecordError(err error, critical bool) bool {
    now := time.Now()

    if critical {
        cb.criticalErrorOccurred = true
        return true // Halt immediately
    }

    // Record non-critical error
    cb.lastErrors = append(cb.lastErrors, now)

    // Remove errors outside window
    cutoff := now.Add(-cb.errorWindow)
    validErrors := []time.Time{}
    for _, t := range cb.lastErrors {
        if t.After(cutoff) {
            validErrors = append(validErrors, t)
        }
    }
    cb.lastErrors = validErrors
    cb.errorCount = len(validErrors)

    // Check if threshold exceeded
    return cb.errorCount >= cb.errorThreshold
}

// Usage in RunStrategy1
cb := &CircuitBreaker{
    errorWindow:    5 * time.Minute,
    errorThreshold: 5,
}

for {
    err := monitorAndRebalance()
    if err != nil {
        critical := isCriticalError(err)
        sendReport(reportChan, StrategyReport{
            EventType: "error",
            Message:   err.Error(),
            Error:     err.Error(),
        })

        if cb.RecordError(err, critical) {
            sendReport(reportChan, StrategyReport{
                EventType: "halt",
                Message:   "Circuit breaker triggered, halting strategy",
            })
            return // Halt strategy
        }
    }
}

func isCriticalError(err error) bool {
    criticalPatterns := []string{
        "insufficient balance",
        "NFT not owned",
        "transaction reverted",
        "invalid position state",
    }
    errStr := err.Error()
    for _, pattern := range criticalPatterns {
        if strings.Contains(errStr, pattern) {
            return true
        }
    }
    return false
}
```

---

### R7: Graceful Shutdown and Context Management

**Question**: How to allow strategy to be stopped gracefully without leaving funds in intermediate state?

**Decision**: **Context-based cancellation with state checkpointing**

**Rationale**:
- Go context pattern is idiomatic for cancellation signaling
- Checking context at safe points (before starting new operation) prevents mid-transaction cancellation
- Current position state is always recoverable from blockchain (via positions() call)
- Shutdown signal via channel combined with context provides flexibility

**Alternatives Considered**:
- **Global stop flag**: Not goroutine-safe without mutex, less idiomatic
- **Signal handling only**: Limited to OS signals, harder to integrate with application logic
- **No graceful shutdown**: Risky if stopped mid-rebalancing, could leave NFT unstaked or liquidity withdrawn

**Implementation Approach**:
```go
func (b *Blackhole) RunStrategy1(
    ctx context.Context,
    reportChan chan<- string,
    config *StrategyConfig,
) error {
    ticker := time.NewTicker(config.MonitoringInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            // Graceful shutdown requested
            sendReport(reportChan, StrategyReport{
                EventType: "shutdown",
                Message:   "Strategy shutdown requested, exiting gracefully",
            })
            return ctx.Err()

        case <-ticker.C:
            // Safe checkpoint: no operations in progress
            err := b.monitorAndRebalance(ctx, reportChan, config)
            if err != nil {
                // Handle error (see R6)
            }
        }
    }
}

// Rebalancing checks context before each major operation
func (b *Blackhole) monitorAndRebalance(ctx context.Context, ...) error {
    // Check context before starting withdrawal
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    // Perform withdrawal (atomic via multicall)
    withdrawResult, err := b.Withdraw(nftTokenID)
    // ...

    // Check context before starting rebalance
    select {
    case <-ctx.Done():
        // Funds are safely in wallet, position is closed
        return ctx.Err()
    default:
    }

    // Rebalance tokens
    // ...
}
```

---

## Summary of Decisions

| Topic | Decision | Impact |
|-------|----------|--------|
| Price Monitoring | Polling at 60s intervals | Balances RPC cost with responsiveness |
| Stability Detection | Sliding window, 0.5% threshold, 5 consecutive intervals | Filters volatility, prevents premature re-entry |
| Ratio Rebalancing | Value-based with pool price | Achieves exact 50:50 split, minimizes swaps |
| Out-of-Range Detection | Direct tick comparison | Simple, accurate, no floating-point errors |
| Financial Reporting | JSON messages via string channel | Structured, non-blocking, flexible consumption |
| Error Handling | Circuit breaker + critical error halt | Balances resilience with safety |
| Graceful Shutdown | Context-based cancellation | Prevents mid-transaction stops, safe fund state |

---

## Open Questions (if any)

None - all technical unknowns resolved. Ready to proceed to Phase 1 (Data Model and Contracts).
