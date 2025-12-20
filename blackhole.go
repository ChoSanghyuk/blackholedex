package blackholedex

import (
	"blackholego/internal/util"
	"blackholego/pkg/types"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const (
	// # Contract addresses
	routerv2                   = "0x04E1dee021Cd12bBa022A72806441B43d8212Fec"
	usdc                       = "0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E"
	wavax                      = "0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7"
	black                      = "0xcd94a87696fac69edae3a70fe5725307ae1c43f6"
	wavaxBlackPair             = "0x14e4a5bed2e5e688ee1a5ca3a4914250d1abd573"
	wavaxUsdcPair              = "0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0"
	deployer                   = "0x5d433a94a4a2aa8f9aa34d8d15692dc2e9960584"
	nonfungiblePositionManager = "0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146"
	gauge                      = "0x3ADE52f9779c07471F4B6d5997444C3c2124C1c0"
	farmingCenter              = "0xa47Ad2C95FaE476a73b85A355A5855aDb4b3A449"
	algebraPool                = "0x41100c6d2c6920b10d12cd8d59c8a9aa2ef56fc7"
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
		return nil, errors.New("no mapped client") // todo. 없으면 생성.
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

// validateBalances validates wallet has sufficient token balances
// Returns error if insufficient balance, nil otherwise
func (b *Blackhole) validateBalances(requiredWAVAX, requiredUSDC *big.Int) error {
	wavaxClient, err := b.Client(wavax)
	if err != nil {
		return fmt.Errorf("failed to get WAVAX client: %w", err)
	}

	usdcClient, err := b.Client(usdc)
	if err != nil {
		return fmt.Errorf("failed to get USDC client: %w", err)
	}

	// Query WAVAX balance
	wavaxResult, err := wavaxClient.Call(&b.myAddr, "balanceOf", b.myAddr)
	if err != nil {
		return fmt.Errorf("failed to get WAVAX balance: %w", err)
	}
	wavaxBalance := wavaxResult[0].(*big.Int)

	// Query USDC balance
	usdcResult, err := usdcClient.Call(&b.myAddr, "balanceOf", b.myAddr)
	if err != nil {
		return fmt.Errorf("failed to get USDC balance: %w", err)
	}
	usdcBalance := usdcResult[0].(*big.Int)

	// Validate WAVAX balance
	if wavaxBalance.Cmp(requiredWAVAX) < 0 {
		return fmt.Errorf("insufficient WAVAX balance: have %s, need %s",
			wavaxBalance.String(), requiredWAVAX.String())
	}

	// Validate USDC balance
	if usdcBalance.Cmp(requiredUSDC) < 0 {
		return fmt.Errorf("insufficient USDC balance: have %s, need %s",
			usdcBalance.String(), requiredUSDC.String())
	}

	return nil
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
		nil, // Use automatic gas limit estimation
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

// Mint stakes liquidity in WAVAX-USDC pool with automatic position calculation
// maxWAVAX: Maximum WAVAX amount to stake (wei)
// maxUSDC: Maximum USDC amount to stake (smallest unit)
// rangeWidth: Position range width (e.g., 6 = ±3 tick ranges)
// slippagePct: Slippage tolerance percentage (e.g., 5 = 5%)
// Returns StakingResult with all transaction details and position info
func (b *Blackhole) Mint(
	maxWAVAX *big.Int,
	maxUSDC *big.Int,
	rangeWidth int,
	slippagePct int,
) (*StakingResult, error) {
	const tickSpacing = 200

	// T012: Input validation
	if err := util.ValidateStakingRequest(maxWAVAX, maxUSDC, rangeWidth, slippagePct); err != nil {
		return &StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("validation failed: %v", err),
		}, err
	}

	// Initialize transaction tracking
	var transactions []TransactionRecord

	// T013: Query pool state
	state, err := b.GetAMMState(common.HexToAddress(wavaxUsdcPair))
	if err != nil {
		return &StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to query pool state: %v", err),
		}, fmt.Errorf("failed to query pool state: %w", err)
	}

	// T014: Calculate tick bounds
	tickLower, tickUpper, err := util.CalculateTickBounds(state.Tick, rangeWidth, tickSpacing)
	if err != nil {
		return &StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to calculate tick bounds: %v", err),
		}, fmt.Errorf("failed to calculate tick bounds: %w", err)
	}

	// T015: Calculate optimal amounts using existing ComputeAmounts utility
	amount0Desired, amount1Desired, _ := util.ComputeAmounts(
		state.SqrtPrice,
		int(state.Tick),
		int(tickLower),
		int(tickUpper),
		maxWAVAX,
		maxUSDC,
	)

	// T033: Compare actual vs desired amounts for capital efficiency
	// T034: Calculate and log capital utilization percentages
	utilization0 := new(big.Int).Mul(amount0Desired, big.NewInt(100)) // (amount0Desired / maxWAVAX) * 100. 최대 가능 금액 대비 staking되는 금액의 비율
	utilization0.Div(utilization0, maxWAVAX)
	utilization1 := new(big.Int).Mul(amount1Desired, big.NewInt(100))
	utilization1.Div(utilization1, maxUSDC)

	log.Printf("Capital Utilization: WAVAX %d%%, USDC %d%%",
		utilization0.Int64(), utilization1.Int64())

	// T032: Warn if >10% of either token will be unused (capital efficiency warning)
	wastedWAVAX := new(big.Int).Sub(maxWAVAX, amount0Desired)
	wastedUSDC := new(big.Int).Sub(maxUSDC, amount1Desired)

	if utilization0.Cmp(big.NewInt(90)) < 0 { // Less than 90% utilized = >10% wasted
		wastePercent := new(big.Int).Mul(wastedWAVAX, big.NewInt(100))
		wastePercent.Div(wastePercent, maxWAVAX)
		log.Printf("⚠️  Capital Efficiency Warning: %d%% of WAVAX (%s wei) will not be staked. Consider adjusting amounts or range width.",
			wastePercent.Int64(), wastedWAVAX.String())
	}
	if utilization1.Cmp(big.NewInt(90)) < 0 { // Less than 90% utilized = >10% wasted
		wastePercent := new(big.Int).Mul(wastedUSDC, big.NewInt(100))
		wastePercent.Div(wastePercent, maxUSDC)
		log.Printf("⚠️  Capital Efficiency Warning: %d%% of USDC (%s smallest unit) will not be staked. Consider adjusting amounts or range width.",
			wastePercent.Int64(), wastedUSDC.String())
	}

	// T016: Validate balances
	if err := b.validateBalances(amount0Desired, amount1Desired); err != nil {
		return &StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("balance validation failed: %v", err),
		}, fmt.Errorf("balance validation failed: %w", err)
	}

	// T017: Calculate slippage protection
	amount0Min := util.CalculateMinAmount(amount0Desired, slippagePct)
	amount1Min := util.CalculateMinAmount(amount1Desired, slippagePct)

	// Get contract clients
	wavaxClient, err := b.Client(wavax)
	if err != nil {
		return &StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get WAVAX client: %v", err),
		}, fmt.Errorf("failed to get WAVAX client: %w", err)
	}

	usdcClient, err := b.Client(usdc)
	if err != nil {
		return &StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get USDC client: %v", err),
		}, fmt.Errorf("failed to get USDC client: %w", err)
	}

	nftManagerAddr := common.HexToAddress(nonfungiblePositionManager)

	// T018: WAVAX approval
	wavaxApproveTxHash, err := b.ensureApproval(wavaxClient, nftManagerAddr, amount0Desired)
	if err != nil {
		return &StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to approve WAVAX: %v", err),
		}, fmt.Errorf("failed to approve WAVAX: %w", err)
	}

	// Wait for WAVAX approval if transaction was sent
	if wavaxApproveTxHash != (common.Hash{}) {
		receipt, err := b.tl.WaitForTransaction(wavaxApproveTxHash)
		if err != nil {
			return &StakingResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("WAVAX approval transaction failed: %v", err),
			}, fmt.Errorf("WAVAX approval transaction failed: %w", err)
		}

		// T024: Extract gas cost
		gasCost, err := util.ExtractGasCost(receipt)
		if err != nil {
			return &StakingResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("failed to extract gas cost: %v", err),
			}, fmt.Errorf("failed to extract gas cost: %w", err)
		}

		// Parse gas price for record
		gasPrice := new(big.Int)
		gasPrice.SetString(receipt.EffectiveGasPrice, 0)

		// Parse gas used
		gasUsed := new(big.Int)
		gasUsed.SetString(receipt.GasUsed, 0)

		transactions = append(transactions, TransactionRecord{
			TxHash:    wavaxApproveTxHash,
			GasUsed:   gasUsed.Uint64(),
			GasPrice:  gasPrice,
			GasCost:   gasCost,
			Timestamp: time.Now(),
			Operation: "ApproveWAVAX",
		})
	}

	// T019: USDC approval
	usdcApproveTxHash, err := b.ensureApproval(usdcClient, nftManagerAddr, amount1Desired)
	if err != nil {
		return &StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to approve USDC: %v", err),
		}, fmt.Errorf("failed to approve USDC: %w", err)
	}

	// Wait for USDC approval if transaction was sent
	if usdcApproveTxHash != (common.Hash{}) {
		receipt, err := b.tl.WaitForTransaction(usdcApproveTxHash)
		if err != nil {
			return &StakingResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("USDC approval transaction failed: %v", err),
			}, fmt.Errorf("USDC approval transaction failed: %w", err)
		}

		// Extract gas cost
		gasCost, err := util.ExtractGasCost(receipt)
		if err != nil {
			return &StakingResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("failed to extract gas cost: %v", err),
			}, fmt.Errorf("failed to extract gas cost: %w", err)
		}

		// Parse gas price for record
		gasPrice := new(big.Int)
		gasPrice.SetString(receipt.EffectiveGasPrice, 0)

		// Parse gas used
		gasUsed := new(big.Int)
		gasUsed.SetString(receipt.GasUsed, 0)

		transactions = append(transactions, TransactionRecord{
			TxHash:    usdcApproveTxHash,
			GasUsed:   gasUsed.Uint64(),
			GasPrice:  gasPrice,
			GasCost:   gasCost,
			Timestamp: time.Now(),
			Operation: "ApproveUSDC",
		})
	}

	// T020: Construct MintParams
	deadline := big.NewInt(time.Now().Add(20 * time.Minute).Unix())
	mintParams := &MintParams{
		Token0:         common.HexToAddress(wavax),
		Token1:         common.HexToAddress(usdc),
		Deployer:       common.HexToAddress(deployer),
		TickLower:      big.NewInt(int64(tickLower)),
		TickUpper:      big.NewInt(int64(tickUpper)),
		Amount0Desired: amount0Desired,
		Amount1Desired: amount1Desired,
		Amount0Min:     amount0Min,
		Amount1Min:     amount1Min,
		Recipient:      b.myAddr,
		Deadline:       deadline,
	}

	// T021: Get NonfungiblePositionManager client
	nftManagerClient, err := b.Client(nonfungiblePositionManager)
	if err != nil {
		return &StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get NFT manager client: %v", err),
		}, fmt.Errorf("failed to get NFT manager client: %w", err)
	}

	// T022: Submit mint transaction
	mintTxHash, err := nftManagerClient.Send(
		types.Standard,
		nil, // Use automatic gas limit estimation
		&b.myAddr,
		b.privateKey,
		"mint",
		mintParams,
	)
	if err != nil {
		return &StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to submit mint transaction: %v", err),
		}, fmt.Errorf("failed to submit mint transaction: %w", err)
	}

	// T023: Wait for mint confirmation
	mintReceipt, err := b.tl.WaitForTransaction(mintTxHash)
	if err != nil {
		return &StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("mint transaction failed: %v", err),
		}, fmt.Errorf("mint transaction failed: %w", err)
	}

	// Extract gas cost for mint
	mintGasCost, err := util.ExtractGasCost(mintReceipt)
	if err != nil {
		return &StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to extract mint gas cost: %v", err),
		}, fmt.Errorf("failed to extract mint gas cost: %w", err)
	}

	// Parse gas price for record
	mintGasPrice := new(big.Int)
	mintGasPrice.SetString(mintReceipt.EffectiveGasPrice, 0)

	// Parse gas used
	mintGasUsed := new(big.Int)
	mintGasUsed.SetString(mintReceipt.GasUsed, 0)

	transactions = append(transactions, TransactionRecord{
		TxHash:    mintTxHash,
		GasUsed:   mintGasUsed.Uint64(),
		GasPrice:  mintGasPrice,
		GasCost:   mintGasCost,
		Timestamp: time.Now(),
		Operation: "Mint",
	})

	// T025: Parse NFT token ID from Transfer event in receipt
	// The Transfer event is emitted when the NFT is minted (from 0x0 to recipient)
	// Event signature: Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
	nftTokenID := MintNftTokenId(nftManagerClient, mintReceipt)

	// T026: Construct StakingResult
	totalGasCost := big.NewInt(0)
	for _, tx := range transactions {
		totalGasCost.Add(totalGasCost, tx.GasCost)
	}

	result := &StakingResult{
		NFTTokenID:     nftTokenID,
		ActualAmount0:  amount0Desired, // Actual amounts would be in mint receipt
		ActualAmount1:  amount1Desired,
		FinalTickLower: tickLower,
		FinalTickUpper: tickUpper,
		Transactions:   transactions,
		TotalGasCost:   totalGasCost,
		Success:        true,
		ErrorMessage:   "",
	}

	// T028: Transaction logging
	fmt.Printf("✓ Liquidity staked successfully\n")
	fmt.Printf("  Position: Tick %d to %d\n", tickLower, tickUpper)
	fmt.Printf("  WAVAX: %s wei\n", amount0Desired.String())
	fmt.Printf("  USDC: %s\n", amount1Desired.String())
	fmt.Printf("  Total Gas Cost: %s wei\n", totalGasCost.String())
	for _, tx := range transactions {
		fmt.Printf("  - %s: %s (gas: %s wei)\n", tx.Operation, tx.TxHash.Hex(), tx.GasCost.String())
	}

	return result, nil
}

