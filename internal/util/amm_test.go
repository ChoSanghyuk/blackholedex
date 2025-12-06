package util

import (
	"math/big"
	"testing"

	"github.com/test-go/testify/assert"
)

// memo. SqrtPrice는 두 tick 사이의 값이기에, safelyGetStateOfAMM 결과로 나오는 sprtPrice 랑 tick 완벽하게 매칭되지 않음
func TestTickToSqrtPriceX96(t *testing.T) {

	sqrtPrice := TickToSqrtPriceX96(-249428)

	expected, _ := big.NewInt(0).SetString("304011615425126403287043", 10)
	assert.Equal(t, expected, sqrtPrice)
}

func TestComputeAmounts(t *testing.T) {

	sqrtPrice := TickToSqrtPriceX96(-249428)

	expected, _ := big.NewInt(0).SetString("304011615425126403287043", 10)
	assert.Equal(t, expected, sqrtPrice)
}
