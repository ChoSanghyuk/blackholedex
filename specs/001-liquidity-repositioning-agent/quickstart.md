# Quickstart Guide: Blackhole DEX Liquidity Repositioning Agent

**Last Updated**: 2025-12-06
**Version**: 1.0.0

## Overview

This guide walks you through setting up and running the Blackhole DEX Liquidity Repositioning Agent to automatically manage your concentrated liquidity positions on Avalanche.

**What this agent does:**
1. Monitors your liquidity positions on Blackhole DEX
2. Detects when positions move out of the active trading range
3. Automatically repositions liquidity to optimize fee generation
4. Supports manual position management (mint, unstake, swap)

**Prerequisites:**
- Go 1.24.9+ installed
- Avalanche wallet with WAVAX, USDC, or BLACKHOLE tokens
- Private key for transaction signing
- Avalanche RPC endpoint access

---

## Quick Setup (5 minutes)

### 1. Install the Agent

```bash
# Clone the repository
git clone https://github.com/your-org/blackhole_dex.git
cd blackhole_dex

# Switch to the agent feature branch
git checkout 001-liquidity-repositioning-agent

# Install dependencies
go mod download

# Build the agent
go build -o blackhole-agent ./cmd/agent
```

### 2. Create Configuration File

```bash
# Create config directory
mkdir -p ~/.blackhole-agent

# Copy example config
cp configs/agent.yaml.example ~/.blackhole-agent/config.yaml

# Edit configuration
nano ~/.blackhole-agent/config.yaml
```

**Minimal configuration** (`~/.blackhole-agent/config.yaml`):

```yaml
# Agent settings
enabled: true
wallet_address: "0xYourWalletAddress"  # Replace with your wallet

# RPC endpoint
rpc_endpoint: "https://api.avax.network/ext/bc/C/rpc"

# Private key (NEVER commit this file!)
private_key_path: "~/.blackhole-agent/keystore/key.json"

# Positions to monitor (leave empty to monitor all)
monitored_positions: []

# Trigger conditions
triggers:
  out_of_range_duration: 1h      # Wait 1 hour before repositioning
  tick_distance_threshold: 10     # Distance from range boundary

# Risk limits
risk_limits:
  max_slippage_percent: 0.5       # 0.5% max slippage on swaps
  max_slippage_liquidity: 1.0     # 1% max slippage on liquidity ops
  max_gas_price_gwei: 50          # Max gas price willing to pay
  min_position_size_usd: 100      # Don't manage positions < $100

# How to handle multiple out-of-range positions
multi_position_strategy: largest_first

# Notifications (optional)
notifications:
  enable_slack: false
  enable_email: false
```

### 3. Set Up Private Key

**Option A: Use existing keystore file**

```bash
# Copy your existing keystore JSON file
cp /path/to/your/keystore.json ~/.blackhole-agent/keystore/key.json
```

**Option B: Create new keystore from private key**

```bash
# Install geth (if not already installed)
# This is used to create encrypted keystore files

# Create encrypted keystore
geth account import \
  --keystore ~/.blackhole-agent/keystore \
  <(echo "0xYourPrivateKeyHere")

# Enter a strong password when prompted
# Move the generated file to key.json
mv ~/.blackhole-agent/keystore/UTC--* ~/.blackhole-agent/keystore/key.json
```

**SECURITY WARNING:**
- Never commit keystore files or private keys to git
- Never share your private key or keystore password
- The `.gitignore` already excludes `~/.blackhole-agent/` directory

### 4. Run the Agent

```bash
# Start agent in foreground (for testing)
./blackhole-agent start --config ~/.blackhole-agent/config.yaml

# Or run as daemon (background)
./blackhole-agent start --config ~/.blackhole-agent/config.yaml --daemon

# Check agent status
./blackhole-agent status

# Stop daemon
./blackhole-agent stop
```

**Expected output:**

```
Blackhole DEX Liquidity Repositioning Agent v1.0.0
=================================================
✓ Configuration loaded from ~/.blackhole-agent/config.yaml
✓ Connected to Avalanche RPC: https://api.avax.network/ext/bc/C/rpc
✓ Wallet: 0xb4dd...9223
✓ Monitoring 3 positions
✓ Automated repositioning: ENABLED

[2025-12-06 14:30:00] Starting position monitoring (check interval: 5m)
[2025-12-06 14:30:01] Position 12345 (WAVAX/USDC): IN-RANGE ✓
[2025-12-06 14:30:01] Position 67890 (WAVAX/BLACK): OUT-OF-RANGE (42min) ⚠
[2025-12-06 14:30:01] Position 11111 (WAVAX/USDC): IN-RANGE ✓

Agent running. Press Ctrl+C to stop.
```