// Stake stakes a liquidity position NFT in a GaugeV2 contract to earn additional rewards
// nftTokenID: ERC721 token ID from previous Mint operation
// gaugeAddress: GaugeV2 contract address (must match pool)
// Returns StakingResult with transaction tracking and gas costs
func (b *Blackhole) Stake(
	nftTokenID *big.Int,
) (*StakingResult, error) {
	// T007-T008: Input validation
	if nftTokenID == nil || nftTokenID.Sign() <= 0 {
		return &StakingResult{
			Success:      false,
			ErrorMessage: "validation failed: invalid token ID",
		}, fmt.Errorf("validation failed: invalid token ID")
	}

	// T009: Initialize transaction tracking
	var transactions []TransactionRecord

	// T011-T014: NFT Ownership Verification
	nftManagerClient, err := b.Client(nonfungiblePositionManager)
	if err != nil {
		return &StakingResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get NFT manager client: %v", err),
		}, fmt.Errorf("failed to get NFT manager client: %w", err)
	}

	// Query NFT ownership
	ownerResult, err := nftManagerClient.Call(&b.myAddr, "ownerOf", nftTokenID)
	if err != nil {
		return &StakingResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to verify NFT ownership: %v", err),
		}, fmt.Errorf("failed to verify NFT ownership: %w", err)
	}

	owner := ownerResult[0].(common.Address)
	if owner != b.myAddr {
		return &StakingResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("NFT not owned by wallet: owned by %s", owner.Hex()),
		}, fmt.Errorf("NFT not owned by wallet")
	}

	// T015-T023: NFT Approval Check and Execution
	approvalResult, err := nftManagerClient.Call(&b.myAddr, "getApproved", nftTokenID)
	if err != nil {
		return &StakingResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to check NFT approval: %v", err),
		}, fmt.Errorf("failed to check NFT approval: %w", err)
	}

	currentApproval := approvalResult[0].(common.Address)

	// Only approve if not already approved for this gauge
	if currentApproval.Hex() != gauge {
		log.Printf("Approving NFT %s for gauge %s", nftTokenID.String(), gauge)

		approveTxHash, err := nftManagerClient.Send(
			types.Standard,
			nil, // Use automatic gas limit estimation
			&b.myAddr,
			b.privateKey,
			"approve",
			common.HexToAddress(gauge),
			nftTokenID,
		)
		if err != nil {
			return &StakingResult{
				NFTTokenID:   nftTokenID,
				Success:      false,
				ErrorMessage: fmt.Sprintf("failed to approve NFT: %v", err),
			}, fmt.Errorf("failed to approve NFT: %w", err)
		}

		// Wait for approval confirmation
		approvalReceipt, err := b.tl.WaitForTransaction(approveTxHash)
		if err != nil {
			return &StakingResult{
				NFTTokenID:   nftTokenID,
				Success:      false,
				ErrorMessage: fmt.Sprintf("NFT approval transaction failed: %v", err),
			}, fmt.Errorf("NFT approval transaction failed: %w", err)
		}

		// Track approval transaction
		gasCost, err := util.ExtractGasCost(approvalReceipt)
		if err != nil {
			return &StakingResult{
				NFTTokenID:   nftTokenID,
				Success:      false,
				ErrorMessage: fmt.Sprintf("failed to extract approval gas cost: %v", err),
			}, fmt.Errorf("failed to extract approval gas cost: %w", err)
		}

		gasPrice := new(big.Int)
		gasPrice.SetString(approvalReceipt.EffectiveGasPrice, 0)
		gasUsed := new(big.Int)
		gasUsed.SetString(approvalReceipt.GasUsed, 0)

		transactions = append(transactions, TransactionRecord{
			TxHash:    approveTxHash,
			GasUsed:   gasUsed.Uint64(),
			GasPrice:  gasPrice,
			GasCost:   gasCost,
			Timestamp: time.Now(),
			Operation: "ApproveNFT",
		})
	} else {
		log.Printf("NFT already approved for gauge, skipping approval")
	}

	// T024-T030: Gauge Deposit Execution
	gaugeClient, err := b.Client(gauge)
	if err != nil {
		// Return with partial transaction records if approval was sent
		totalGasCost := big.NewInt(0)
		for _, tx := range transactions {
			totalGasCost.Add(totalGasCost, tx.GasCost)
		}
		return &StakingResult{
			NFTTokenID:   nftTokenID,
			Transactions: transactions,
			TotalGasCost: totalGasCost,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get gauge client: %v", err),
		}, fmt.Errorf("failed to get gauge client: %w", err)
	}

	// Submit deposit transaction
	log.Printf("Depositing NFT %s into gauge %s", nftTokenID.String(), gauge)

	depositTxHash, err := gaugeClient.Send(
		types.Standard,
		nil, // Use automatic gas limit estimation
		&b.myAddr,
		b.privateKey,
		"deposit",
		nftTokenID, // Token ID is the "amount" parameter
	)
	if err != nil {
		totalGasCost := big.NewInt(0)
		for _, tx := range transactions {
			totalGasCost.Add(totalGasCost, tx.GasCost)
		}
		return &StakingResult{
			NFTTokenID:   nftTokenID,
			Transactions: transactions,
			TotalGasCost: totalGasCost,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to submit deposit transaction: %v", err),
		}, fmt.Errorf("failed to submit deposit transaction: %w", err)
	}

	// Wait for deposit confirmation
	depositReceipt, err := b.tl.WaitForTransaction(depositTxHash)
	if err != nil {
		totalGasCost := big.NewInt(0)
		for _, tx := range transactions {
			totalGasCost.Add(totalGasCost, tx.GasCost)
		}
		return &StakingResult{
			NFTTokenID:   nftTokenID,
			Transactions: transactions,
			TotalGasCost: totalGasCost,
			Success:      false,
			ErrorMessage: fmt.Sprintf("deposit transaction failed: %v", err),
		}, fmt.Errorf("deposit transaction failed: %w", err)
	}

	// Track deposit transaction
	gasCost, err := util.ExtractGasCost(depositReceipt)
	if err != nil {
		totalGasCost := big.NewInt(0)
		for _, tx := range transactions {
			totalGasCost.Add(totalGasCost, tx.GasCost)
		}
		return &StakingResult{
			NFTTokenID:   nftTokenID,
			Transactions: transactions,
			TotalGasCost: totalGasCost,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to extract deposit gas cost: %v", err),
		}, fmt.Errorf("failed to extract deposit gas cost: %w", err)
	}

	gasPrice := new(big.Int)
	gasPrice.SetString(depositReceipt.EffectiveGasPrice, 0)
	gasUsed := new(big.Int)
	gasUsed.SetString(depositReceipt.GasUsed, 0)

	transactions = append(transactions, TransactionRecord{
		TxHash:    depositTxHash,
		GasUsed:   gasUsed.Uint64(),
		GasPrice:  gasPrice,
		GasCost:   gasCost,
		Timestamp: time.Now(),
		Operation: "DepositNFT",
	})

	// T031-T037: Result Construction and Gas Tracking
	totalGasCost := big.NewInt(0)
	for _, tx := range transactions {
		totalGasCost.Add(totalGasCost, tx.GasCost)
	}

	result := &StakingResult{
		NFTTokenID:     nftTokenID,
		ActualAmount0:  big.NewInt(0), // Not populated by Stake
		ActualAmount1:  big.NewInt(0), // Not populated by Stake
		FinalTickLower: 0,             // Not populated by Stake
		FinalTickUpper: 0,             // Not populated by Stake
		Transactions:   transactions,
		TotalGasCost:   totalGasCost,
		Success:        true,
		ErrorMessage:   "",
	}

	// T038-T043: Logging and User Feedback
	fmt.Printf("✓ NFT staked successfully\n")
	fmt.Printf("  Token ID: %s\n", nftTokenID.String())
	fmt.Printf("  Gauge: %s\n", gauge)
	fmt.Printf("  Total Gas Cost: %s wei\n", totalGasCost.String())
	for _, tx := range transactions {
		fmt.Printf("  - %s: %s (gas: %s wei)\n", tx.Operation, tx.TxHash.Hex(), tx.GasCost.String())
	}

	return result, nil
}

