package util

import (
	"math/big"
	"testing"
)

// TestCalculateOptimalRangeWidthForCL1 tests the optimal range width calculation
func TestCalculateOptimalRangeWidthForCL1(t *testing.T) {
	// Test case: CL1 pool with imbalanced liquidity
	currentTick := int32(-253050)
	baseRangeWidth := 10 // Starting with 10 ticks
	tickSpacing := 200

	// Simulate prices
	sqrtPrice := TickToSqrtPriceX96(int(currentTick))

	// Test amounts (1 WAVAX = 18 decimals, 40 USDC = 6 decimals)
	maxWAVAX := new(big.Int)
	maxWAVAX.SetString("69232198285737839799", 10) // 1 WAVAX
	maxUSDC := new(big.Int)
	maxUSDC.SetString("933808849", 10) // 40 USDC

	utilizationThreshold := int64(90)
	maxIterations := 10

	t.Run("FindOptimalRange", func(t *testing.T) {
		tickLower, tickUpper, amount0, amount1, err := CalculateOptimalRangeWidthForCL1(
			currentTick,
			baseRangeWidth,
			tickSpacing,
			sqrtPrice,
			maxWAVAX,
			maxUSDC,
			utilizationThreshold,
			maxIterations,
		)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Calculate utilization
		utilization0 := new(big.Int).Mul(amount0, big.NewInt(100))
		utilization0.Div(utilization0, maxWAVAX)
		utilization1 := new(big.Int).Mul(amount1, big.NewInt(100))
		utilization1.Div(utilization1, maxUSDC)

		// Calculate base range for comparison
		baseLower, baseUpper, _ := CalculateTickBounds(currentTick, baseRangeWidth, tickSpacing)

		t.Logf("Base range: TickLower=%d, TickUpper=%d", baseLower, baseUpper)
		t.Logf("Optimal range: TickLower=%d, TickUpper=%d", tickLower, tickUpper)
		t.Logf("WAVAX utilization: %d%%", utilization0.Int64())
		t.Logf("USDC utilization: %d%%", utilization1.Int64())
		t.Logf("Amount0 (WAVAX): %s wei", amount0.String())
		t.Logf("Amount1 (USDC): %s units", amount1.String())

		// Verify tick bounds are valid
		if tickLower >= tickUpper {
			t.Errorf("TickLower (%d) must be < TickUpper (%d)", tickLower, tickUpper)
		}
	})

	t.Run("CompareUtilization", func(t *testing.T) {
		// Calculate with base range
		baseLower, baseUpper, err := CalculateTickBounds(currentTick, baseRangeWidth, tickSpacing)
		if err != nil {
			t.Fatalf("Failed to calculate base tick bounds: %v", err)
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

		// Calculate with optimal range
		optLower, optUpper, optAmount0, optAmount1, err := CalculateOptimalRangeWidthForCL1(
			currentTick,
			baseRangeWidth,
			tickSpacing,
			sqrtPrice,
			maxWAVAX,
			maxUSDC,
			utilizationThreshold,
			maxIterations,
		)
		if err != nil {
			t.Fatalf("Failed to calculate optimal range: %v", err)
		}

		optUtil0 := new(big.Int).Mul(optAmount0, big.NewInt(100))
		optUtil0.Div(optUtil0, maxWAVAX)
		optUtil1 := new(big.Int).Mul(optAmount1, big.NewInt(100))
		optUtil1.Div(optUtil1, maxUSDC)

		t.Logf("\n=== Base Range ===")
		t.Logf("Ticks: Lower=%d, Upper=%d", baseLower, baseUpper)
		t.Logf("WAVAX: %d%% utilized (%s wei)", baseUtil0.Int64(), baseAmount0.String())
		t.Logf("USDC: %d%% utilized (%s units)", baseUtil1.Int64(), baseAmount1.String())

		t.Logf("\n=== Optimal Range ===")
		t.Logf("Ticks: Lower=%d, Upper=%d", optLower, optUpper)
		t.Logf("WAVAX: %d%% utilized (%s wei)", optUtil0.Int64(), optAmount0.String())
		t.Logf("USDC: %d%% utilized (%s units)", optUtil1.Int64(), optAmount1.String())

		// Calculate waste reduction
		baseWaste0 := new(big.Int).Sub(maxWAVAX, baseAmount0)
		baseWaste1 := new(big.Int).Sub(maxUSDC, baseAmount1)
		optWaste0 := new(big.Int).Sub(maxWAVAX, optAmount0)
		optWaste1 := new(big.Int).Sub(maxUSDC, optAmount1)

		t.Logf("\n=== Waste Reduction ===")
		t.Logf("WAVAX waste: %s -> %s wei", baseWaste0.String(), optWaste0.String())
		t.Logf("USDC waste: %s -> %s units", baseWaste1.String(), optWaste1.String())

		// Optimal range should improve utilization
		minBaseUtil := baseUtil0
		if baseUtil1.Cmp(minBaseUtil) < 0 {
			minBaseUtil = baseUtil1
		}
		minOptUtil := optUtil0
		if optUtil1.Cmp(minOptUtil) < 0 {
			minOptUtil = optUtil1
		}

		if minOptUtil.Cmp(minBaseUtil) < 0 {
			t.Errorf("Optimal range should not decrease utilization. Base min: %d%%, Opt min: %d%%",
				minBaseUtil.Int64(), minOptUtil.Int64())
		}
	})
}

// TestCalculateOptimalRangeWidthForCL1_EdgeCases tests edge cases
func TestCalculateOptimalRangeWidthForCL1_EdgeCases(t *testing.T) {
	currentTick := int32(-249587)
	tickSpacing := 200
	sqrtPrice := TickToSqrtPriceX96(int(currentTick))

	t.Run("AlreadyOptimal", func(t *testing.T) {
		// Create balanced amounts that already meet the threshold
		maxWAVAX := new(big.Int)
		maxWAVAX.SetString("1000000000000000000", 10)
		maxUSDC := new(big.Int)
		maxUSDC.SetString("30000000", 10)

		tickLower, tickUpper, _, _, err := CalculateOptimalRangeWidthForCL1(
			currentTick,
			20, // Start with wider range
			tickSpacing,
			sqrtPrice,
			maxWAVAX,
			maxUSDC,
			90,
			5,
		)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		t.Logf("Optimal ticks when already efficient: Lower=%d, Upper=%d", tickLower, tickUpper)
	})

	t.Run("ExtremeImbalance", func(t *testing.T) {
		// Very imbalanced amounts
		maxWAVAX := new(big.Int)
		maxWAVAX.SetString("10000000000000000000", 10) // 10 WAVAX
		maxUSDC := new(big.Int)
		maxUSDC.SetString("1000000", 10) // 1 USDC

		tickLower, tickUpper, amount0, amount1, err := CalculateOptimalRangeWidthForCL1(
			currentTick,
			10,
			tickSpacing,
			sqrtPrice,
			maxWAVAX,
			maxUSDC,
			90,
			10,
		)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		utilization0 := new(big.Int).Mul(amount0, big.NewInt(100))
		utilization0.Div(utilization0, maxWAVAX)
		utilization1 := new(big.Int).Mul(amount1, big.NewInt(100))
		utilization1.Div(utilization1, maxUSDC)

		t.Logf("Extreme imbalance - Optimal ticks: Lower=%d, Upper=%d", tickLower, tickUpper)
		t.Logf("WAVAX utilization: %d%%", utilization0.Int64())
		t.Logf("USDC utilization: %d%%", utilization1.Int64())
	})
}