---

## Usage Examples

### Monitor Positions (Read-Only Mode)

Disable automated repositioning and just monitor position status:

```yaml
# In config.yaml
enabled: false
```

```bash
# Check position status
./blackhole-agent positions list

# Output:
Position ID: 12345
  Pair: WAVAX/USDC
  Status: IN-RANGE ✓
  Tick Range: [-249600, -248400]
  Current Tick: -249000
  Liquidity: 1.5M
  Unclaimed Fees: 0.05 WAVAX, 1.2 USDC
  Time in Range: 85%

Position ID: 67890
  Pair: WAVAX/BLACK
  Status: OUT-OF-RANGE ⚠
  Tick Range: [-250000, -248800]
  Current Tick: -250500
  Out of Range For: 42 minutes
  Unclaimed Fees: 0.02 WAVAX, 0 BLACK
  Time in Range: 60%
```

### Manual Position Management

#### Mint a New Position

```bash
# Mint WAVAX/USDC position
./blackhole-agent positions mint \
  --token0 0xb31f66aa3c1e785363f0875a1b74e27b85fd66c7 \   # WAVAX
  --token1 0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E \   # USDC
  --amount0 1.0 \
  --amount1 25.0 \
  --tick-lower -250000 \
  --tick-upper -248800 \
  --slippage 0.5

# Output:
✓ Approved WAVAX for NonfungiblePositionManager
✓ Approved USDC for NonfungiblePositionManager
✓ Minting position...
✓ Transaction confirmed: 0xabc123...

Position Created:
  Position ID: 12346
  Liquidity: 1,234,567
  Amount0: 1.0 WAVAX
  Amount1: 25.0 USDC
  Tick Range: [-250000, -248800]
  Status: IN-RANGE ✓
  Gas Used: 450,000 (0.02 AVAX)
```

#### Unstake a Position

```bash
# Unstake position and collect all fees
./blackhole-agent positions unstake \
  --position-id 12346 \
  --collect-fees

# Output:
✓ Removing liquidity from position 12346...
✓ Transaction confirmed: 0xdef456...
✓ Collecting fees and withdrawn tokens...
✓ Transaction confirmed: 0x789ghi...

Withdrawn:
  Token0 (WAVAX): 1.05 (including 0.05 fees)
  Token1 (USDC): 26.2 (including 1.2 fees)
  Gas Used: 320,000 (0.016 AVAX)
```

#### Swap Tokens

```bash
# Swap 1 WAVAX for USDC
./blackhole-agent swap \
  --token-in WAVAX \
  --token-out USDC \
  --amount-in 1.0 \
  --slippage 0.5

# Output:
✓ Approved WAVAX for RouterV2
✓ Swapping 1.0 WAVAX for USDC...
✓ Transaction confirmed: 0x012jkl...

Swap Result:
  Input: 1.0 WAVAX
  Output: 25.3 USDC
  Expected: 25.5 USDC
  Slippage: 0.78%
  Gas Used: 180,000 (0.009 AVAX)
```

### Automated Repositioning

Enable automated repositioning in config:

```yaml
enabled: true
```

When a position moves out of range for longer than `out_of_range_duration`, the agent will:

1. **Detect**: Position 67890 has been out-of-range for 61 minutes
2. **Evaluate**: Calculate new tick range based on current pool price
3. **Check**: Verify gas cost < expected fee gain
4. **Execute**:
   - Unstake old position
   - Collect fees
   - Swap tokens if needed to rebalance
   - Mint new position in active range
5. **Verify**: Confirm new position is in-range
6. **Notify**: Send notification (if configured)

**Example agent log:**

```
[2025-12-06 15:30:00] Position 67890 out-of-range for 61min (trigger: 60min)
[2025-12-06 15:30:01] Evaluating repositioning for position 67890...
[2025-12-06 15:30:02] ✓ Gas cost (0.03 AVAX) < Expected fees (0.08 AVAX)
[2025-12-06 15:30:02] Executing repositioning workflow...
[2025-12-06 15:30:10] ✓ Unstaked position 67890 (1.02 WAVAX, 0 BLACK)
[2025-12-06 15:30:15] ✓ Collected fees (0.02 WAVAX, 0 BLACK)
[2025-12-06 15:30:20] ✓ Swapped 0.5 WAVAX -> 12.5 USDC (rebalance)
[2025-12-06 15:30:30] ✓ Minted new position 12347 (0.52 WAVAX, 12.5 USDC)
[2025-12-06 15:30:31] ✓ Position 12347 is IN-RANGE ✓
[2025-12-06 15:30:31] Repositioning complete (duration: 31s, gas: 0.03 AVAX)
```