// Unstake withdraws a staked NFT position from FarmingCenter
// nftTokenID: ERC721 token ID from previous Mint operation
// incentiveKey: Identifies the farming program to exit
// collectRewards: Whether to claim accumulated rewards during unstake
// Returns UnstakeResult with transaction tracking and gas costs
func (b *Blackhole) Unstake(
	nftTokenID *big.Int,
	nonce *big.Int, // nonce가 왜 3으로 고정되는거 같지???. todo. 1. nonce 확인 2. reward value에 0 넣는거 확인
) (*UnstakeResult, error) {
	// T006: Input validation - NFT token ID
	if nftTokenID == nil || nftTokenID.Sign() <= 0 {
		return &UnstakeResult{
			Success:      false,
			ErrorMessage: "validation failed: invalid token ID",
		}, fmt.Errorf("validation failed: invalid token ID")
	}

	// Initialize transaction tracking
	var transactions []TransactionRecord

	// T008: Verify NFT ownership
	nftManagerClient, err := b.Client(nonfungiblePositionManager)
	if err != nil {
		return &UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get NFT manager client: %v", err),
		}, fmt.Errorf("failed to get NFT manager client: %w", err)
	}

	ownerResult, err := nftManagerClient.Call(&b.myAddr, "ownerOf", nftTokenID)
	if err != nil {
		return &UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to verify NFT ownership: %v", err),
		}, fmt.Errorf("failed to verify NFT ownership: %w", err)
	}

	owner := ownerResult[0].(common.Address)
	if owner != b.myAddr {
		return &UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("NFT not owned by wallet: owned by %s", owner.Hex()),
		}, fmt.Errorf("NFT not owned by wallet")
	}

	// T009: Verify NFT is currently farmed
	farmingCenterClient, err := b.Client(farmingCenter)
	if err != nil {
		return &UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get FarmingCenter client: %v", err),
		}, fmt.Errorf("failed to get FarmingCenter client: %w", err)
	}

	depositsResult, err := farmingCenterClient.Call(&b.myAddr, "deposits", nftTokenID)
	if err != nil {
		return &UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to check farming status: %v", err),
		}, fmt.Errorf("failed to check farming status: %w", err)
	}

	currentIncentiveId := depositsResult[0].([32]byte)
	if currentIncentiveId == [32]byte{} {
		return &UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: "NFT is not currently staked in farming",
		}, fmt.Errorf("NFT is not currently staked")
	}

	// T010: Build multicall data - encode exitFarming call
	var multicallData [][]byte

	incentiveKey := IncentiveKey{
		RewardToken:      common.HexToAddress(black),
		BonusRewardToken: common.HexToAddress(black),
		Pool:             common.HexToAddress(algebraPool),
		Nonce:            nonce,
	}

	farmingCenterABI := farmingCenterClient.Abi()
	exitFarmingData, err := farmingCenterABI.Pack("exitFarming", incentiveKey, nftTokenID)
	if err != nil {
		return &UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to encode exitFarming: %v", err),
		}, fmt.Errorf("failed to encode exitFarming: %w", err)
	}
	multicallData = append(multicallData, exitFarmingData)

	// T011: Conditionally encode collectRewards call
	collectRewardsData, err := farmingCenterABI.Pack("claimReward", common.HexToAddress(black), b.myAddr, big.NewInt(0)) // todo. reward 0원인거 확인.
	if err != nil {
		return &UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to encode collectRewards: %v", err),
		}, fmt.Errorf("failed to encode collectRewards: %w", err)
	}
	multicallData = append(multicallData, collectRewardsData)

	// T012: Execute multicall transaction
	log.Printf("Unstaking NFT %s from FarmingCenter %s", nftTokenID.String(), farmingCenter)

	multicallTxHash, err := farmingCenterClient.Send(
		types.Standard,
		nil, // Use automatic gas limit estimation (Principle 4)
		&b.myAddr,
		b.privateKey,
		"multicall",
		multicallData,
	)
	if err != nil {
		return &UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to submit multicall transaction: %v", err),
		}, fmt.Errorf("failed to submit multicall transaction: %w", err)
	}

	// T013: Wait for transaction confirmation and extract gas cost
	multicallReceipt, err := b.tl.WaitForTransaction(multicallTxHash)
	if err != nil {
		return &UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("multicall transaction failed: %v", err),
		}, fmt.Errorf("multicall transaction failed: %w", err)
	}

	gasCost, err := util.ExtractGasCost(multicallReceipt)
	if err != nil {
		return &UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to extract gas cost: %v", err),
		}, fmt.Errorf("failed to extract gas cost: %w", err)
	}

	gasPrice := new(big.Int)
	gasPrice.SetString(multicallReceipt.EffectiveGasPrice, 0)
	gasUsed := new(big.Int)
	gasUsed.SetString(multicallReceipt.GasUsed, 0)

	transactions = append(transactions, TransactionRecord{
		TxHash:    multicallTxHash,
		GasUsed:   gasUsed.Uint64(),
		GasPrice:  gasPrice,
		GasCost:   gasCost,
		Timestamp: time.Now(),
		Operation: "Unstake",
	})

	// T014: Parse reward amounts from multicall results (if collected)
	// Note: Reward parsing from multicall results would require decoding the return data
	// For now, we set rewards to default values - this should be enhanced with actual parsing
	rewards := &RewardAmounts{
		Reward:           big.NewInt(0),
		BonusReward:      big.NewInt(0),
		RewardToken:      incentiveKey.RewardToken,
		BonusRewardToken: incentiveKey.BonusRewardToken,
	}
	// TODO: Parse actual reward amounts from multicallReceipt logs or return data
	log.Printf("Rewards collected (parsing from receipt not yet implemented)")

	// T015: Construct and return UnstakeResult
	totalGasCost := big.NewInt(0)
	for _, tx := range transactions {
		totalGasCost.Add(totalGasCost, tx.GasCost)
	}

	result := &UnstakeResult{
		NFTTokenID:   nftTokenID,
		Rewards:      rewards,
		Transactions: transactions,
		TotalGasCost: totalGasCost,
		Success:      true,
		ErrorMessage: "",
	}

	// T016: Logging with troubleshooting context
	fmt.Printf("✓ NFT unstaked successfully\n")
	fmt.Printf("  Token ID: %s\n", nftTokenID.String())
	fmt.Printf("  FarmingCenter: %s\n", farmingCenter)
	if rewards != nil {
		fmt.Printf("  Rewards: %s / %s\n", rewards.Reward.String(), rewards.BonusReward.String())
	}
	fmt.Printf("  Total Gas Cost: %s wei\n", totalGasCost.String())
	for _, tx := range transactions {
		fmt.Printf("  - %s: %s (gas: %s wei)\n", tx.Operation, tx.TxHash.Hex(), tx.GasCost.String())
	}

	return result, nil
}

