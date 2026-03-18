package util

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ChoSanghyuk/blackholedex/pkg/types"
)

// Validation and helper functions for liquidity staking operations

// ValidateStakingRequest validates input parameters for staking operation
// Returns error if validation fails, nil otherwise
func ValidateStakingRequest(maxWAVAX, maxUSDC *big.Int, rangeWidth, slippagePct int) error {
	// Range width validation (1-20 tick ranges)
	// if rangeWidth <= 0 || rangeWidth > 20 {
	// 	return fmt.Errorf("range width must be between 1 and 20, got %d (valid examples: 2, 6, 10)", rangeWidth)
	// }

	// Slippage validation (1-50 percent)
	if slippagePct <= 0 || slippagePct > 50 {
		return fmt.Errorf("slippage tolerance must be between 1 and 50 percent, got %d", slippagePct)
	}

	// Amount validation
	if maxWAVAX == nil || maxWAVAX.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("maxWAVAX must be > 0")
	}
	if maxUSDC == nil || maxUSDC.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("maxUSDC must be > 0")
	}

	return nil
}

// CalculateTickBounds calculates tick bounds from current tick and range width
// rangeWidth N means ±(N/2) tick ranges from current tick
// Returns tickLower, tickUpper, or error if bounds invalid
// Edge case: For extreme ticks near ±887272, bounds are clamped to valid range
func CalculateTickBounds(currentTick int32, rangeWidth int, tickSpacing int) (int32, int32, error) {
	const maxTick = 887272

	halfWidth := rangeWidth / 2
	// tickIndex := int(currentTick) / tickSpacing
	tickIndex := int(math.Round(float64(currentTick) / float64(tickSpacing)))

	// Calculate raw bounds
	rawTickLower := (tickIndex - halfWidth) * tickSpacing
	rawTickUpper := (tickIndex + halfWidth) * tickSpacing

	// Clamp to valid tick range for edge cases near ±maxTick
	// This handles extreme ticks where calculated bounds would exceed limits
	tickLower := int32(rawTickLower)
	tickUpper := int32(rawTickUpper)

	if tickLower < -maxTick {
		tickLower = -maxTick
	}
	if tickLower > maxTick {
		tickLower = maxTick
	}
	if tickUpper < -maxTick {
		tickUpper = -maxTick
	}
	if tickUpper > maxTick {
		tickUpper = maxTick
	}

	// Validate tickLower < tickUpper (should always be true after clamping)
	if tickLower >= tickUpper {
		return 0, 0, fmt.Errorf("tickLower (%d) must be < tickUpper (%d) - current tick %d with range width %d creates invalid bounds", tickLower, tickUpper, currentTick, rangeWidth)
	}

	return tickLower, tickUpper, nil
}

