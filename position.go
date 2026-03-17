package blackholedex

import (
	"blackholego/pkg/types"
	"blackholego/pkg/util"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

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
) (*types.StakingResult, error) {
	tickSpacing := b.poolType.TickSpacing()

	// T012: Input validation
	if err := util.ValidateStakingRequest(maxWAVAX, maxUSDC, rangeWidth, slippagePct); err != nil {
		return &types.StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("validation failed: %v", err),
		}, err
	}

	// Initialize transaction tracking
	var transactions []types.TransactionRecord

	// T013: Query pool state
	// wavaxUsdcPairAddr, _ := b.GetAddress(wavaxUsdcPair)
	state, err := b.GetAMMState()
	if err != nil {
		return &types.StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to query pool state: %v", err),
		}, fmt.Errorf("failed to query pool state: %w", err)
	}

	// T014: Calculate tick bounds
	log.Printf("CalculateTickBounds: %d,rangeWidth: %d, tickSpacing: %d", state.Tick, rangeWidth, tickSpacing)
	tickLower, tickUpper, err := util.CalculateTickBounds(state.Tick, rangeWidth, tickSpacing)
	if err != nil {
		return &types.StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to calculate tick bounds: %v", err),
		}, fmt.Errorf("failed to calculate tick bounds: %w", err)
	}
	log.Printf("CurrentTick: %d,TickLower: %d, TickUpper: %d", state.Tick, tickLower, tickUpper)
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

	// T032: For CL1 pools, automatically adjust range if utilization is low
	// This helps minimize wasted tokens by extending the range asymmetrically
	if b.poolType == types.CL1 && (utilization0.Cmp(big.NewInt(90)) < 0 || utilization1.Cmp(big.NewInt(90)) < 0) {
		originalTickLower := tickLower
		originalTickUpper := tickUpper

		log.Printf("🔄 CL1 Pool: Low capital utilization detected (WAVAX: %d%%, USDC: %d%%). Attempting to optimize range...",
			utilization0.Int64(), utilization1.Int64())

		optTickLower, optTickUpper, optAmount0, optAmount1, optErr := util.CalculateOptimalRangeWidthForCL1(
			state.Tick,
			rangeWidth,
			tickSpacing,
			state.SqrtPrice,
			maxWAVAX,
			maxUSDC,
			90, // 90% utilization threshold
			20, // Try up to 20 iterations
		)

		if optErr == nil {
			// Use optimized tick bounds
			tickLower = optTickLower
			tickUpper = optTickUpper
			amount0Desired = optAmount0
			amount1Desired = optAmount1

			// Recalculate utilization
			utilization0 = new(big.Int).Mul(amount0Desired, big.NewInt(100))
			utilization0.Div(utilization0, maxWAVAX)
			utilization1 = new(big.Int).Mul(amount1Desired, big.NewInt(100))
			utilization1.Div(utilization1, maxUSDC)

			log.Printf("✅ Optimized tick range: TickLower: %d → %d, TickUpper: %d → %d",
				originalTickLower, tickLower, originalTickUpper, tickUpper)
			log.Printf("✅ Improved Capital Utilization: WAVAX %d%%, USDC %d%%",
				utilization0.Int64(), utilization1.Int64())
		} else {
			log.Printf("⚠️  Failed to optimize range: %v", optErr)
		}
	}

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
		return &types.StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("balance validation failed: %v", err),
		}, fmt.Errorf("balance validation failed: %w", err)
	}

	// T017: Calculate slippage protection
	amount0Min := util.CalculateMinAmount(amount0Desired, slippagePct)
	amount1Min := util.CalculateMinAmount(amount1Desired, slippagePct)

	// Get contract clients
	wavaxClient, err := b.registry.Client(wavax)
	if err != nil {
		return &types.StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get WAVAX client: %v", err),
		}, fmt.Errorf("failed to get WAVAX client: %w", err)
	}

	usdcClient, err := b.registry.Client(usdc)
	if err != nil {
		return &types.StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get USDC client: %v", err),
		}, fmt.Errorf("failed to get USDC client: %w", err)
	}

	nftManagerAddr, _ := b.registry.GetAddress(nonfungiblePositionManager)

	// T018: WAVAX approval
	wavaxApproveTxHash, err := b.ensureApproval(wavaxClient, nftManagerAddr, amount0Desired)
	if err != nil {
		return &types.StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to approve WAVAX: %v", err),
		}, fmt.Errorf("failed to approve WAVAX: %w", err)
	}

	// Wait for WAVAX approval if transaction was sent
	if wavaxApproveTxHash != (common.Hash{}) {
		receipt, err := b.tl.WaitForTransaction(wavaxApproveTxHash)
		if err != nil {
			return &types.StakingResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("WAVAX approval transaction failed: %v", err),
			}, fmt.Errorf("WAVAX approval transaction failed: %w", err)
		}

		// T024: Extract gas cost
		gasCost, err := util.ExtractGasCost(receipt)
		if err != nil {
			return &types.StakingResult{
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

		transactions = append(transactions, types.TransactionRecord{
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
		return &types.StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to approve USDC: %v", err),
		}, fmt.Errorf("failed to approve USDC: %w", err)
	}

	// Wait for USDC approval if transaction was sent
	if usdcApproveTxHash != (common.Hash{}) {
		receipt, err := b.tl.WaitForTransaction(usdcApproveTxHash)
		if err != nil {
			return &types.StakingResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("USDC approval transaction failed: %v", err),
			}, fmt.Errorf("USDC approval transaction failed: %w", err)
		}

		// Extract gas cost
		gasCost, err := util.ExtractGasCost(receipt)
		if err != nil {
			return &types.StakingResult{
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

		transactions = append(transactions, types.TransactionRecord{
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
	wavaxAddr, _ := b.registry.GetAddress(wavax)
	usdcAddr, _ := b.registry.GetAddress(usdc)
	deployerAddr, _ := b.registry.GetAddress(deployer)
	mintParams := &types.MintParams{
		Token0:         wavaxAddr,
		Token1:         usdcAddr,
		Deployer:       deployerAddr,
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
	nftManagerClient, err := b.registry.Client(nonfungiblePositionManager)
	if err != nil {
		return &types.StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get NFT manager client: %v", err),
		}, fmt.Errorf("failed to get NFT manager client: %w", err)
	}

	// T022: Submit mint transaction
	mintTxHash, err := nftManagerClient.Send(
		types.Standard,
		&b.myAddr,
		b.privateKey,
		"mint",
		mintParams,
	)
	if err != nil {
		return &types.StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to submit mint transaction: %v", err),
		}, fmt.Errorf("failed to submit mint transaction: %w", err)
	}

	// T023: Wait for mint confirmation
	mintReceipt, err := b.tl.WaitForTransaction(mintTxHash)
	if err != nil {
		return &types.StakingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("mint transaction failed: %v", err),
		}, fmt.Errorf("mint transaction failed: %w", err)
	}

	// Extract gas cost for mint
	mintGasCost, err := util.ExtractGasCost(mintReceipt)
	if err != nil {
		return &types.StakingResult{
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

	transactions = append(transactions, types.TransactionRecord{
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

	result := &types.StakingResult{
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
	fmt.Printf("  NFT ID: %s", result.NFTTokenID.String())
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
) (*types.StakingResult, error) {
	// T007-T008: Input validation
	if nftTokenID == nil || nftTokenID.Sign() <= 0 {
		return &types.StakingResult{
			Success:      false,
			ErrorMessage: "validation failed: invalid token ID",
		}, fmt.Errorf("validation failed: invalid token ID")
	}

	// T009: Initialize transaction tracking
	var transactions []types.TransactionRecord

	// T011-T014: NFT Ownership Verification
	nftManagerClient, err := b.registry.Client(nonfungiblePositionManager)
	if err != nil {
		return &types.StakingResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get NFT manager client: %v", err),
		}, fmt.Errorf("failed to get NFT manager client: %w", err)
	}

	// Query NFT ownership
	ownerResult, err := nftManagerClient.Call(&b.myAddr, "ownerOf", nftTokenID)
	if err != nil {
		return &types.StakingResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to verify NFT %d ownership: %v", nftTokenID, err),
		}, fmt.Errorf("failed to verify NFT ownership: %w", err)
	}

	owner := ownerResult[0].(common.Address)
	if owner != b.myAddr {
		return &types.StakingResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("NFT not owned by wallet: owned by %s", owner.Hex()),
		}, fmt.Errorf("NFT not owned by wallet")
	}

	// T015-T023: NFT Approval Check and Execution
	approvalResult, err := nftManagerClient.Call(&b.myAddr, "getApproved", nftTokenID)
	if err != nil {
		return &types.StakingResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to check NFT %d approval: %v", nftTokenID, err),
		}, fmt.Errorf("failed to check NFT approval: %w", err)
	}

	currentApproval := approvalResult[0].(common.Address)

	// Only approve if not already approved for this gauge
	gaugeAddr, _ := b.registry.GetAddress(gauge)
	if currentApproval != gaugeAddr {
		log.Printf("Approving NFT %s for gauge %s", nftTokenID.String(), gaugeAddr.Hex())

		approveTxHash, err := nftManagerClient.Send(
			types.Standard,
			&b.myAddr,
			b.privateKey,
			"approve",
			gaugeAddr,
			nftTokenID,
		)
		if err != nil {
			return &types.StakingResult{
				NFTTokenID:   nftTokenID,
				Success:      false,
				ErrorMessage: fmt.Sprintf("failed to approve NFT: %v", err),
			}, fmt.Errorf("failed to approve NFT: %w", err)
		}

		// Wait for approval confirmation
		approvalReceipt, err := b.tl.WaitForTransaction(approveTxHash)
		if err != nil {
			return &types.StakingResult{
				NFTTokenID:   nftTokenID,
				Success:      false,
				ErrorMessage: fmt.Sprintf("NFT approval transaction failed: %v", err),
			}, fmt.Errorf("NFT approval transaction failed: %w", err)
		}

		// Track approval transaction
		gasCost, err := util.ExtractGasCost(approvalReceipt)
		if err != nil {
			return &types.StakingResult{
				NFTTokenID:   nftTokenID,
				Success:      false,
				ErrorMessage: fmt.Sprintf("failed to extract approval gas cost: %v", err),
			}, fmt.Errorf("failed to extract approval gas cost: %w", err)
		}

		gasPrice := new(big.Int)
		gasPrice.SetString(approvalReceipt.EffectiveGasPrice, 0)
		gasUsed := new(big.Int)
		gasUsed.SetString(approvalReceipt.GasUsed, 0)

		transactions = append(transactions, types.TransactionRecord{
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
	gaugeClient, err := b.registry.Client(gauge)
	if err != nil {
		// Return with partial transaction records if approval was sent
		totalGasCost := big.NewInt(0)
		for _, tx := range transactions {
			totalGasCost.Add(totalGasCost, tx.GasCost)
		}
		return &types.StakingResult{
			NFTTokenID:   nftTokenID,
			Transactions: transactions,
			TotalGasCost: totalGasCost,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get gauge client: %v", err),
		}, fmt.Errorf("failed to get gauge client: %w", err)
	}

	// Submit deposit transaction
	log.Printf("Depositing NFT %s into gauge %s", nftTokenID.String(), gaugeAddr.Hex())

	depositTxHash, err := gaugeClient.Send(
		types.Standard,
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
		return &types.StakingResult{
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
		return &types.StakingResult{
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
		return &types.StakingResult{
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

	transactions = append(transactions, types.TransactionRecord{
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

	result := &types.StakingResult{
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
	fmt.Printf("  Gauge: %s\n", gaugeAddr.Hex())
	fmt.Printf("  Total Gas Cost: %s wei\n", totalGasCost.String())
	for _, tx := range transactions {
		fmt.Printf("  - %s: %s (gas: %s wei)\n", tx.Operation, tx.TxHash.Hex(), tx.GasCost.String())
	}

	return result, nil
}

// executeUnstake calls the existing Unstake method with correct nonce (T025)
func (b *Blackhole) executeUnstake(
	nftTokenID *big.Int,
	nonce *big.Int,
	state *types.StrategyState,
	reportChan chan<- string,
) (*types.UnstakeResult, error) {
	sendReport(reportChan, types.StrategyReport{
		Timestamp:  time.Now(),
		EventType:  "rebalance_start",
		Message:    fmt.Sprintf("Unstaking NFT %s", nftTokenID.String()),
		Phase:      &state.CurrentState,
		NFTTokenID: nftTokenID,
	})

	result, err := b.Unstake(nftTokenID, nonce)
	if err != nil {
		return nil, fmt.Errorf("unstake failed: %w", err)
	}

	// Update cumulative gas
	state.CumulativeGas = new(big.Int).Add(state.CumulativeGas, result.TotalGasCost)
	sendReport(reportChan, types.StrategyReport{
		Timestamp:     time.Now(),
		EventType:     "gas_cost",
		Message:       "Unstake transaction completed",
		GasCost:       result.TotalGasCost,
		CumulativeGas: state.CumulativeGas,
		Profit:        result.Rewards.Reward,
		Phase:         &state.CurrentState,
	})

	return result, nil
}

/*
memo. nonce = unique identifier for a farming program incentive.
IncentiveKey에 대응되는 nonce 값을 사용해야만 함. 내 경우에는 3만을 사용.
"incentiveKeys" 함수를 호출하면 내 incentiveId에 대응되는 nonce를 알 수 있음
*/
func (b *Blackhole) Unstake(
	nftTokenID *big.Int,
	nonce *big.Int,
) (*types.UnstakeResult, error) {
	// T006: Input validation - NFT token ID
	if nftTokenID == nil || nftTokenID.Sign() <= 0 {
		return &types.UnstakeResult{
			Success:      false,
			ErrorMessage: "validation failed: invalid token ID",
		}, fmt.Errorf("validation failed: invalid token ID")
	}

	// Initialize transaction tracking
	var transactions []types.TransactionRecord

	// T008: Verify NFT ownership
	nftManagerClient, err := b.registry.Client(nonfungiblePositionManager)
	if err != nil {
		return &types.UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get NFT manager client: %v", err),
		}, fmt.Errorf("failed to get NFT manager client: %w", err)
	}

	ownerResult, err := nftManagerClient.Call(&b.myAddr, "ownerOf", nftTokenID)
	if err != nil {
		return &types.UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to verify NFT ownership: %v", err),
		}, fmt.Errorf("failed to verify NFT ownership: %w", err)
	}

	owner := ownerResult[0].(common.Address)
	if owner != b.myAddr {
		return &types.UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("NFT not owned by wallet: owned by %s", owner.Hex()),
		}, fmt.Errorf("NFT not owned by wallet")
	}

	// T009: Verify NFT is currently farmed
	farmingCenterClient, err := b.registry.Client(farmingCenter)
	if err != nil {
		return &types.UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get FarmingCenter client: %v", err),
		}, fmt.Errorf("failed to get FarmingCenter client: %w", err)
	}

	depositsResult, err := farmingCenterClient.Call(&b.myAddr, "deposits", nftTokenID)
	if err != nil {
		return &types.UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to check farming status: %v", err),
		}, fmt.Errorf("failed to check farming status: %w", err)
	}

	currentIncentiveId := depositsResult[0].([32]byte)
	if currentIncentiveId == [32]byte{} {
		return &types.UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: "NFT is not currently staked in farming",
		}, fmt.Errorf("NFT is not currently staked")
	}

	// T010: Build multicall data - encode exitFarming call
	var multicallData [][]byte

	blackAddr, _ := b.registry.GetAddress(black)
	algebraPoolAddr, _ := b.registry.GetAddress(wavaxUsdcPair)
	incentiveKey := types.IncentiveKey{
		RewardToken:      blackAddr,
		BonusRewardToken: blackAddr,
		Pool:             algebraPoolAddr,
		Nonce:            nonce,
	}

	farmingCenterABI := farmingCenterClient.Abi()
	exitFarmingData, err := farmingCenterABI.Pack("exitFarming", incentiveKey, nftTokenID)
	if err != nil {
		return &types.UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to encode exitFarming: %v", err),
		}, fmt.Errorf("failed to encode exitFarming: %w", err)
	}
	multicallData = append(multicallData, exitFarmingData)

	// T011: Conditionally encode collectRewards call
	collectRewardsData, err := farmingCenterABI.Pack("claimReward", blackAddr, b.myAddr, big.NewInt(0)) // todo. reward 0원인거 확인.
	if err != nil {
		return &types.UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to encode collectRewards: %v", err),
		}, fmt.Errorf("failed to encode collectRewards: %w", err)
	}
	multicallData = append(multicallData, collectRewardsData)

	// T012: Execute multicall transaction
	farmingCenterAddr, _ := b.registry.GetAddress(farmingCenter)
	log.Printf("Unstaking NFT %s from FarmingCenter %s", nftTokenID.String(), farmingCenterAddr.Hex())

	multicallTxHash, err := farmingCenterClient.Send(
		types.Standard,
		&b.myAddr,
		b.privateKey,
		"multicall",
		multicallData,
	)
	if err != nil {
		return &types.UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to submit multicall transaction: %v", err),
		}, fmt.Errorf("failed to submit multicall transaction: %w", err)
	}

	// T013: Wait for transaction confirmation and extract gas cost
	multicallReceipt, err := b.tl.WaitForTransaction(multicallTxHash)
	if err != nil {
		return &types.UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("multicall transaction failed: %v", err),
		}, fmt.Errorf("multicall transaction failed: %w", err)
	}

	gasCost, err := util.ExtractGasCost(multicallReceipt)
	if err != nil {
		return &types.UnstakeResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to extract gas cost: %v", err),
		}, fmt.Errorf("failed to extract gas cost: %w", err)
	}

	gasPrice := new(big.Int)
	gasPrice.SetString(multicallReceipt.EffectiveGasPrice, 0)
	gasUsed := new(big.Int)
	gasUsed.SetString(multicallReceipt.GasUsed, 0)

	transactions = append(transactions, types.TransactionRecord{
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
	rewards := &types.RewardAmounts{
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

	result := &types.UnstakeResult{
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
	fmt.Printf("  FarmingCenter: %s\n", farmingCenterAddr.Hex())
	if rewards != nil {
		fmt.Printf("  Rewards: %s / %s\n", rewards.Reward.String(), rewards.BonusReward.String())
	}
	fmt.Printf("  Total Gas Cost: %s wei\n", totalGasCost.String())
	for _, tx := range transactions {
		fmt.Printf("  - %s: %s (gas: %s wei)\n", tx.Operation, tx.TxHash.Hex(), tx.GasCost.String())
	}

	return result, nil
}

// executeWithdraw calls the existing Withdraw method and tracks results (T026)
func (b *Blackhole) executeWithdraw(
	nftTokenID *big.Int,
	state *types.StrategyState,
	reportChan chan<- string,
) (*types.WithdrawResult, error) {
	sendReport(reportChan, types.StrategyReport{
		Timestamp:  time.Now(),
		EventType:  "rebalance_start",
		Message:    fmt.Sprintf("Withdrawing liquidity from NFT %s", nftTokenID.String()),
		Phase:      &state.CurrentState,
		NFTTokenID: nftTokenID,
	})

	result, err := b.Withdraw(nftTokenID)
	if err != nil {
		return nil, fmt.Errorf("withdraw failed: %w", err)
	}

	// Update cumulative gas
	state.CumulativeGas = new(big.Int).Add(state.CumulativeGas, result.TotalGasCost)
	sendReport(reportChan, types.StrategyReport{
		Timestamp:     time.Now(),
		EventType:     "gas_cost",
		Message:       "Withdraw transaction completed",
		GasCost:       result.TotalGasCost,
		CumulativeGas: state.CumulativeGas,
		Phase:         &state.CurrentState,
	})

	return result, nil
}

// Withdraw removes all liquidity from an NFT position and burns the NFT
// nftTokenID: ERC721 token ID from previous Mint operation
// Returns WithdrawResult with transaction tracking and gas costs
func (b *Blackhole) Withdraw(nftTokenID *big.Int) (*types.WithdrawResult, error) {
	// T008: Input validation
	if nftTokenID == nil || nftTokenID.Sign() <= 0 {
		return &types.WithdrawResult{
			Success:      false,
			ErrorMessage: "validation failed: NFT token ID must be positive",
		}, fmt.Errorf("validation failed: NFT token ID must be positive")
	}

	// T009: Get nonfungiblePositionManager ContractClient
	nftManagerClient, err := b.registry.Client(nonfungiblePositionManager)
	if err != nil {
		return &types.WithdrawResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get NFT manager client: %v", err),
		}, fmt.Errorf("failed to get NFT manager client: %w", err)
	}

	// T010: Verify NFT ownership
	ownerResult, err := nftManagerClient.Call(&b.myAddr, "ownerOf", nftTokenID)
	if err != nil {
		return &types.WithdrawResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to verify NFT ownership: %v", err),
		}, fmt.Errorf("failed to verify NFT ownership: %w", err)
	}

	owner := ownerResult[0].(common.Address)
	if owner != b.myAddr {
		return &types.WithdrawResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("NFT not owned by wallet: owned by %s", owner.Hex()),
		}, fmt.Errorf("NFT not owned by wallet: owned by %s", owner.Hex())
	}

	// T011: Query position details to get liquidity amount
	positionsResult, err := nftManagerClient.Call(&b.myAddr, "positions", nftTokenID)
	if err != nil {
		return &types.WithdrawResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to query position: %v", err),
		}, fmt.Errorf("failed to query position: %w", err)
	}

	liquidity := positionsResult[7].(*big.Int) // uint128 liquidity at index 7

	// T012-T016: Build multicall data
	// The multicall will execute three operations atomically in this order:
	// 1. decreaseLiquidity: Removes liquidity from the position (tokens become withdrawable)
	// 2. collect: Actually transfers the tokens and fees to the recipient
	// 3. burn: Destroys the NFT after all tokens are collected
	// If any operation fails, the entire transaction reverts (atomicity guarantee)
	var multicallData [][]byte
	deadline := big.NewInt(time.Now().Add(20 * time.Minute).Unix())

	// Slippage protection via amount0Min/amount1Min
	// These minimums protect against price manipulation and sandwich attacks
	// For now use zero minimums (production should calculate based on slippage percentage)
	// TODO: Calculate proper minimums: amount0Min = expectedAmount0 * (100 - slippagePct) / 100
	amount0Min := big.NewInt(0)
	amount1Min := big.NewInt(0)

	// T012-T013: Encode decreaseLiquidity
	decreaseParams := &types.DecreaseLiquidityParams{
		TokenId:    nftTokenID,
		Liquidity:  liquidity,
		Amount0Min: amount0Min,
		Amount1Min: amount1Min,
		Deadline:   deadline,
	}

	nftManagerABI := nftManagerClient.Abi()
	decreaseData, err := nftManagerABI.Pack("decreaseLiquidity", decreaseParams)
	if err != nil {
		return &types.WithdrawResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to encode decreaseLiquidity: %v", err),
		}, fmt.Errorf("failed to encode decreaseLiquidity: %w", err)
	}
	multicallData = append(multicallData, decreaseData)

	// T014-T015: Encode collect
	maxUint128 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	collectParams := &types.CollectParams{
		TokenId:    nftTokenID,
		Recipient:  b.myAddr,
		Amount0Max: maxUint128,
		Amount1Max: maxUint128,
	}

	collectData, err := nftManagerABI.Pack("collect", collectParams)
	if err != nil {
		return &types.WithdrawResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to encode collect: %v", err),
		}, fmt.Errorf("failed to encode collect: %w", err)
	}
	multicallData = append(multicallData, collectData)

	// T016: Encode burn
	burnData, err := nftManagerABI.Pack("burn", nftTokenID)
	if err != nil {
		return &types.WithdrawResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to encode burn: %v", err),
		}, fmt.Errorf("failed to encode burn: %w", err)
	}
	multicallData = append(multicallData, burnData)

	// T017: Execute multicall transaction
	txHash, err := nftManagerClient.Send(
		types.Standard,
		&b.myAddr,
		b.privateKey,
		"multicall",
		multicallData,
	)
	if err != nil {
		return &types.WithdrawResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to submit multicall transaction: %v", err),
		}, fmt.Errorf("failed to submit multicall transaction: %w", err)
	}

	// T018: Wait for transaction confirmation
	receipt, err := b.tl.WaitForTransaction(txHash)
	if err != nil {
		return &types.WithdrawResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("multicall transaction failed: %v", err),
		}, fmt.Errorf("multicall transaction failed: %w", err)
	}

	// T019: Extract gas cost from receipt
	gasCost, err := util.ExtractGasCost(receipt)
	if err != nil {
		return &types.WithdrawResult{
			NFTTokenID:   nftTokenID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to extract gas cost: %v", err),
		}, fmt.Errorf("failed to extract gas cost: %w", err)
	}

	gasPrice := new(big.Int)
	gasPrice.SetString(receipt.EffectiveGasPrice, 0)
	gasUsed := new(big.Int)
	gasUsed.SetString(receipt.GasUsed, 0)

	// T020: Create TransactionRecord
	var transactions []types.TransactionRecord
	transactions = append(transactions, types.TransactionRecord{
		TxHash:    txHash,
		GasUsed:   gasUsed.Uint64(),
		GasPrice:  gasPrice,
		GasCost:   gasCost,
		Timestamp: time.Now(),
		Operation: "Withdraw",
	})

	// T021: Build and return WithdrawResult
	result := &types.WithdrawResult{
		NFTTokenID:   nftTokenID,
		Amount0:      big.NewInt(0), // Will be enhanced in Polish phase to parse from multicall results
		Amount1:      big.NewInt(0), // Will be enhanced in Polish phase to parse from multicall results
		Transactions: transactions,
		TotalGasCost: gasCost,
		Success:      true,
		ErrorMessage: "",
	}

	// T022: Add success logging
	fmt.Printf("✓ Liquidity withdrawn successfully\n")
	fmt.Printf("  NFT ID: %s\n", nftTokenID.String())
	fmt.Printf("  Gas cost: %s wei\n", gasCost.String())

	return result, nil
}