/**********************************************************************************************************************************************/

func MintNftTokenId(nftManagerClient ContractClient, mintReceipt *types.TxReceipt) *big.Int {
	nftTokenID := big.NewInt(0) // Default fallback
	// Parse receipt to extract events
	eventsJson, err := nftManagerClient.ParseReceipt(mintReceipt)
	if err != nil {
		log.Printf("Warning: Failed to parse mint receipt for token ID: %v", err)
	} else {
		// Parse the JSON to find Transfer event
		var events []map[string]interface{}
		if err := json.Unmarshal([]byte(eventsJson), &events); err == nil {
			for _, event := range events {
				if eventName, ok := event["EventName"].(string); ok && eventName == "Transfer" {
					if params, ok := event["Parameter"].(map[string]interface{}); ok {
						// Check if this is a mint (from zero address to recipient)
						if fromAddr, ok := params["from"].(string); ok {
							zeroAddr := common.Address{}
							if fromAddr == "0x0000000000000000000000000000000000000000" || fromAddr == zeroAddr.Hex() {
								// Extract tokenId from the Transfer event
								if tokenIdVal, ok := params["tokenId"]; ok {
									switch v := tokenIdVal.(type) {
									case *big.Int:
										nftTokenID = v
									case float64:
										nftTokenID = big.NewInt(int64(v))
									case string:
										if tokenIdBig, ok := new(big.Int).SetString(v, 10); ok {
											nftTokenID = tokenIdBig
										}
									}
									log.Printf("Extracted NFT token ID from mint receipt: %s", nftTokenID.String())
									break
								}
							}
						}
					}
				}
			}
		}
	}

	return nftTokenID
}
