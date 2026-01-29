package util

import (
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqrtPriceToPrice(t *testing.T) {

	val, _ := big.NewInt(0).SetString("267326922672530907272725", 0)
	priceRaw := SqrtPriceToPrice(val)

	// expected, _ := big.NewInt(0).SetString("304011615425126403287043", 10)
	// assert.Equal(t, expected, sqrtPrice)
	price, _ := priceRaw.Float64()

	fmt.Printf("%v\n", val)
	fmt.Printf("%v\n", priceRaw)
	fmt.Printf("%v\n", price)
}

func TestCalculateRebalanceAmounts(t *testing.T) {

	// 1AVAX = 12.49 USDC일 때의 값
	sqrtPrice, _ := big.NewInt(0).SetString("280057970020625981233062", 0)
	// price := 12.49

	t.Run("USDC_TO_WAVAX", func(t *testing.T) {
		wavaxBalance := big.NewInt(2 * 1000000000000000000) // 2AVAX. 25USDC
		usdcBalance := big.NewInt(50000000)                 // 50 USDC

		tokenToSwap, swapAmount, err := CalculateRebalanceAmounts(
			wavaxBalance,
			usdcBalance,
			sqrtPrice,
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, tokenToSwap)
		fmt.Printf("tokenToSwap : %v, swapAmount: %v\n", tokenToSwap, swapAmount)
	})

	t.Run("WAVX_TO_USDC", func(t *testing.T) {
		wavaxBalance := big.NewInt(5 * 1000000000000000000) // 5AVAX. 62.5 USDC
		usdcBalance := big.NewInt(50000000)                 // 50 USDC

		tokenToSwap, swapAmount, err := CalculateRebalanceAmounts(
			wavaxBalance,
			usdcBalance,
			sqrtPrice,
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, tokenToSwap)
		fmt.Printf("tokenToSwap : %v, swapAmount: %v\n", tokenToSwap, swapAmount)
	})
}

// CalculateTickBounds + TickToSqrtPriceX96 + SqrtPriceToPrice
func TestCalculatePriceBounds(t *testing.T) {

	var currentTick int32 = -249587 // GetAMMState 결과
	rangeWidth := 2
	tickSpacing := 200
	// T014: Calculate tick bounds
	tickLower, tickUpper, err := CalculateTickBounds(currentTick, rangeWidth, tickSpacing)
	if err != nil {

	}
	log.Printf("CurrentTick: %d,TickLower: %d, TickUpper: %d", currentTick, tickLower, tickUpper)

	currentSqrtPrice := TickToSqrtPriceX96(int(currentTick))
	lowerSqrtPrice := TickToSqrtPriceX96(int(tickLower))
	upperSqrtPrice := TickToSqrtPriceX96(int(tickUpper))

	decimalAdjustment := new(big.Float).SetInt64(1_000_000_000_000) // 10^12
	currentPrice := new(big.Float).Mul(decimalAdjustment, SqrtPriceToPrice(currentSqrtPrice))
	lowerPrice := new(big.Float).Mul(decimalAdjustment, SqrtPriceToPrice(lowerSqrtPrice))
	upperPrice := new(big.Float).Mul(decimalAdjustment, SqrtPriceToPrice(upperSqrtPrice))

	log.Printf("PriceCurrent: %.02f, PriceLower: %.02f, PriceUpper: %.02f", currentPrice, lowerPrice, upperPrice)
}

/* -1247 -289400
2026/01/07 12:51:51 CurrentTick: -249587,TickLower: -249600, TickUpper: -249200
2026/01/07 12:51:51 PriceCurrent: 14.49, PriceLower: 14.47, PriceUpper: 15.06

2026/01/07 18:19:02 CurrentTick: -249587,TickLower: -249800, TickUpper: -249400
2026/01/07 18:19:02 PriceCurrent: 14.49, PriceLower: 14.19, PriceUpper: 14.77


2026/01/07 12:53:26 CurrentTick: -249587,TickLower: -249800, TickUpper: -249000
2026/01/07 12:53:26 PriceCurrent: 14.49, PriceLower: 14.19, PriceUpper: 15.37
*/
