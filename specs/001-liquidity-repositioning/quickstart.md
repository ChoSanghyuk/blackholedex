# Quickstart Guide: RunStrategy1 - Automated Liquidity Repositioning

**Feature**: Automated Liquidity Repositioning Strategy
**Date**: 2025-12-23
**Target Audience**: Developers implementing or using RunStrategy1

---

## Overview

RunStrategy1 is an autonomous liquidity management strategy for the WAVAX/USDC concentrated liquidity pool on Blackhole DEX. It automatically:

1. **Rebalances** tokens to maintain 50:50 value ratio
2. **Creates** concentrated liquidity positions with configurable tick width
3. **Monitors** pool price continuously
4. **Withdraws** positions when price moves out of active range
5. **Waits** for price stabilization before re-entering
6. **Reports** all operations, gas costs, and profits via a string channel

---

## Prerequisites

- Go 1.24.10+
- Access to Avalanche C-Chain RPC endpoint
- Wallet with WAVAX and USDC tokens
- Private key for wallet (keep secure!)
- Existing Blackhole instance created via `NewBlackhole()`

---

## Quick Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/big"
    "time"

    "blackholego"
    "blackholego/contracts" // Import strategy contracts
)

func main() {
    // 1. Create Blackhole instance
    rpcURL := "https://api.avax.network/ext/bc/C/rpc"
    privateKey := "your_private_key_hex" // Keep secure!

    configs := []blackholedex.ContractClientConfig{
        {address: "0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7", abipath: "abis/WAVAX.json"},
        {address: "0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E", abipath: "abis/USDC.json"},
        {address: "0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0", abipath: "abis/AlgebraPool.json"},
        {address: "0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146", abipath: "abis/NonfungiblePositionManager.json"},
        {address: "0x3ADE52f9779c07471F4B6d5997444C3c2124C1c0", abipath: "abis/GaugeV2.json"},
        {address: "0xa47Ad2C95FaE476a73b85A355A5855aDb4b3A449", abipath: "abis/FarmingCenter.json"},
        {address: "0x04E1dee021Cd12bBa022A72806441B43d8212Fec", abipath: "abis/RouterV2.json"},
    }

    txListener := NewTxListener(rpcURL) // Your TxListener implementation

    blackhole, err := blackholedex.NewBlackhole(privateKey, txListener, rpcURL, configs)
    if err != nil {
        log.Fatalf("Failed to create Blackhole: %v", err)
    }

    // 2. Configure strategy
    config := contracts.DefaultStrategyConfig()
    config.MaxWAVAX = big.NewInt(1000000000000000000) // 1 WAVAX (18 decimals)
    config.MaxUSDC = big.NewInt(10000000)             // 10 USDC (6 decimals)
    config.MonitoringInterval = 60 * time.Second      // Check price every minute
    config.RangeWidth = 10                            // ±5 ticks from center

    // 3. Create report channel
    reportChan := make(chan string, 100) // Buffered to prevent blocking

    // 4. Start report consumer (in separate goroutine)
    go func() {
        for report := range reportChan {
            fmt.Println(report) // Or: parse JSON, log to file, send to monitoring system
        }
    }()

    // 5. Create context for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 6. Run strategy
    log.Println("Starting RunStrategy1...")
    err = blackhole.RunStrategy1(ctx, reportChan, config)
    if err != nil {
        log.Fatalf("Strategy error: %v", err)
    }

    log.Println("Strategy completed successfully")
}
```

---

## Configuration Guide

### Default Configuration

```go
config := contracts.DefaultStrategyConfig()
// Returns:
// MonitoringInterval: 60s (constitutional minimum)
// StabilityThreshold: 0.005 (0.5%)
// StabilityIntervals: 5 (consecutive stable checks)
// RangeWidth: 10 ticks (±5 from center)
// SlippagePct: 1 (1% slippage tolerance)
// CircuitBreakerWindow: 5 minutes
// CircuitBreakerThreshold: 5 errors
```

### Customizing Configuration

```go
config := contracts.DefaultStrategyConfig()

// Set capital allocation (REQUIRED)
config.MaxWAVAX = big.NewInt(5000000000000000000) // 5 WAVAX
config.MaxUSDC = big.NewInt(100000000)           // 100 USDC

// Adjust monitoring frequency (optional)
config.MonitoringInterval = 30 * time.Second // More responsive, more RPC calls

// Adjust position range (optional)
config.RangeWidth = 20 // Wider range, less frequent rebalancing

// Adjust stability detection (optional)
config.StabilityThreshold = 0.01     // 1% threshold (less strict)
config.StabilityIntervals = 3        // Faster re-entry

