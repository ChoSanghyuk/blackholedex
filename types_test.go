package blackholedex

import (
	"math/big"
	"testing"

	"blackholego/internal/util"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestPacking(t *testing.T) {
	t.Run("SWAPExactETHForTokensParams", func(t *testing.T) {

		// txHash 0x1600e68bfd607a5e8452f7533b162eeb4afd4f0435f31639999aa46fbaef79b1의 txData 값.
		swapExactTokensForTokensTxData := "6ba16543000000000000000000000000000000000000000000000038b4034b62cec2f5a10000000000000000000000000000000000000000000000000000000000000080000000000000000000000000b4dd4fb3d4bced984cce972991fb100488b59223000000000000000000000000000000000000000000000000000000006927fa81000000000000000000000000000000000000000000000000000000000000000100000000000000000000000014e4a5bed2e5e688ee1a5ca3a4914250d1abd573000000000000000000000000b31f66aa3c1e785363f0875a1b74e27b85fd66c7000000000000000000000000cd94a87696fac69edae3a70fe5725307ae1c43f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000b4dd4fb3d4bced984cce972991fb100488b59223"

		// 로컬 Parameter로 동일한 데이터 packing
		amountOutMin, _ := big.NewInt(0).SetString("1045988962367239812513", 10)
		params := SWAPExactETHForTokensParams{
			AmountOutMin: amountOutMin,
			Routes: []Route{
				{
					Pair:         common.HexToAddress("0x14e4a5bed2e5e688ee1a5ca3a4914250d1abd573"),
					From:         common.HexToAddress("0xb31f66aa3c1e785363f0875a1b74e27b85fd66c7"),
					To:           common.HexToAddress("0xcd94a87696fac69edae3a70fe5725307ae1c43f6"),
					Stable:       false,
					Concentrated: false,
					Receiver:     common.HexToAddress("0xb4dd4fb3d4bced984cce972991fb100488b59223"),
				},
			},
			To:       common.HexToAddress("0xb4dd4fb3D4bCED984cce972991fB100488b59223"),
			Deadline: big.NewInt(1764227713),
		}

		routerABI, err := util.LoadABI("blackholedex-contracts/artifacts/contracts/RouterV2.sol/RouterV2.json")
		if err != nil {
			t.Skipf("Could not load RouterV2 artifact: %v", err)
		}
		packed, err := routerABI.Pack("swapExactETHForTokens", params.AmountOutMin, params.Routes, params.To, params.Deadline)
		if err != nil {
			t.Fatalf("Failed to pack: %v", err)
		}

		// t.Logf("Packed data length: %d bytes", len(packed))
		// t.Logf("Method selector: 0x%x", packed[:4])
		t.Logf("Packed data: 0x%x", packed)

		assert.Equal(t, common.Bytes2Hex(packed), swapExactTokensForTokensTxData)
	})

	t.Run("SWAPExactTokensForTokensParams", func(t *testing.T) {

		// txHash 0x1600e68bfd607a5e8452f7533b162eeb4afd4f0435f31639999aa46fbaef79b1의 txData 값.
		swapExactTokensForTokensTxData := ""

		// 로컬 Parameter로 동일한 데이터 packing
		amountIn, _ := big.NewInt(0).SetString("", 10)
		amountOutMin, _ := big.NewInt(0).SetString("", 10)
		params := SWAPExactTokensForTokensParams{
			AmountIn:     amountIn,
			AmountOutMin: amountOutMin,
			Routes: []Route{
				{
					Pair:         common.HexToAddress(""),
					From:         common.HexToAddress(""),
					To:           common.HexToAddress(""),
					Stable:       false,
					Concentrated: false,
					Receiver:     common.HexToAddress(""),
				},
			},
			To:       common.HexToAddress(""),
			Deadline: big.NewInt(1764227713),
		}

		routerABI, err := util.LoadABI("blackholedex-contracts/artifacts/contracts/RouterV2.sol/RouterV2.json")
		if err != nil {
			t.Skipf("Could not load RouterV2 artifact: %v", err)
		}
		packed, err := routerABI.Pack("swapExactTokensForTokens", params.AmountIn, params.AmountOutMin, params.Routes, params.To, params.Deadline)
		if err != nil {
			t.Fatalf("Failed to pack: %v", err)
		}

		// t.Logf("Packed data length: %d bytes", len(packed))
		// t.Logf("Method selector: 0x%x", packed[:4])
		t.Logf("Packed data: 0x%x", packed)

		assert.Equal(t, common.Bytes2Hex(packed), swapExactTokensForTokensTxData)
	})

	t.Run("MintParams", func(t *testing.T) {

		// txHash 0x9e2247a0210448cab301475eef741eba0ee9a9351188a92b8127fce27206b9d0의 txData 값.
		// txData (0x 없이)
		swapExactTokensForTokensTxData := "fe3f3be7000000000000000000000000b31f66aa3c1e785363f0875a1b74e27b85fd66c7000000000000000000000000b97ef9ef8734c71904d8002f8b6bc66dd9c48a6e0000000000000000000000005d433a94a4a2aa8f9aa34d8d15692dc2e9960584fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc3100fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc35b0000000000000000000000000000000000000000000000000340d7f1b384fc2cb0000000000000000000000000000000000000000000000000000000003a8a540000000000000000000000000000000000000000000000000317338c0424bc5da000000000000000000000000000000000000000000000000000000000379d030000000000000000000000000b4dd4fb3d4bced984cce972991fb100488b592230000000000000000000000000000000000000000000000000000019a9267bb33"

		// 로컬 Parameter로 동일한 데이터 packing
		amount0Desired, _ := big.NewInt(0).SetString("3750793819555087051", 10)
		amount1Desired := big.NewInt(61384000)
		amount0Min, _ := big.NewInt(0).SetString("3563254128577332698", 10)
		amount1Min := big.NewInt(58314800)
		deadline, _ := big.NewInt(0).SetString("1763392863027", 10)

		params := MintParams{
			Token0:         common.HexToAddress("0xb31f66aa3c1e785363f0875a1b74e27b85fd66c7"),
			Token1:         common.HexToAddress("0xb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e"),
			Deployer:       common.HexToAddress("0x5d433a94a4a2aa8f9aa34d8d15692dc2e9960584"),
			TickLower:      big.NewInt(-249600),
			TickUpper:      big.NewInt(-248400),
			Amount0Desired: amount0Desired,
			Amount1Desired: amount1Desired,
			Amount0Min:     amount0Min,
			Amount1Min:     amount1Min,
			Recipient:      common.HexToAddress("0xb4dd4fb3d4bced984cce972991fb100488b59223"),
			Deadline:       deadline,
		}

		positionManagerABI, err := util.LoadABI("blackholedex-contracts/artifacts/@cryptoalgebra/integral-periphery/contracts/interfaces/INonfungiblePositionManager.sol/INonfungiblePositionManager.json")
		if err != nil {
			t.Skipf("Could not load INonfungiblePositionManager artifact: %v", err)
		}
		packed, err := positionManagerABI.Pack("mint", params)
		if err != nil {
			t.Fatalf("Failed to pack: %v", err)
		}

		// t.Logf("Packed data length: %d bytes", len(packed))
		// t.Logf("Method selector: 0x%x", packed[:4])
		t.Logf("Packed data: 0x%x", packed)

		assert.Equal(t, common.Bytes2Hex(packed), swapExactTokensForTokensTxData)
	})

	t.Run("ERC20 Approve", func(t *testing.T) {

		approveTxData := "095ea7b30000000000000000000000003fed017ec0f5517cdf2e8a9a4156c64d74252146000000000000000000000000000000000000000000000000340d7f1b384fc2cb"
		// Sample data for approve function
		amount, _ := big.NewInt(0).SetString("3750793819555087051", 10)

		params := ApproveParams{
			Spender: common.HexToAddress("0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146"),
			Amount:  amount,
		}

		erc20ABI, err := util.LoadABI("blackholedex-contracts/artifacts/@openzeppelin/contracts/token/ERC20/ERC20.sol/ERC20.json")
		if err != nil {
			t.Skipf("Could not load ERC20 artifact: %v", err)
		}

		packed, err := erc20ABI.Pack("approve", params.Spender, params.Amount)
		if err != nil {
			t.Fatalf("Failed to pack: %v", err)
		}

		assert.Equal(t, common.Bytes2Hex(packed), approveTxData)
	})
}

func TestAddLiquidityParams(t *testing.T) {
	params := AddLiquidityParams{
		TokenA:         common.HexToAddress("0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E"),
		TokenB:         common.HexToAddress("0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7"),
		Stable:         false,
		AmountADesired: big.NewInt(1000000000000000000),
		AmountBDesired: big.NewInt(1000000000000000000),
		AmountAMin:     big.NewInt(990000000000000000),
		AmountBMin:     big.NewInt(990000000000000000),
		To:             common.HexToAddress("0x3333333333333333333333333333333333333333"),
		Deadline:       big.NewInt(1700000000),
	}

	if params.TokenA == params.TokenB {
		t.Error("TokenA and TokenB should be different")
	}

	t.Logf("AddLiquidityParams validated successfully")
}

func TestVoteParams(t *testing.T) {
	params := VoteParams{
		TokenID: big.NewInt(123),
		Pools: []common.Address{
			common.HexToAddress("0x1111111111111111111111111111111111111111"),
			common.HexToAddress("0x2222222222222222222222222222222222222222"),
		},
		Weights: []*big.Int{
			big.NewInt(5000),
			big.NewInt(5000),
		},
	}

	if len(params.Pools) != len(params.Weights) {
		t.Error("Pools and Weights should have same length")
	}

	totalWeight := new(big.Int).Add(params.Weights[0], params.Weights[1])
	if totalWeight.Cmp(big.NewInt(10000)) != 0 {
		t.Logf("Total weight: %s (should typically be 10000 for 100%%)", totalWeight.String())
	}

	t.Logf("VoteParams validated successfully")
}
