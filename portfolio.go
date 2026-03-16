package blackholedex

import (
	"blackholego/pkg/types"
	"blackholego/pkg/util"
	"context"
	"fmt"
	"log"
	"math/big"
	"time"
)

// RecordCurrentAssetSnapshot records a snapshot of the current asset state
// Used by RunStrategy1 to track portfolio value over time during strategy execution
func (b *Blackhole) RecordCurrentAssetSnapshot(state types.StrategyPhase) {
	if b.recorder != nil {
		snapshot, err := b.GetCurrentAssetSnapshot(state)
		if err != nil {
			log.Printf("Warning: failed to get initial asset snapshot: %v", err)
		} else {
			if err := b.recorder.RecordReport(*snapshot); err != nil {
				log.Printf("Warning: failed to record initial snapshot: %v", err)
			} else {
				log.Printf("Initial asset snapshot recorded at strategy start")
			}
		}
	}
}

// GetCurrentAssetSnapshot fetches a complete snapshot of user's assets
// including wallet balances (WAVAX, USDC, BLACK, AVAX) and position values
// state: Current strategy phase (can be 0/Initializing if not in strategy mode)
// Returns CurrentAssetSnapshot with all balances and estimated total value in USDC
func (b *Blackhole) GetCurrentAssetSnapshot(state types.StrategyPhase) (*types.CurrentAssetSnapshot, error) {
	// Get WAVAX balance from wallet
	wavaxClient, err := b.registry.Client(wavax)
	if err != nil {
		return nil, fmt.Errorf("failed to get WAVAX client: %w", err)
	}
	wavaxBalanceResult, err := wavaxClient.Call(&b.myAddr, "balanceOf", b.myAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get WAVAX balance: %w", err)
	}
	wavaxBalance := wavaxBalanceResult[0].(*big.Int)

	// Get USDC balance from wallet
	usdcClient, err := b.registry.Client(usdc)
	if err != nil {
		return nil, fmt.Errorf("failed to get USDC client: %w", err)
	}
	usdcBalanceResult, err := usdcClient.Call(&b.myAddr, "balanceOf", b.myAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get USDC balance: %w", err)
	}
	usdcBalance := usdcBalanceResult[0].(*big.Int)

	// Get BLACK balance from wallet
	blackClient, err := b.registry.Client(black)
	if err != nil {
		return nil, fmt.Errorf("failed to get BLACK client: %w", err)
	}
	blackBalanceResult, err := blackClient.Call(&b.myAddr, "balanceOf", b.myAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get BLACK balance: %w", err)
	}
	blackBalance := blackBalanceResult[0].(*big.Int)

	// Get native AVAX balance from wallet
	avaxBalance, err := b.client.BalanceAt(context.Background(), b.myAddr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get native AVAX balance: %w", err)
	}

	// Get all user positions to include liquidity values
	positions, err := b.GetUserPositions()
	if err != nil {
		return nil, fmt.Errorf("failed to get user positions: %w", err)
	}

	// Add position values to balances
	for _, tokenID := range positions {
		position, err := b.GetPositionDetails(tokenID)
		if err != nil {
			log.Printf("Warning: failed to get position details for token %s: %v", tokenID.String(), err)
			continue
		}

		// Only count positions for WAVAX/USDC pair
		wavaxAddr, _ := b.registry.GetAddress(wavax)
		usdcAddr, _ := b.registry.GetAddress(usdc)
		if (position.Token0 == wavaxAddr || position.Token1 == wavaxAddr) &&
			(position.Token0 == usdcAddr || position.Token1 == usdcAddr) {

			// Get current pool state for price calculation
			poolState, err := b.GetAMMState()
			if err != nil {
				log.Printf("Warning: failed to get pool state: %v", err)
				continue
			}

			// Calculate token amounts in the position using liquidity and ticks
			amount0, amount1, err := util.CalculateTokenAmountsFromLiquidity(
				position.Liquidity,
				poolState.SqrtPrice,
				position.TickLower,
				position.TickUpper,
			)
			if err != nil {
				log.Printf("Warning: failed to calculate token amounts for position %s: %v", tokenID.String(), err)
				continue
			}

			// Add position token amounts to total balances
			// Token0 is WAVAX, Token1 is USDC
			wavaxBalance = new(big.Int).Add(wavaxBalance, amount0)
			usdcBalance = new(big.Int).Add(usdcBalance, amount1)
		}
	}

	// Calculate total value in USDC (6 decimals)
	// Get current WAVAX/USDC pool price
	poolState, err := b.GetAMMState()
	if err != nil {
		return nil, fmt.Errorf("failed to get pool state for price: %w", err)
	}

	// Convert sqrtPrice to actual price (USDC per WAVAX)
	price := util.SqrtPriceToPrice(poolState.SqrtPrice)

	// Calculate total value = USDC + (WAVAX * price) + (AVAX * price)
	// Convert WAVAX to USDC value
	wavaxValueFloat := new(big.Float).Mul(new(big.Float).SetInt(wavaxBalance), price)
	wavaxValueInUSDC, _ := wavaxValueFloat.Int(nil)

	// Convert native AVAX to USDC value (AVAX ≈ WAVAX price)
	avaxValueFloat := new(big.Float).Mul(new(big.Float).SetInt(avaxBalance), price)
	avaxValueInUSDC, _ := avaxValueFloat.Int(nil)

	// For BLACK token, we would need BLACK/USDC or BLACK/WAVAX price
	// For now, we'll skip BLACK in total value calculation or estimate it
	// TODO: Add BLACK price conversion when BLACK pool data is available
	blackValueInUSDC := big.NewInt(0)

	// Sum total value in USDC
	totalValue := new(big.Int).Add(usdcBalance, wavaxValueInUSDC)
	totalValue = new(big.Int).Add(totalValue, avaxValueInUSDC)
	totalValue = new(big.Int).Add(totalValue, blackValueInUSDC)

	// Calculate EstimatedAvax from TotalValue using current price
	// EstimatedAvax = TotalValue / price
	totalValueFloat := new(big.Float).SetInt(totalValue)
	estimatedAvaxFloat := new(big.Float).Quo(totalValueFloat, price)
	estimatedAvax, _ := estimatedAvaxFloat.Int(nil)

	snapshot := &types.CurrentAssetSnapshot{
		Timestamp:     time.Now(),
		CurrentState:  state,
		TotalValue:    totalValue,
		EstimatedAvax: estimatedAvax,
		AmountWavax:   wavaxBalance,
		AmountUsdc:    usdcBalance,
		AmountBlack:   blackBalance,
		AmountAvax:    avaxBalance,
	}

	return snapshot, nil
}

// sendReport records all StrategyReports and conditionally sends to the reporting channel
// Always records the report via TransactionRecorder
// Only sends to reportChan when stateChanged is true (state transition occurred)
// If the channel is full, the message is dropped to prevent strategy deadlock
// Implements non-blocking send pattern from research.md R5
func sendReport(reportChan chan<- string, report types.StrategyReport) {

	// Only send to channel if state changed
	if reportChan == nil {
		return
	}

	jsonStr, err := report.ToJSON()
	if err != nil {
		log.Printf("Failed to marshal strategy report: %v", err)
		return
	}

	reportChan <- jsonStr
}