// Adjust error tolerance (optional)
config.CircuitBreakerThreshold = 10  // Allow more errors before halting
```

---

## Report Channel Usage

### Report Structure

All reports are JSON strings conforming to `StrategyReport`:

```json
{
  "timestamp": "2025-12-23T10:30:00Z",
  "event_type": "position_created",
  "message": "New liquidity position created and staked",
  "phase": "ActiveMonitoring",
  "gas_cost": "150000000000000",
  "cumulative_gas": "450000000000000",
  "nft_token_id": "12345"
}
```

### Parsing Reports

```go
import (
    "encoding/json"
    "blackholego/contracts"
)

func processReport(reportJSON string) {
    var report contracts.StrategyReport
    err := json.Unmarshal([]byte(reportJSON), &report)
    if err != nil {
        log.Printf("Failed to parse report: %v", err)
        return
    }

    switch report.EventType {
    case "strategy_start":
        log.Println("Strategy started")

    case "out_of_range":
        log.Printf("Position %s out of range, rebalancing...", report.NFTTokenID)

    case "gas_cost":
        gasCostETH := new(big.Float).Quo(
            new(big.Float).SetInt(report.GasCost),
            big.NewFloat(1e18),
        )
        log.Printf("Gas cost: %s AVAX", gasCostETH.String())

    case "profit":
        log.Printf("Rewards collected: %s BLACK", report.Profit)

    case "error":
        log.Printf("ERROR: %s - %s", report.Message, report.Error)

    case "halt":
        log.Printf("STRATEGY HALTED: %s", report.Message)
        // Trigger alert, notify operator, etc.

    case "shutdown":
        log.Println("Strategy shutdown gracefully")
    }
}

// Usage
go func() {
    for report := range reportChan {
        processReport(report)
    }
}()
```

---

## Lifecycle and Shutdown

### Graceful Shutdown

```go
ctx, cancel := context.WithCancel(context.Background())

// Start strategy in goroutine
go func() {
    err := blackhole.RunStrategy1(ctx, reportChan, config)
    if err != nil {
        log.Printf("Strategy error: %v", err)
    }
}()

// Shutdown after 24 hours (or on signal)
time.Sleep(24 * time.Hour)
cancel() // Triggers graceful shutdown

// Wait for strategy to exit
time.Sleep(5 * time.Second)
close(reportChan) // Close channel after strategy exits
```

### Handling OS Signals

```go
import (
    "os"
    "os/signal"
    "syscall"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Start strategy
    go func() {
        err := blackhole.RunStrategy1(ctx, reportChan, config)
        if err != nil {
            log.Printf("Strategy error: %v", err)
        }
    }()

    // Wait for signal
    <-sigChan
    log.Println("Shutdown signal received, stopping strategy...")
    cancel()

    // Allow time for graceful shutdown
    time.Sleep(10 * time.Second)
}
```

---

## Monitoring and Alerts

### Real-Time Dashboard

```go
type StrategyMetrics struct {
    TotalGasSpent   *big.Int
    TotalRewards    *big.Int
    NetPnL          *big.Int
    Rebalances      int
    ErrorCount      int
    CurrentPhase    string
    Uptime          time.Duration
}

var metrics StrategyMetrics
var metricsLock sync.Mutex

func updateMetrics(report contracts.StrategyReport) {
    metricsLock.Lock()
    defer metricsLock.Unlock()

    if report.CumulativeGas != nil {
        metrics.TotalGasSpent = report.CumulativeGas
    }
    if report.Profit != nil {
        metrics.TotalRewards = new(big.Int).Add(metrics.TotalRewards, report.Profit)
    }
    if report.NetPnL != nil {
        metrics.NetPnL = report.NetPnL
    }
    if report.EventType == "position_created" {
        metrics.Rebalances++
    }
    if report.EventType == "error" {
        metrics.ErrorCount++
    }
    if report.Phase != nil {
        metrics.CurrentPhase = report.Phase.String()
    }
}

// Expose metrics via HTTP endpoint
http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
    metricsLock.Lock()
    defer metricsLock.Unlock()
    json.NewEncoder(w).Encode(metrics)
})
```

### Alert Conditions

```go
func checkAlerts(report contracts.StrategyReport) {
    // Alert on high gas costs
    if report.GasCost != nil {
        threshold := big.NewInt(100000000000000) // 0.0001 AVAX
        if report.GasCost.Cmp(threshold) > 0 {
            sendAlert("High gas cost: " + report.GasCost.String())
        }
    }

    // Alert on errors
    if report.EventType == "error" {
        sendAlert("Strategy error: " + report.Error)
    }

    // Alert on circuit breaker
    if report.EventType == "halt" {
        sendAlert("CRITICAL: Strategy halted - " + report.Message)
    }

    // Alert on negative P&L
    if report.NetPnL != nil && report.NetPnL.Sign() < 0 {
        sendAlert("Negative P&L: " + report.NetPnL.String())
    }
}

