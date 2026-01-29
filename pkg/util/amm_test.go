package util

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

// memo. SqrtPrice는 두 tick 사이의 값이기에, safelyGetStateOfAMM 결과로 나오는 sprtPrice 랑 tick 완벽하게 매칭되지 않음
func TestTickToSqrtPriceX96(t *testing.T) {

	sqrtPrice := TickToSqrtPriceX96(-252000) // -249428

	expected, _ := big.NewInt(0).SetString("304011615425126403287043", 10)
	assert.Equal(t, expected, sqrtPrice)
}

func TestComputeAmounts(t *testing.T) {

	sqrtPriceX96, _ := big.NewInt(0).SetString("275467826341246019486853", 10)
	tick := -251400
	tickLower := -252000
	tickUpper := -250800
	amount0Max, _ := big.NewInt(0).SetString("99999309985252461722", 10)
	amount1Max, _ := big.NewInt(0).SetString("1208870000", 10)
	amount0, amount1, l := ComputeAmounts(sqrtPriceX96, tick, tickLower, tickUpper, amount0Max, amount1Max)

	t.Log("amount0:", amount0)
	t.Log("amount1:", amount1.String())
	t.Log("liquidity:", l)

	// // Verify we got non-zero results
	// assert.Greater(t, l.Cmp(big.NewInt(0)), 0, "liquidity should be > 0")
	// assert.Greater(t, amount0.Cmp(big.NewInt(0)), -1, "amount0 should be >= 0")
	// assert.Greater(t, amount1.Cmp(big.NewInt(0)), -1, "amount1 should be >= 0")

	// // Verify amounts don't exceed the max budget
	// assert.LessOrEqual(t, amount0.Cmp(amount0Max), 0, "amount0 should not exceed amount0Max")
	// assert.LessOrEqual(t, amount1.Cmp(amount1Max), 0, "amount1 should not exceed amount1Max")
}

func TestCalculateTokenAmountsFromLiquidity(t *testing.T) {

	liquidity := big.NewInt(845179049218237)
	sqrtPriceX96, _ := big.NewInt(0).SetString("275467826341246019486853", 10)
	tickLower := -252000
	tickUpper := -240800
	amount0, amount1, err := CalculateTokenAmountsFromLiquidity(liquidity, sqrtPriceX96, int32(tickLower), int32(tickUpper))
	if err != nil {
		t.Error(err)
	}
	t.Log("amount0:", amount0)
	t.Log("amount1:", amount1)
}