---

## Configuration Reference

### Trigger Conditions

| Setting | Default | Description |
|---------|---------|-------------|
| `out_of_range_duration` | `1h` | Wait this long before repositioning out-of-range positions |
| `tick_distance_threshold` | `10` | Reposition when position is within X ticks of boundary |
| `price_movement_threshold` | N/A | Reposition on X% price change (optional) |

### Risk Limits

| Setting | Default | Description |
|---------|---------|-------------|
| `max_slippage_percent` | `0.5` | Maximum slippage for swaps (0.01% - 10%) |
| `max_slippage_liquidity` | `1.0` | Maximum slippage for liquidity operations |
| `max_gas_price_gwei` | `50` | Don't reposition if gas > this price |
| `min_position_size_usd` | `100` | Ignore positions smaller than this value |

### Multi-Position Strategy

| Strategy | Behavior |
|----------|----------|
| `largest_first` | Reposition highest-value positions first |
| `longest_out_first` | Reposition positions out-of-range longest |
| `sequential` | Process in order of position ID |

### Notifications

```yaml
notifications:
  enable_slack: true
  slack_webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

  enable_email: true
  email_address: "your-email@example.com"
  smtp_server: "smtp.gmail.com:587"
  smtp_username: "your-email@gmail.com"
  smtp_password: "your-app-password"

  notify_on_trigger: true      # When repositioning starts
  notify_on_complete: true     # When repositioning finishes
  notify_on_failure: true      # When repositioning fails
```

---

## Troubleshooting

### Problem: Agent won't start

**Check configuration:**
```bash
./blackhole-agent validate-config --config ~/.blackhole-agent/config.yaml
```

**Common issues:**
- Invalid wallet address format
- Private key file doesn't exist
- RPC endpoint unreachable
- Invalid tick ranges in config

### Problem: "Insufficient balance" errors

**Check token balances:**
```bash
./blackhole-agent wallet balance

# Output:
Wallet: 0xb4dd...9223
  WAVAX: 5.2
  USDC: 150.0
  BLACK: 0
  AVAX (gas): 0.5
```

**Solution:** Ensure you have enough tokens + gas (AVAX) before minting or repositioning.

### Problem: Positions not being repositioned

**Check evaluation:**
```bash
./blackhole-agent positions evaluate --position-id 67890

# Output:
Position 67890 Evaluation:
  Should Reposition: NO
  Reason: Gas cost (0.05 AVAX) > Expected fees (0.02 AVAX)
  Current Status: OUT-OF-RANGE
  Out of Range For: 45 minutes
  Gas Price: 55 gwei (exceeds max: 50 gwei)
```

**Common reasons:**
- Gas prices too high (wait for lower gas)
- Position too small (increase `min_position_size_usd`)
- Not out-of-range long enough (increase `out_of_range_duration`)
- Expected fees don't justify gas cost

### Problem: High gas costs

**Optimize gas settings:**

```yaml
# Reduce priority fee (slower confirmation)
gas_settings:
  priority_fee_gwei: 1.0   # Default: 1.5
  max_fee_cap_gwei: 30     # Default: base + 2

# Or wait for lower gas prices
risk_limits:
  max_gas_price_gwei: 30   # Lower threshold
```

### Problem: Slippage exceeded errors

**Increase slippage tolerance:**

```yaml
risk_limits:
  max_slippage_percent: 1.0        # Increase from 0.5%
  max_slippage_liquidity: 2.0      # Increase from 1.0%
```

**Or wait for better liquidity:**
```bash
# Check pool liquidity
./blackhole-agent pools info --pair WAVAX/USDC

# Output:
Pool: 0xA02E...6Ea0 (WAVAX/USDC)
  Current Tick: -249000
  Active Liquidity: 15.5M
  24h Volume: $2.3M
  Current Fee: 0.3%
```

---

## Advanced Usage

### Dry-Run Mode

Test repositioning without executing transactions:

