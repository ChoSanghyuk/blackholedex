package util

import (
	"math/big"
	"testing"
)

// TestPriceMovementSimulation simulates how total assets change when price moves
// Initial conditions:
// - Total assets: 1000 USD
// - Current tick: -251400
// - Liquidity width: 6 (6 * 200 tick spacing = 1200 ticks)
// - Initial split: 50/50 AVAX/USDC by value
func TestPriceMovementSimulation(t *testing.T) {
	// Initial parameters
	totalAssetsUSD := big.NewFloat(1000)
	currentTick := int32(-251400)
	tickSpacing := 200
	rangeWidth := 6

	// Calculate tick bounds
	tickLower, tickUpper, err := CalculateTickBounds(currentTick, rangeWidth, tickSpacing)
	if err != nil {
		t.Fatalf("Failed to calculate tick bounds: %v", err)
	}

	t.Logf("═══════════════════════════════════════════════════")
	t.Logf("INITIAL SETUP")
	t.Logf("═══════════════════════════════════════════════════")
	t.Logf("Total Assets: $%.2f USD", totalAssetsUSD)
	t.Logf("Current Tick: %d", currentTick)
	t.Logf("Tick Lower: %d", tickLower)
	t.Logf("Tick Upper: %d", tickUpper)
	t.Logf("Range Width: %d ticks (%d * %d)", rangeWidth*tickSpacing, rangeWidth, tickSpacing)

	// Get sqrt prices for current, lower, and upper bounds
	sqrtPriceCurrent := TickToSqrtPriceX96(int(currentTick))
	sqrtPriceLower := TickToSqrtPriceX96(int(tickLower))
	sqrtPriceUpper := TickToSqrtPriceX96(int(tickUpper))

	// Convert to human-readable prices (USDC per AVAX)
	// Adjustment factor for decimals: AVAX (18) / USDC (6) = 10^12
	decimalAdjustment := new(big.Float).SetInt64(1_000_000_000_000)

	priceCurrent := new(big.Float).Mul(SqrtPriceToPrice(sqrtPriceCurrent), decimalAdjustment)
	priceLower := new(big.Float).Mul(SqrtPriceToPrice(sqrtPriceLower), decimalAdjustment)
	priceUpper := new(big.Float).Mul(SqrtPriceToPrice(sqrtPriceUpper), decimalAdjustment)

	t.Logf("\n")
	t.Logf("PRICE BOUNDS")
	t.Logf("Current Price: %.4f USDC per AVAX", priceCurrent)
	t.Logf("Lower Bound:   %.4f USDC per AVAX", priceLower)
	t.Logf("Upper Bound:   %.4f USDC per AVAX", priceUpper)

	// Calculate initial 50/50 split
	halfAssets := new(big.Float).Quo(totalAssetsUSD, big.NewFloat(2))

	// Amount of AVAX = half assets / current_price
	avaxAmount := new(big.Float).Quo(halfAssets, priceCurrent)
	usdcAmount := halfAssets

	// Convert to token units (AVAX: 18 decimals, USDC: 6 decimals)
	avaxWei := new(big.Float).Mul(avaxAmount, big.NewFloat(1e18))
	usdcUnits := new(big.Float).Mul(usdcAmount, big.NewFloat(1e6))

	avaxWeiBigInt := new(big.Int)
	usdcUnitsBigInt := new(big.Int)
	avaxWei.Int(avaxWeiBigInt)
	usdcUnits.Int(usdcUnitsBigInt)

	t.Logf("\n")
	t.Logf("INITIAL TOKEN AMOUNTS (50/50 split)")
	t.Logf("AVAX: %.6f AVAX ($%.2f)", avaxAmount, halfAssets)
	t.Logf("USDC: %.2f USDC ($%.2f)", usdcAmount, halfAssets)

	// Calculate liquidity and actual amounts deposited
	amount0Deposited, amount1Deposited, liquidity := ComputeAmounts(
		sqrtPriceCurrent,
		int(currentTick),
		int(tickLower),
		int(tickUpper),
		avaxWeiBigInt,
		usdcUnitsBigInt,
	)

	// Convert deposited amounts back to human-readable
	avaxDepositedFloat := new(big.Float).Quo(new(big.Float).SetInt(amount0Deposited), big.NewFloat(1e18))
	usdcDepositedFloat := new(big.Float).Quo(new(big.Float).SetInt(amount1Deposited), big.NewFloat(1e6))

	// Calculate remaining (unused) tokens
	avaxRemaining := new(big.Int).Sub(avaxWeiBigInt, amount0Deposited)
	usdcRemaining := new(big.Int).Sub(usdcUnitsBigInt, amount1Deposited)

	avaxRemainingFloat := new(big.Float).Quo(new(big.Float).SetInt(avaxRemaining), big.NewFloat(1e18))
	usdcRemainingFloat := new(big.Float).Quo(new(big.Float).SetInt(usdcRemaining), big.NewFloat(1e6))

	// Calculate value of deposited amounts
	avaxDepositedValue := new(big.Float).Mul(avaxDepositedFloat, priceCurrent)
	totalDepositedValue := new(big.Float).Add(avaxDepositedValue, usdcDepositedFloat)

	// Calculate value of remaining amounts
	avaxRemainingValue := new(big.Float).Mul(avaxRemainingFloat, priceCurrent)
	totalRemainingValue := new(big.Float).Add(avaxRemainingValue, usdcRemainingFloat)

	// Total value = deposited + remaining
	totalValueInitial := new(big.Float).Add(totalDepositedValue, totalRemainingValue)

	t.Logf("\n")
	t.Logf("AMOUNTS DEPOSITED INTO LIQUIDITY POOL")
	t.Logf("AVAX: %.6f AVAX ($%.2f)", avaxDepositedFloat, avaxDepositedValue)
	t.Logf("USDC: %.2f USDC ($%.2f)", usdcDepositedFloat, usdcDepositedFloat)
	t.Logf("Total Deposited Value: $%.2f USD", totalDepositedValue)
	t.Logf("Liquidity (L): %s", liquidity.String())

	t.Logf("\n")
	t.Logf("REMAINING (UNUSED) TOKENS")
	t.Logf("AVAX: %.6f AVAX ($%.2f)", avaxRemainingFloat, avaxRemainingValue)
	t.Logf("USDC: %.2f USDC ($%.2f)", usdcRemainingFloat, usdcRemainingFloat)
	t.Logf("Total Remaining Value: $%.2f USD", totalRemainingValue)

	t.Logf("\n")
	t.Logf("TOTAL ASSETS (Deposited + Remaining)")
	t.Logf("Total Value: $%.2f USD", totalValueInitial)

	// Case 1: Price moves DOWN to lower bound
	t.Run("Case1_PriceMovesDownToLowerBound", func(t *testing.T) {
		t.Logf("\n")
		t.Logf("═══════════════════════════════════════════════════")
		t.Logf("CASE 1: PRICE MOVES DOWN TO LOWER BOUND")
		t.Logf("═══════════════════════════════════════════════════")
		t.Logf("Price: %.4f → %.4f USDC per AVAX", priceCurrent, priceLower)
		t.Logf("─────────────────────────────────────────────────")

		// At lower bound, all liquidity becomes AVAX (amount0)
		amount0AtLower, amount1AtLower, err := CalculateTokenAmountsFromLiquidity(
			liquidity,
			sqrtPriceLower,
			tickLower,
			tickUpper,
		)
		if err != nil {
			t.Fatalf("Failed to calculate amounts at lower bound: %v", err)
		}

		// Convert to human-readable
		avaxAtLowerFloat := new(big.Float).Quo(new(big.Float).SetInt(amount0AtLower), big.NewFloat(1e18))
		usdcAtLowerFloat := new(big.Float).Quo(new(big.Float).SetInt(amount1AtLower), big.NewFloat(1e6))

		// Calculate value of LP position at lower price
		avaxValueAtLower := new(big.Float).Mul(avaxAtLowerFloat, priceLower)
		lpValueAtLower := new(big.Float).Add(avaxValueAtLower, usdcAtLowerFloat)

		// Add remaining tokens (their value also changes with price)
		remainingAvaxValue := new(big.Float).Mul(avaxRemainingFloat, priceLower)
		remainingValueAtLower := new(big.Float).Add(remainingAvaxValue, usdcRemainingFloat)

		// Total value = LP value + remaining value
		totalValueAtLower := new(big.Float).Add(lpValueAtLower, remainingValueAtLower)

		// Calculate impermanent loss
		impermanentLoss := new(big.Float).Sub(totalValueInitial, totalValueAtLower)
		impermanentLossPercent := new(big.Float).Quo(
			new(big.Float).Mul(impermanentLoss, big.NewFloat(100)),
			totalValueInitial,
		)

		t.Logf("\n")
		t.Logf("LP POSITION AT LOWER BOUND")
		t.Logf("AVAX: %.6f AVAX ($%.2f)", avaxAtLowerFloat, avaxValueAtLower)
		t.Logf("USDC: %.2f USDC ($%.2f)", usdcAtLowerFloat, usdcAtLowerFloat)
		t.Logf("LP Value: $%.2f USD", lpValueAtLower)
		t.Logf("\n")
		t.Logf("REMAINING TOKENS AT LOWER PRICE")
		t.Logf("AVAX: %.6f AVAX ($%.2f)", avaxRemainingFloat, remainingAvaxValue)
		t.Logf("USDC: %.2f USDC ($%.2f)", usdcRemainingFloat, usdcRemainingFloat)
		t.Logf("Remaining Value: $%.2f USD", remainingValueAtLower)
		t.Logf("\n")
		t.Logf("TOTAL VALUE AT LOWER BOUND")
		t.Logf("Total Assets: $%.2f USD (LP + Remaining)", totalValueAtLower)
		t.Logf("\n")
		t.Logf("IMPERMANENT LOSS ANALYSIS")
		t.Logf("Initial Value:  $%.2f USD", totalValueInitial)
		t.Logf("Current Value:  $%.2f USD", totalValueAtLower)
		t.Logf("Loss:           $%.2f USD (%.2f%%)", impermanentLoss, impermanentLossPercent)

		// Additional insight: what if you just held 50/50?
		holdAvaxValue := new(big.Float).Mul(avaxDepositedFloat, priceLower)
		holdTotalValue := new(big.Float).Add(holdAvaxValue, usdcDepositedFloat)

		t.Logf("\n")
		t.Logf("COMPARISON: IF YOU JUST HELD 50/50 (NO LP)")
		t.Logf("AVAX: %.6f AVAX ($%.2f)", avaxDepositedFloat, holdAvaxValue)
		t.Logf("USDC: %.2f USDC ($%.2f)", usdcDepositedFloat, usdcDepositedFloat)
		t.Logf("Total Value: $%.2f USD", holdTotalValue)

		lpVsHold := new(big.Float).Sub(totalValueAtLower, holdTotalValue)
		t.Logf("\n")
		t.Logf("LP vs HOLD Difference: $%.2f USD", lpVsHold)
		t.Logf("═══════════════════════════════════════════════════")
	})

	// Case 2: Price moves UP to upper bound
	t.Run("Case2_PriceMovesUpToUpperBound", func(t *testing.T) {
		t.Logf("\n")
		t.Logf("═══════════════════════════════════════════════════")
		t.Logf("CASE 2: PRICE MOVES UP TO UPPER BOUND")
		t.Logf("═══════════════════════════════════════════════════")
		t.Logf("Price: %.4f → %.4f USDC per AVAX", priceCurrent, priceUpper)
		t.Logf("─────────────────────────────────────────────────")

		// At upper bound, all liquidity becomes USDC (amount1)
		amount0AtUpper, amount1AtUpper, err := CalculateTokenAmountsFromLiquidity(
			liquidity,
			sqrtPriceUpper,
			tickLower,
			tickUpper,
		)
		if err != nil {
			t.Fatalf("Failed to calculate amounts at upper bound: %v", err)
		}

		// Convert to human-readable
		avaxAtUpperFloat := new(big.Float).Quo(new(big.Float).SetInt(amount0AtUpper), big.NewFloat(1e18))
		usdcAtUpperFloat := new(big.Float).Quo(new(big.Float).SetInt(amount1AtUpper), big.NewFloat(1e6))

		// Calculate value of LP position at upper price
		avaxValueAtUpper := new(big.Float).Mul(avaxAtUpperFloat, priceUpper)
		lpValueAtUpper := new(big.Float).Add(avaxValueAtUpper, usdcAtUpperFloat)

		// Add remaining tokens (their value also changes with price)
		remainingAvaxValue := new(big.Float).Mul(avaxRemainingFloat, priceUpper)
		remainingValueAtUpper := new(big.Float).Add(remainingAvaxValue, usdcRemainingFloat)

		// Total value = LP value + remaining value
		totalValueAtUpper := new(big.Float).Add(lpValueAtUpper, remainingValueAtUpper)

		// Calculate impermanent loss
		impermanentLoss := new(big.Float).Sub(totalValueInitial, totalValueAtUpper)
		impermanentLossPercent := new(big.Float).Quo(
			new(big.Float).Mul(impermanentLoss, big.NewFloat(100)),
			totalValueInitial,
		)

		t.Logf("\n")
		t.Logf("LP POSITION AT UPPER BOUND")
		t.Logf("AVAX: %.6f AVAX ($%.2f)", avaxAtUpperFloat, avaxValueAtUpper)
		t.Logf("USDC: %.2f USDC ($%.2f)", usdcAtUpperFloat, usdcAtUpperFloat)
		t.Logf("LP Value: $%.2f USD", lpValueAtUpper)
		t.Logf("\n")
		t.Logf("REMAINING TOKENS AT UPPER PRICE")
		t.Logf("AVAX: %.6f AVAX ($%.2f)", avaxRemainingFloat, remainingAvaxValue)
		t.Logf("USDC: %.2f USDC ($%.2f)", usdcRemainingFloat, usdcRemainingFloat)
		t.Logf("Remaining Value: $%.2f USD", remainingValueAtUpper)
		t.Logf("\n")
		t.Logf("TOTAL VALUE AT UPPER BOUND")
		t.Logf("Total Assets: $%.2f USD (LP + Remaining)", totalValueAtUpper)
		t.Logf("\n")
		t.Logf("IMPERMANENT LOSS ANALYSIS")
		t.Logf("Initial Value:  $%.2f USD", totalValueInitial)
		t.Logf("Current Value:  $%.2f USD", totalValueAtUpper)
		t.Logf("Gain/Loss:      $%.2f USD (%.2f%%)", new(big.Float).Neg(impermanentLoss), new(big.Float).Neg(impermanentLossPercent))

		// Additional insight: what if you just held 50/50?
		holdAvaxValue := new(big.Float).Mul(avaxDepositedFloat, priceUpper)
		holdTotalValue := new(big.Float).Add(holdAvaxValue, usdcDepositedFloat)

		t.Logf("\n")
		t.Logf("COMPARISON: IF YOU JUST HELD 50/50 (NO LP)")
		t.Logf("AVAX: %.6f AVAX ($%.2f)", avaxDepositedFloat, holdAvaxValue)
		t.Logf("USDC: %.2f USDC ($%.2f)", usdcDepositedFloat, usdcDepositedFloat)
		t.Logf("Total Value: $%.2f USD", holdTotalValue)

		lpVsHold := new(big.Float).Sub(totalValueAtUpper, holdTotalValue)
		t.Logf("\n")
		t.Logf("LP vs HOLD Difference: $%.2f USD", lpVsHold)
		t.Logf("═══════════════════════════════════════════════════")
	})

	// Case 3: Price moves down to lower bound, rebalance to 50/50, then price returns to original
	t.Run("Case3_DownRebalanceAndReturn", func(t *testing.T) {
		t.Logf("\n")
		t.Logf("═══════════════════════════════════════════════════")
		t.Logf("CASE 3: DOWN → REBALANCE → RETURN TO ORIGINAL")
		t.Logf("═══════════════════════════════════════════════════")
		t.Logf("Step 1: Price %.4f → %.4f (lower bound)", priceCurrent, priceLower)
		t.Logf("Step 2: Withdraw and rebalance to 50/50")
		t.Logf("Step 3: Price %.4f → %.4f (back to original)", priceLower, priceCurrent)
		t.Logf("─────────────────────────────────────────────────")

		// Step 1: Get amounts at lower bound
		amount0AtLower, amount1AtLower, err := CalculateTokenAmountsFromLiquidity(
			liquidity,
			sqrtPriceLower,
			tickLower,
			tickUpper,
		)
		if err != nil {
			t.Fatalf("Failed to calculate amounts at lower bound: %v", err)
		}

		avaxAtLowerFloat := new(big.Float).Quo(new(big.Float).SetInt(amount0AtLower), big.NewFloat(1e18))
		usdcAtLowerFloat := new(big.Float).Quo(new(big.Float).SetInt(amount1AtLower), big.NewFloat(1e6))

		// Add remaining tokens
		totalAvaxAtLower := new(big.Float).Add(avaxAtLowerFloat, avaxRemainingFloat)
		totalUsdcAtLower := new(big.Float).Add(usdcAtLowerFloat, usdcRemainingFloat)

		avaxValueAtLower := new(big.Float).Mul(totalAvaxAtLower, priceLower)
		usdcValueAtLower := totalUsdcAtLower
		totalValueAtLower := new(big.Float).Add(avaxValueAtLower, usdcValueAtLower)

		t.Logf("\n")
		t.Logf("STEP 1: TOTAL POSITION AT LOWER BOUND (LP + Remaining)")
		t.Logf("AVAX: %.6f AVAX ($%.2f)", totalAvaxAtLower, avaxValueAtLower)
		t.Logf("USDC: %.2f USDC ($%.2f)", totalUsdcAtLower, totalUsdcAtLower)
		t.Logf("Total Value: $%.2f USD", totalValueAtLower)

		// Step 2: Rebalance to 50/50 at lower price
		halfValueAtLower := new(big.Float).Quo(totalValueAtLower, big.NewFloat(2))
		avaxRebalanced := new(big.Float).Quo(halfValueAtLower, priceLower)
		usdcRebalanced := halfValueAtLower

		t.Logf("\n")
		t.Logf("STEP 2: REBALANCE TO 50/50 AT LOWER PRICE")
		t.Logf("AVAX: %.6f AVAX ($%.2f)", avaxRebalanced, halfValueAtLower)
		t.Logf("USDC: %.2f USDC ($%.2f)", usdcRebalanced, halfValueAtLower)
		t.Logf("Total Value: $%.2f USD", totalValueAtLower)

		// Convert rebalanced amounts to token units
		avaxRebalancedWei := new(big.Float).Mul(avaxRebalanced, big.NewFloat(1e18))
		usdcRebalancedUnits := new(big.Float).Mul(usdcRebalanced, big.NewFloat(1e6))

		avaxRebalancedBigInt := new(big.Int)
		usdcRebalancedBigInt := new(big.Int)
		avaxRebalancedWei.Int(avaxRebalancedBigInt)
		usdcRebalancedUnits.Int(usdcRebalancedBigInt)

		// Calculate new tick range
		tickLowerNew, tickUpperNew, err := CalculateTickBounds(tickLower, rangeWidth, tickSpacing)

		// Calculate new liquidity with rebalanced amounts at lower bound
		amount0DepositedNew, amount1DepositedNew, liquidityRebalanced := ComputeAmounts(
			sqrtPriceLower,
			int(tickLower),
			int(tickLowerNew),
			int(tickUpperNew),
			avaxRebalancedBigInt,
			usdcRebalancedBigInt,
		)

		// Calculate remaining tokens after rebalancing
		avaxRemainingNew := new(big.Int).Sub(avaxRebalancedBigInt, amount0DepositedNew)
		usdcRemainingNew := new(big.Int).Sub(usdcRebalancedBigInt, amount1DepositedNew)

		avaxRemainingNewFloat := new(big.Float).Quo(new(big.Float).SetInt(avaxRemainingNew), big.NewFloat(1e18))
		usdcRemainingNewFloat := new(big.Float).Quo(new(big.Float).SetInt(usdcRemainingNew), big.NewFloat(1e6))

		t.Logf("New Liquidity (L): %s", liquidityRebalanced.String())
		t.Logf("New Remaining - AVAX: %.6f, USDC: %.2f", avaxRemainingNewFloat, usdcRemainingNewFloat)

		// Step 3: Price returns to original
		amount0AtReturn, amount1AtReturn, err := CalculateTokenAmountsFromLiquidity(
			liquidityRebalanced,
			sqrtPriceCurrent,
			tickLowerNew,
			tickUpperNew,
		)
		if err != nil {
			t.Fatalf("Failed to calculate amounts at return: %v", err)
		}

		avaxAtReturnFloat := new(big.Float).Quo(new(big.Float).SetInt(amount0AtReturn), big.NewFloat(1e18))
		usdcAtReturnFloat := new(big.Float).Quo(new(big.Float).SetInt(amount1AtReturn), big.NewFloat(1e6))

		// Total tokens = LP position + remaining
		totalAvaxAtReturn := new(big.Float).Add(avaxAtReturnFloat, avaxRemainingNewFloat)
		totalUsdcAtReturn := new(big.Float).Add(usdcAtReturnFloat, usdcRemainingNewFloat)

		avaxValueAtReturn := new(big.Float).Mul(totalAvaxAtReturn, priceCurrent)
		usdcValueAtReturn := totalUsdcAtReturn
		totalValueAtReturn := new(big.Float).Add(avaxValueAtReturn, usdcValueAtReturn)

		t.Logf("\n")
		t.Logf("STEP 3: PRICE RETURNS TO ORIGINAL")
		t.Logf("LP Position - AVAX: %.6f AVAX, USDC: %.2f USDC", avaxAtReturnFloat, usdcAtReturnFloat)
		t.Logf("Remaining   - AVAX: %.6f AVAX, USDC: %.2f USDC", avaxRemainingNewFloat, usdcRemainingNewFloat)
		t.Logf("Total       - AVAX: %.6f AVAX ($%.2f)", totalAvaxAtReturn, avaxValueAtReturn)
		t.Logf("              USDC: %.2f USDC ($%.2f)", totalUsdcAtReturn, totalUsdcAtReturn)
		t.Logf("Total Value: $%.2f USD", totalValueAtReturn)

		t.Logf("\n")
		t.Logf("FINAL COMPARISON")
		t.Logf("Initial Value:        $%.2f USD", totalValueInitial)
		t.Logf("After Down+Rebalance: $%.2f USD", totalValueAtLower)
		t.Logf("After Return:         $%.2f USD", totalValueAtReturn)

		netChange := new(big.Float).Sub(totalValueAtReturn, totalValueInitial)
		netChangePercent := new(big.Float).Quo(
			new(big.Float).Mul(netChange, big.NewFloat(100)),
			totalValueInitial,
		)

		t.Logf("\n")
		t.Logf("NET CHANGE: $%.2f USD (%.2f%%)", netChange, netChangePercent)
		t.Logf("═══════════════════════════════════════════════════")
	})
}