// executeRebalancing orchestrates the full rebalancing workflow (T027-T034)
// Steps: unstake → withdraw → calculate rebalance → swap → update state
// Does NOT create new position - that happens after stability check
// Supports checkpoint/resume: resumes from state.CurrentStep if retrying after failure
func (b *Blackhole) executeRebalancing(
	config *types.StrategyConfig,
	state *types.StrategyState,
	nonce *big.Int,
	reportChan chan<- string,
) (*types.RebalanceWorkflow, error) {

	// T028: Create RebalanceWorkflow for tracking
	workflow := &types.RebalanceWorkflow{
		StartTime:    time.Now(),
		OldPosition:  nil, // Will be populated if we query position details
		SwapResults:  []types.TransactionRecord{},
		TotalGas:     big.NewInt(0),
		Success:      false,
		ErrorMessage: "",
	}

	sendReport(reportChan, types.StrategyReport{
		Timestamp: time.Now(),
		EventType: "rebalance_start",
		Message:   fmt.Sprintf("Starting rebalancing workflow from step: %s", state.CurrentStep.String()),
		Phase:     &state.CurrentState,
	})

	if state.NFTTokenID == nil {
		nftId, err := b.TokenOfOwnerByIndex(big.NewInt(0))
		if err != nil {
			workflow.Success = false
			workflow.ErrorMessage = err.Error()
			return workflow, err
		}
		state.NFTTokenID = nftId
	}

	// Step: Execute unstake (skip if already completed)
	if state.CurrentStep < types.Step_Rebalance_UnstakeCompleted {
		unstakeResult, err := b.executeUnstake(state.NFTTokenID, nonce, state, reportChan)
		if err != nil {
			workflow.Success = false
			workflow.ErrorMessage = err.Error()
			return workflow, err
		}

		// T030: Track cumulative gas
		workflow.TotalGas = new(big.Int).Add(workflow.TotalGas, unstakeResult.TotalGasCost)

		// T031: Track cumulative rewards
		if unstakeResult.Rewards != nil {
			state.CumulativeRewards = new(big.Int).Add(state.CumulativeRewards, unstakeResult.Rewards.Reward)
		}

		// Checkpoint: unstake completed
		state.CurrentStep = types.Step_Rebalance_UnstakeCompleted
		log.Printf("[Checkpoint] Unstake completed: NFT ID=%s, gas=%s", state.NFTTokenID.String(), unstakeResult.TotalGasCost.String())
	} else {
		log.Printf("[Resume] Unstake already completed, NFT ID=%s", state.NFTTokenID.String())
	}

	// Step: Execute withdraw (skip if already completed)
	if state.CurrentStep < types.Step_Rebalance_WithdrawCompleted {
		withdrawResult, err := b.executeWithdraw(state.NFTTokenID, state, reportChan)
		if err != nil {
			workflow.Success = false
			workflow.ErrorMessage = err.Error()
			return workflow, err
		}

		workflow.WithdrawResult = withdrawResult
		// T030: Track cumulative gas
		workflow.TotalGas = new(big.Int).Add(workflow.TotalGas, withdrawResult.TotalGasCost)

		// Checkpoint: withdraw completed
		state.CurrentStep = types.Step_Rebalance_WithdrawCompleted
		log.Printf("[Checkpoint] Withdraw completed: NFT ID=%s, amount0=%s, amount1=%s, gas=%s",
			state.NFTTokenID.String(), withdrawResult.Amount0.String(), withdrawResult.Amount1.String(), withdrawResult.TotalGasCost.String())
	} else {
		log.Printf("[Resume] Withdraw already completed, NFT ID=%s", state.NFTTokenID.String())
	}

	// T032, T033: Calculate and report net P&L
	netPnL := new(big.Int).Sub(state.CumulativeRewards, state.CumulativeGas)
	netPnL = new(big.Int).Sub(netPnL, state.TotalSwapFees)

	sendReport(reportChan, types.StrategyReport{
		Timestamp:     time.Now(),
		EventType:     "profit",
		Message:       "Rebalancing workflow completed (unstake + withdrawal)",
		CumulativeGas: state.CumulativeGas,
		Profit:        state.CumulativeRewards,
		NetPnL:        netPnL,
		Phase:         &state.CurrentState,
	})

	workflow.Duration = time.Since(workflow.StartTime)
	workflow.Success = true

	// Reset step counter for next phase
	state.CurrentStep = types.Step_None
	log.Printf("[Phase Complete] RebalancingRequired phase completed, resetting step to None")

	return workflow, nil
}