func sendAlert(message string) {
    // Send to Slack, email, SMS, etc.
    log.Printf("ALERT: %s", message)
}
```

---

## Testing

### Integration Test Example

```go
func TestRunStrategy1_FullCycle(t *testing.T) {
    // Setup mock RPC client
    mockRPC := NewMockRPCClient()

    // Mock pool state (in-range initially)
    mockRPC.SetAMMState(&blackholedex.AMMState{
        Tick: 1000, // Center of position
        SqrtPrice: big.NewInt(1000000000000000000),
    })

    // Create Blackhole with mock
    blackhole := createTestBlackhole(mockRPC)

    // Configure strategy
    config := contracts.DefaultStrategyConfig()
    config.MaxWAVAX = big.NewInt(1000000000000000000)
    config.MaxUSDC = big.NewInt(10000000)
    config.MonitoringInterval = 100 * time.Millisecond // Fast for testing

    // Run strategy
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    reportChan := make(chan string, 100)

    go func() {
        err := blackhole.RunStrategy1(ctx, reportChan, config)
        assert.NoError(t, err)
    }()

    // Verify initial position creation
    report := waitForReport(reportChan, "position_created", 2*time.Second)
    assert.NotNil(t, report)

    // Simulate price moving out of range
    mockRPC.SetAMMState(&blackholedex.AMMState{
        Tick: 2000, // Out of range
        SqrtPrice: big.NewInt(2000000000000000000),
    })

    // Verify rebalancing triggered
    report = waitForReport(reportChan, "rebalance_start", 2*time.Second)
    assert.NotNil(t, report)

    // Verify new position created
    report = waitForReport(reportChan, "position_created", 5*time.Second)
    assert.NotNil(t, report)
}
```

---

## Troubleshooting

### Common Issues

**1. Strategy halts immediately**
- Check wallet balances (MaxWAVAX/MaxUSDC must be <= available)
- Verify RPC endpoint is accessible
- Check private key is valid

**2. Excessive rebalancing**
- Increase `StabilityThreshold` (e.g., 0.01 for 1%)
- Widen `RangeWidth` (e.g., 20 instead of 10)
- Increase `MonitoringInterval` to reduce frequency

**3. Channel blocking warnings**
- Increase reportChan buffer size
- Ensure report consumer is running
- Use non-blocking send (default behavior)

**4. High gas costs**
- Check Avalanche C-Chain gas prices
- Review rebalancing frequency
- Consider wider tick ranges

**5. Position not rebalancing when expected**
- Verify StabilityIntervals requirement is met
- Check circuit breaker hasn't triggered
- Review error reports for transaction failures

---

## Advanced Usage

### Custom Error Handling

```go
type CustomCircuitBreaker struct {
    *contracts.CircuitBreaker
    alertFunc func(string)
}

func (ccb *CustomCircuitBreaker) RecordError(err error, critical bool) bool {
    shouldHalt := ccb.CircuitBreaker.RecordError(err, critical)

    if critical {
        ccb.alertFunc("CRITICAL ERROR: " + err.Error())
    }

    if shouldHalt {
        ccb.alertFunc("Circuit breaker triggered, halting strategy")
    }

    return shouldHalt
}
```

### Dynamic Configuration Adjustment

```go
// Adjust configuration based on market conditions
func adjustConfigForVolatility(config *contracts.StrategyConfig, volatility float64) {
    if volatility > 0.1 { // High volatility
        config.StabilityThreshold = 0.01        // More lenient
        config.StabilityIntervals = 10          // Wait longer
        config.MonitoringInterval = 30 * time.Second // Check more often
    } else { // Low volatility
        config.StabilityThreshold = 0.005       // Stricter
        config.StabilityIntervals = 5           // Default
        config.MonitoringInterval = 60 * time.Second
    }
}
```

---

## Performance Expectations

Based on spec Success Criteria:

| Metric | Target | Notes |
|--------|--------|-------|
| Position uptime | 80%+ | During normal market conditions |
| Rebalancing time | < 10 min | Stable network, includes stability wait |
| Ratio accuracy | < 1% deviation | From 50:50 target |
| Gas cost per rebalance | < 2% of position value | Avalanche has low fees |
| Out-of-range detection | < 2 minutes | With 60s monitoring interval |
| Stability prevention | 90%+ | Avoids rebalancing during volatility |
| Continuous operation | 24+ hours | Without manual intervention |

---

## Next Steps

1. **Read the specification**: [spec.md](spec.md)
2. **Review data model**: [data-model.md](data-model.md)
3. **Understand research decisions**: [research.md](research.md)
4. **Check API contracts**: [contracts/strategy_api.go](contracts/strategy_api.go)
5. **Implement tasks**: See [tasks.md](tasks.md) (generated by `/speckit.tasks`)

---

## Support

For issues or questions:
- Review logs and error reports via reportChan
- Check circuit breaker status
- Verify constitutional compliance
- Consult [CLAUDE.md](../../CLAUDE.md) project guide