// CalculateOptimalRangeWidthForCL1 calculates optimal tick bounds to minimize wasted tokens
// This function is specifically for CL1 pools where wasted tokens are a concern
// It adjusts the range asymmetrically based on which token is underutilized
// utilizationThreshold is the minimum acceptable utilization percentage (e.g., 90 for 90%)
func CalculateOptimalRangeWidthForCL1(
	currentTick int32,
	baseRangeWidth int,
	tickSpacing int,
	sqrtPrice *big.Int,
	maxWAVAX *big.Int,
	maxUSDC *big.Int,
	utilizationThreshold int64,
	maxIterations int,
) (tickLower int32, tickUpper int32, amount0 *big.Int, amount1 *big.Int, err error) {
	const maxTick = 887272

	// Start with base range
	baseLower, baseUpper, calcErr := CalculateTickBounds(currentTick, baseRangeWidth, tickSpacing)
	if calcErr != nil {
		return 0, 0, nil, nil, fmt.Errorf("failed to calculate base tick bounds: %w", calcErr)
	}

	// Calculate initial amounts
	baseAmount0, baseAmount1, _ := ComputeAmounts(
		sqrtPrice,
		int(currentTick),
		int(baseLower),
		int(baseUpper),
		maxWAVAX,
		maxUSDC,
	)

	// Calculate utilization percentages
	utilization0 := new(big.Int).Mul(baseAmount0, big.NewInt(100))
	utilization0.Div(utilization0, maxWAVAX)
	utilization1 := new(big.Int).Mul(baseAmount1, big.NewInt(100))
	utilization1.Div(utilization1, maxUSDC)

	// If both are above threshold, no optimization needed
	if utilization0.Cmp(big.NewInt(utilizationThreshold)) >= 0 &&
		utilization1.Cmp(big.NewInt(utilizationThreshold)) >= 0 {
		return baseLower, baseUpper, baseAmount0, baseAmount1, nil
	}

	// Track best result
	bestLower := baseLower
	bestUpper := baseUpper
	bestAmount0 := baseAmount0
	bestAmount1 := baseAmount1
	bestMinUtilization := utilization0.Int64()
	if utilization1.Int64() < bestMinUtilization {
		bestMinUtilization = utilization1.Int64()
	}

	// Determine which token is underutilized and adjust accordingly
	// When current tick is in range [tickLower, tickUpper]:
	// - Extending tickUpper (higher) allows more token0 (WAVAX) to be used
	// - Extending tickLower (lower) allows more token1 (USDC) to be used

	scale := 5 // memo. CL1 에서만 사용할 것이기에 조정 단위를 5으로 고정.
	for i := 1; i <= maxIterations; i++ {
		var testLower, testUpper int32

		if utilization0.Int64() < utilizationThreshold {
			// WAVAX underutilized - extend upper bound to allow more WAVAX
			testLower = baseLower
			testUpper = baseUpper + int32(i*scale)
			if testUpper > maxTick {
				testUpper = maxTick
			}
		} else if utilization1.Int64() < utilizationThreshold {
			// USDC underutilized - extend lower bound to allow more USDC
			testLower = baseLower - int32(i*scale)
			if testLower < -maxTick {
				testLower = -maxTick
			}
			testUpper = baseUpper
		} else {
			// Both meet threshold
			break
		}

		// Validate bounds
		if testLower >= testUpper {
			continue
		}

		// Calculate amounts with adjusted range
		testAmount0, testAmount1, _ := ComputeAmounts(
			sqrtPrice,
			int(currentTick),
			int(testLower),
			int(testUpper),
			maxWAVAX,
			maxUSDC,
		)

		// Calculate new utilization
		testUtil0 := new(big.Int).Mul(testAmount0, big.NewInt(100))
		testUtil0.Div(testUtil0, maxWAVAX)
		testUtil1 := new(big.Int).Mul(testAmount1, big.NewInt(100))
		testUtil1.Div(testUtil1, maxUSDC)

		// Find minimum utilization
		minUtil := testUtil0.Int64()
		if testUtil1.Int64() < minUtil {
			minUtil = testUtil1.Int64()
		}

		// Update best if this is better
		if minUtil > bestMinUtilization {
			bestMinUtilization = minUtil
			bestLower = testLower
			bestUpper = testUpper
			bestAmount0 = testAmount0
			bestAmount1 = testAmount1
			utilization0 = testUtil0
			utilization1 = testUtil1
		}

		// If both tokens meet threshold, we're done
		if testUtil0.Cmp(big.NewInt(utilizationThreshold)) >= 0 &&
			testUtil1.Cmp(big.NewInt(utilizationThreshold)) >= 0 {
			return testLower, testUpper, testAmount0, testAmount1, nil
		}
	}

	return bestLower, bestUpper, bestAmount0, bestAmount1, nil
}

// CalculateMinAmount calculates minimum amount with slippage protection
// amountMin = amountDesired * (100 - slippagePct) / 100
func CalculateMinAmount(amountDesired *big.Int, slippagePct int) *big.Int {
	if amountDesired == nil {
		return big.NewInt(0)
	}

	// amountMin = amountDesired * (100 - slippagePct) / 100
	multiplier := big.NewInt(int64(100 - slippagePct))
	divisor := big.NewInt(100)

	result := new(big.Int).Mul(amountDesired, multiplier)
	result.Div(result, divisor)

	return result
}

// ExtractGasCost extracts gas cost from transaction receipt
// Returns gas cost in wei (GasUsed * EffectiveGasPrice)
func ExtractGasCost(receipt *types.TxReceipt) (*big.Int, error) {
	if receipt == nil {
		return nil, fmt.Errorf("receipt is nil")
	}

	// Parse GasUsed from string
	gasUsed := new(big.Int)
	if _, ok := gasUsed.SetString(receipt.GasUsed, 0); !ok {
		return nil, fmt.Errorf("failed to parse GasUsed: %s", receipt.GasUsed)
	}

	// Parse EffectiveGasPrice from string
	gasPrice := new(big.Int)
	if _, ok := gasPrice.SetString(receipt.EffectiveGasPrice, 0); !ok {
		return nil, fmt.Errorf("failed to parse EffectiveGasPrice: %s", receipt.EffectiveGasPrice)
	}

	// Calculate gas cost
	gasCost := new(big.Int).Mul(gasUsed, gasPrice)

	return gasCost, nil
}

// deprecated. no critical error
// IsCriticalError determines if an error is critical and requires immediate halt (T015)
func IsCriticalError(err error) bool {
	// if err == nil {
	// 	return false
	// }

	// errStr := strings.ToLower(err.Error())

	// // Critical error patterns that require immediate halt
	// criticalPatterns := []string{
	// 	"insufficient balance",
	// 	"insufficient funds",
	// 	"nft not owned",
	// 	"not owner",
	// 	"transaction reverted",
	// 	"execution reverted",
	// 	"invalid position state",
	// 	"position does not exist",
	// 	"unauthorized",
	// 	"contract paused",
	// }

	// for _, pattern := range criticalPatterns {
	// 	if strings.Contains(errStr, pattern) {
	// 		return true
	// 	}
	// }

	return false
}