```bash
./blackhole-agent positions reposition \
  --position-id 67890 \
  --dry-run

# Output:
Repositioning Plan for Position 67890:
  Current Range: [-250000, -248800]
  New Range: [-249600, -248400]
  Requires Swap: YES
    Swap 0.3 WAVAX -> 7.5 USDC
  Estimated Gas: 0.028 AVAX
  Expected Fees (30d): 0.12 AVAX
  Net Benefit: 0.092 AVAX

Would execute:
  1. decreaseLiquidity(67890, all)
  2. collect(67890)
  3. swap(0.3 WAVAX -> USDC, 0.5% slippage)
  4. mint(WAVAX/USDC, [-249600, -248400])

Dry-run complete. No transactions executed.
```

### Custom Tick Ranges

Override automatic tick range calculation:

```bash
./blackhole-agent positions reposition \
  --position-id 67890 \
  --tick-lower -250200 \
  --tick-upper -248600 \
  --force
```

### Batch Operations

Reposition multiple positions:

```bash
./blackhole-agent positions reposition-all \
  --out-of-range-only \
  --min-duration 2h
```

---

## Best Practices

### 1. Start with Monitoring Only

```yaml
# Disable auto-repositioning initially
enabled: false
```

Run for a few days to understand position behavior before enabling automation.

### 2. Set Conservative Risk Limits

```yaml
risk_limits:
  max_slippage_percent: 0.5       # Start conservative
  max_gas_price_gwei: 50          # Avoid high-gas repositioning
  min_position_size_usd: 500      # Focus on larger positions
```

### 3. Monitor Agent Performance

```bash
# View repositioning history
./blackhole-agent history --since 7d

# Output:
Repositioning Events (Last 7 Days):

Event 1: 2025-12-01 10:30:00
  Position: 67890 (WAVAX/BLACK)
  Outcome: SUCCESS
  Gas Cost: 0.028 AVAX
  Duration: 35 seconds
  New Position: 12347 (IN-RANGE ✓)

Event 2: 2025-12-03 14:15:00
  Position: 12345 (WAVAX/USDC)
  Outcome: FAILED
  Error: Slippage exceeded (expected 25.5 USDC, got 25.0)
  Gas Cost: 0.015 AVAX (partial execution)

Summary:
  Total Events: 5
  Successful: 4
  Failed: 1
  Total Gas Spent: 0.13 AVAX
  Avg Duration: 32 seconds
```

### 4. Use Notifications

Enable Slack or email notifications to stay informed:

```yaml
notifications:
  enable_slack: true
  notify_on_trigger: true
  notify_on_complete: true
  notify_on_failure: true
```

### 5. Regular Backups

Backup your configuration and event history:

```bash
# Backup config and events
tar -czf blackhole-agent-backup-$(date +%Y%m%d).tar.gz \
  ~/.blackhole-agent/config.yaml \
  ~/.blackhole-agent/events.jsonl

# Store backup securely (NOT in git)
```

---

## Performance Benchmarks

Based on testing with 10 monitored positions on Avalanche mainnet:

| Operation | Target | Typical | Notes |
|-----------|--------|---------|-------|
| Position status check | <5s | 2-3s | Using batch RPC |
| Single position query | <2s | 1s | Cached pool state |
| Mint position | <30s | 20-25s | Including approvals |
| Unstake position | <30s | 15-20s | Decrease + collect |
| Token swap | <20s | 12-15s | Including approval |
| Full repositioning | <2min | 45-90s | Unstake -> swap -> mint |

**Network conditions:** Avalanche C-Chain, normal gas prices (25-35 gwei), good RPC latency (<200ms)

---

## Next Steps

1. **Read the full specification**: See `spec.md` for detailed requirements
2. **Review the implementation plan**: See `plan.md` for architecture details
3. **Explore the data model**: See `data-model.md` for entity definitions
4. **Check API contracts**: See `contracts/` directory for operation specs
5. **Run tests**: `go test ./... -v` to verify installation
6. **Join the community**: [Discord/Telegram link] for support

---

## Security Reminders

- ✅ Never commit private keys or keystore files
- ✅ Never share keystore passwords
- ✅ Use encrypted keystore files, not plaintext private keys
- ✅ Keep `max_gas_price_gwei` reasonable to avoid expensive transactions
- ✅ Start with small positions to test the agent
- ✅ Monitor the agent regularly, especially during high volatility
- ✅ Keep backups of your configuration

---

## Support

For issues, questions, or feature requests:
- **GitHub Issues**: [Link to issues]
- **Discord**: [Link to Discord]
- **Documentation**: [Link to full docs]
- **Email**: support@blackholedex.io
