package blackholedex

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ChoSanghyuk/blackholedex/pkg/types"

	"github.com/ethereum/go-ethereum/common"
)

// Swap performs a token-to-token swap on Blackhole DEX
// It first approves the swap router to spend the input token, then executes the swap
func (b *Blackhole) Swap(
	params *types.SWAPExactTokensForTokensParams,
) (common.Hash, error) { // todo. 다른 함수들처럼 result 반환으로 수정 필요?
	if len(params.Routes) == 0 {
		return common.Hash{}, errors.New("no routes provided")
	}

	swapClient, err := b.registry.Client(routerv2)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get swap client %s: %w", routerv2, err)
	}

	fromTokenAddress := params.Routes[0].From.Hex()
	tokenClient, err := b.registry.ClientByAddress(fromTokenAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get from client for token %s: %w", fromTokenAddress, err)
	}

	// Get the ERC20 client for the input token (first token in the route)
	// Step 1: Approve the swap router to spend the input tokens

	approveTxHash, err := b.ensureApproval(tokenClient, *swapClient.ContractAddress(), params.AmountIn)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to approve tokens: %w", err)
	}

	if approveTxHash != (common.Hash{}) {
		// Log approval transaction hash (in production, you might want to wait for confirmation)
		_, err = b.tl.WaitForTransaction(approveTxHash)
		if err != nil {
			return common.Hash{}, fmt.Errorf("failed to approve tokens: %w", err)
		}
	}

	// Step 2: Execute the swap
	swapTxHash, err := swapClient.Send(
		types.Standard,
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

// ensureApproval ensures token approval exists, optimizing to reuse existing allowances
// Returns transaction hash (zero if approval not needed), or error
func (b *Blackhole) ensureApproval(
	tokenClient ContractClient,
	spender common.Address,
	requiredAmount *big.Int,
) (common.Hash, error) {
	// Check existing allowance
	result, err := tokenClient.Call(&b.myAddr, "allowance", b.myAddr, spender)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to check allowance: %w", err)
	}

	currentAllowance := result[0].(*big.Int)

	// Only approve if insufficient
	if currentAllowance.Cmp(requiredAmount) >= 0 {
		// Sufficient allowance already exists
		return common.Hash{}, nil
	}

	// Approve required amount
	txHash, err := tokenClient.Send(
		types.Standard,
		&b.myAddr,
		b.privateKey,
		"approve",
		spender,
		requiredAmount,
	)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to approve tokens: %w", err)
	}

	return txHash, nil
}
