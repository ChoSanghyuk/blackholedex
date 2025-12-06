package blackholedex

import (
	"blackholego/internal/util"
	"blackholego/pkg/types"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const (
	// # Contract addresses
	routerv2       = "0x04E1dee021Cd12bBa022A72806441B43d8212Fec"
	usdc           = "0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E"
	wavax          = "0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7"
	black          = "0xcd94a87696fac69edae3a70fe5725307ae1c43f6"
	wavaxBlackPair = "0x14e4a5bed2e5e688ee1a5ca3a4914250d1abd573" //
	wavaxUsdcPair  = "0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0"
	deployer       = "0x5d433a94a4a2aa8f9aa34d8d15692dc2e9960584"
)

// Blackhole manages interactions with Blackhole DEX contracts
type Blackhole struct {
	privateKey *ecdsa.PrivateKey
	myAddr     common.Address
	tl         TxListener
	ccm        map[string]ContractClient // ContractClientMap
}

func (b Blackhole) Client(address string) (ContractClient, error) {

	c := b.ccm[address]
	if c == nil {
		return nil, errors.New("no mapped client")
	}
	return c, nil
}

// Swap performs a token-to-token swap on Blackhole DEX
// It first approves the swap router to spend the input token, then executes the swap
func (b *Blackhole) Swap(
	params *SWAPExactTokensForTokensParams,
) (common.Hash, error) {
	if len(params.Routes) == 0 {
		return common.Hash{}, errors.New("no routes provided")
	}

	swapClient, err := b.Client(routerv2)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get swap client %s: %w", routerv2, err)
	}

	fromTokenAddress := params.Routes[0].From.Hex()
	tokenClient, err := b.Client(fromTokenAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get from client for token %s: %w", fromTokenAddress, err)
	}

	// Get the ERC20 client for the input token (first token in the route)
	// Step 1: Approve the swap router to spend the input tokens

	approveTxHash, err := tokenClient.Send(
		types.Standard,
		nil, // Use automatic gas limit estimation
		&b.myAddr,
		b.privateKey,
		"approve",
		swapClient.ContractAddress(),
		params.AmountIn,
	)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to approve tokens: %w", err)
	}

	// Log approval transaction hash (in production, you might want to wait for confirmation)
	_, err = b.tl.WaitForTransaction(approveTxHash)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to approve tokens: %w", err)
	}

	// Step 2: Execute the swap
	swapTxHash, err := swapClient.Send(
		types.Standard,
		nil, // Use automatic gas limit estimation
		&b.myAddr,
		b.privateKey,
		"swapExactTokensForTokens",
		params.AmountIn,
		params.AmountOutMin,
		params.Routes,
		params.To,
		params.Deadline,
	)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to execute swap: %w", err)
	}

	return swapTxHash, nil
}

// GetAMMState retrieves the current state of an AMM pool
// This is a read-only operation that does not create a transaction
func (b *Blackhole) GetAMMState(poolAddress common.Address) (*AMMState, error) {
	poolClient, err := b.Client(poolAddress.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get pool client for %s: %w", poolAddress.Hex(), err)
	}

	// Call safelyGetStateOfAMM - this is a read-only operation
	result, err := poolClient.Call(nil, "safelyGetStateOfAMM")
	if err != nil {
		return nil, fmt.Errorf("failed to call safelyGetStateOfAMM: %w", err)
	}

	// Validate result length
	if len(result) != 7 {
		return nil, fmt.Errorf("unexpected result length: expected 7, got %d", len(result))
	}

	// Parse results into AMMState struct
	// The order matches the ABI outputs: sqrtPrice, tick, lastFee, pluginConfig, activeLiquidity, nextTick, previousTick
	state := &AMMState{
		SqrtPrice:       result[0].(*big.Int),
		Tick:            int32(result[1].(*big.Int).Int64()),
		LastFee:         result[2].(uint16),
		PluginConfig:    result[3].(uint8),
		ActiveLiquidity: result[4].(*big.Int),
		NextTick:        int32(result[5].(*big.Int).Int64()),
		PreviousTick:    int32(result[6].(*big.Int).Int64()),
	}

	return state, nil
}

func (b *Blackhole) Mint(token0 *common.Address, token1 *common.Address, maxP *big.Int, slp *big.Int) (*common.Hash, error) {
	clg := 200 // CL Gap
	lpg := 3   // Liquidity Providing Gap

	state, err := b.GetAMMState(common.HexToAddress(wavaxUsdcPair))
	if err != nil {
		return nil, err
	}

	tickLower := (int(state.Tick)/clg - lpg) * 200
	tickUpper := (int(state.Tick)/clg + lpg) * 200

	wavaxClient, err := b.Client(wavax)
	outputs, err := wavaxClient.Call(&b.myAddr, "balanceOf")
	wavaxBalace := outputs[0].(*big.Int)
	// wavaxMax := wavaxBalace.Sub()

	usdcClient, err := b.Client(usdc)
	outputs, err = usdcClient.Call(&b.myAddr, "balanceOf")
	usdcBalace := outputs[0].(*big.Int)

	amount0Desired, amount1Desired, _ := util.ComputeAmounts(state.SqrtPrice, int(state.Tick), tickLower, tickUpper, wavaxBalace, usdcBalace)

	deadline := big.NewInt(time.Now().Add(20 * time.Minute).Unix())
	params := &MintParams{
		Token0:         common.HexToAddress(wavax),
		Token1:         common.HexToAddress(usdc),
		Deployer:       common.HexToAddress(deployer),
		TickLower:      big.NewInt(int64(tickLower)),
		TickUpper:      big.NewInt(int64(tickUpper)),
		Amount0Desired: amount0Desired,
		Amount1Desired: amount1Desired,
		Amount0Min:     amount0Desired.Mul(amount0Desired, big.NewInt(100-slp.Int64())).Div(amount0Desired, big.NewInt(100)),
		Amount1Min:     amount1Desired.Mul(amount1Desired, big.NewInt(100-slp.Int64())).Div(amount0Desired, big.NewInt(100)),
		Recipient:      b.myAddr,
		Deadline:       deadline,
	}
	_ = params

	return nil, nil
	// t.Logf("MintParams : %v\n", params)
}
