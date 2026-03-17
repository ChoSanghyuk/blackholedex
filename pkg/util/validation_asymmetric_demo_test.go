package util

import (
	"math/big"
	"testing"
)

// TestAsymmetricOptimization_RealWorldScenario tests the exact scenario from user's log
// Log showed: CurrentTick: -253022, WAVAX 45%, USDC 99% with rangeWidth: 200, tickSpacing: 1
func TestAsymmetricOptimization_RealWorldScenario(t *testing.T) {
	// Real values from user's log
	currentTick := int32(-253022)
	baseRangeWidth := 200
	tickSpacing := 1

	// Simulate the pool state
	sqrtPrice := TickToSqrtPriceX96(int(currentTick))

	// Assume some token amounts that would give 45% WAVAX, 99% USDC utilization
	// These are example amounts - adjust based on your actual balances
	maxWAVAX := new(big.Int)
	maxWAVAX.SetString("10000000000000000000", 10) // 10 WAVAX
	maxUSDC := new(big.Int)
	maxUSDC.SetString("300000000", 10) // 300 USDC

	t.Run("ShowOptimizationImprovement", func(t *testing.T) {
		// Calculate base range (what user had)
		baseLower, baseUpper, err := CalculateTickBounds(currentTick, baseRangeWidth, tickSpacing)
		if err != nil {
			t.Fatalf("Failed to calculate base bounds: %v", err)
		}

		baseAmount0, baseAmount1, _ := ComputeAmounts(
			sqrtPrice,
			int(currentTick),
			int(baseLower),
			int(baseUpper),
			maxWAVAX,
			maxUSDC,
		)

		baseUtil0 := new(big.Int).Mul(baseAmount0, big.NewInt(100))
		baseUtil0.Div(baseUtil0, maxWAVAX)
		baseUtil1 := new(big.Int).Mul(baseAmount1, big.NewInt(100))
		baseUtil1.Div(baseUtil1, maxUSDC)

		t.Logf("\n━━━ BEFORE OPTIMIZATION ━━━")
		t.Logf("CurrentTick: %d", currentTick)
		t.Logf("Base Range: TickLower=%d, TickUpper=%d (width=%d)", baseLower, baseUpper, baseRangeWidth)
		t.Logf("WAVAX Utilization: %d%%", baseUtil0.Int64())
		t.Logf("USDC Utilization: %d%%", baseUtil1.Int64())
		t.Logf("WAVAX Amount: %s wei (%.4f WAVAX)", baseAmount0.String(), toWAVAX(baseAmount0))
		t.Logf("USDC Amount: %s units (%.2f USDC)", baseAmount1.String(), toUSDC(baseAmount1))

		wastedWAVAX := new(big.Int).Sub(maxWAVAX, baseAmount0)
		wastedUSDC := new(big.Int).Sub(maxUSDC, baseAmount1)
		t.Logf("Wasted WAVAX: %s wei (%.4f WAVAX)", wastedWAVAX.String(), toWAVAX(wastedWAVAX))
		t.Logf("Wasted USDC: %s units (%.2f USDC)", wastedUSDC.String(), toUSDC(wastedUSDC))

		// Now optimize
		optLower, optUpper, optAmount0, optAmount1, err := CalculateOptimalRangeWidthForCL1(
			currentTick,
			baseRangeWidth,
			tickSpacing,
			sqrtPrice,
			maxWAVAX,
			maxUSDC,
			90, // 90% threshold
			50, // Try up to 50 iterations
		)

		if err != nil {
			t.Fatalf("Optimization failed: %v", err)
		}

		optUtil0 := new(big.Int).Mul(optAmount0, big.NewInt(100))
		optUtil0.Div(optUtil0, maxWAVAX)
		optUtil1 := new(big.Int).Mul(optAmount1, big.NewInt(100))
		optUtil1.Div(optUtil1, maxUSDC)

		t.Logf("\n━━━ AFTER OPTIMIZATION ━━━")
		t.Logf("Optimal Range: TickLower=%d, TickUpper=%d", optLower, optUpper)
		t.Logf("Range Adjustment: Lower %+d ticks, Upper %+d ticks",
			optLower-baseLower, optUpper-baseUpper)
		t.Logf("WAVAX Utilization: %d%% (improved by %+d%%)", optUtil0.Int64(), optUtil0.Int64()-baseUtil0.Int64())
		t.Logf("USDC Utilization: %d%% (changed by %+d%%)", optUtil1.Int64(), optUtil1.Int64()-baseUtil1.Int64())
		t.Logf("WAVAX Amount: %s wei (%.4f WAVAX)", optAmount0.String(), toWAVAX(optAmount0))
		t.Logf("USDC Amount: %s units (%.2f USDC)", optAmount1.String(), toUSDC(optAmount1))

		optWastedWAVAX := new(big.Int).Sub(maxWAVAX, optAmount0)
		optWastedUSDC := new(big.Int).Sub(maxUSDC, optAmount1)
		t.Logf("Wasted WAVAX: %s wei (%.4f WAVAX)", optWastedWAVAX.String(), toWAVAX(optWastedWAVAX))
		t.Logf("Wasted USDC: %s units (%.2f USDC)", optWastedUSDC.String(), toUSDC(optWastedUSDC))

		t.Logf("\n━━━ IMPROVEMENT ━━━")
		wasteReductionWAVAX := new(big.Int).Sub(wastedWAVAX, optWastedWAVAX)
		wasteReductionUSDC := new(big.Int).Sub(wastedUSDC, optWastedUSDC)

		if wasteReductionWAVAX.Sign() > 0 {
			t.Logf("WAVAX waste reduced by: %.4f WAVAX", toWAVAX(wasteReductionWAVAX))
		}
		if wasteReductionUSDC.Sign() > 0 {
			t.Logf("USDC waste reduced by: %.2f USDC", toUSDC(wasteReductionUSDC))
		}

		// Verify improvement
		minBaseUtil := baseUtil0.Int64()
		if baseUtil1.Int64() < minBaseUtil {
			minBaseUtil = baseUtil1.Int64()
		}
		minOptUtil := optUtil0.Int64()
		if optUtil1.Int64() < minOptUtil {
			minOptUtil = optUtil1.Int64()
		}

		if minOptUtil <= minBaseUtil {
			t.Errorf("Expected optimization to improve min utilization from %d%% to >%d%%, got %d%%",
				minBaseUtil, minBaseUtil, minOptUtil)
		} else {
			t.Logf("✓ Minimum utilization improved from %d%% to %d%%", minBaseUtil, minOptUtil)
		}
	})
}

// Helper functions to convert wei/units to human-readable format
func toWAVAX(wei *big.Int) float64 {
	if wei == nil {
		return 0
	}
	f := new(big.Float).SetInt(wei)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	result := new(big.Float).Quo(f, divisor)
	val, _ := result.Float64()
	return val
}

func toUSDC(units *big.Int) float64 {
	if units == nil {
		return 0
	}
	f := new(big.Float).SetInt(units)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil))
	result := new(big.Float).Quo(f, divisor)
	val, _ := result.Float64()
	return val
}
