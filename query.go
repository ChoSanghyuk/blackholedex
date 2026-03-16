package blackholedex

import (
	"blackholego/pkg/types"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// GetAMMState retrieves the current state of an AMM pool
// This is a read-only operation that does not create a transaction
func (b *Blackhole) GetAMMState() (*types.AMMState, error) {
	poolClient, err := b.registry.Client(wavaxUsdcPair)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool client for %s: %w", wavaxUsdcPair, err)
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
	state := &types.AMMState{
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
	wavaxClient, err := b.registry.Client(wavax)
	if err != nil {
		return fmt.Errorf("failed to get WAVAX client: %w", err)
	}

	usdcClient, err := b.registry.Client(usdc)
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

// Unstake withdraws a staked NFT position from FarmingCenter
// nftTokenID: ERC721 token ID from previous Mint operation
// incentiveKey: Identifies the farming program to exit
// collectRewards: Whether to claim accumulated rewards during unstake
// Returns UnstakeResult with transaction tracking and gas costs

func (b *Blackhole) TokenOfOwnerByIndex(index *big.Int) (*big.Int, error) {
	nftManagerClient, err := b.registry.Client(nonfungiblePositionManager)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFT manager client: %w", err)
	}
	rtnRaw, err := nftManagerClient.Call(nil, "tokenOfOwnerByIndex", b.myAddr, index)
	if err != nil {
		return nil, fmt.Errorf("failed to call tokenOfOwnerByIndex: %w", err)
	}

	return rtnRaw[0].(*big.Int), nil
}

// GetUserPositions retrieves all NFT position token IDs owned by the user
// Returns a slice of token IDs and an error if the operation fails
func (b *Blackhole) GetUserPositions() ([]*big.Int, error) {
	nftManagerClient, err := b.registry.Client(nonfungiblePositionManager)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFT manager client: %w", err)
	}

	// Get the balance of NFTs owned by the user
	balanceResult, err := nftManagerClient.Call(nil, "balanceOf", b.myAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFT balance: %w", err)
	}

	balance := balanceResult[0].(*big.Int)
	if balance.Sign() == 0 {
		return []*big.Int{}, nil // No positions owned
	}

	// Iterate through all token IDs
	tokenIDs := make([]*big.Int, 0, balance.Int64())
	for i := int64(0); i < balance.Int64(); i++ {
		tokenID, err := b.TokenOfOwnerByIndex(big.NewInt(i))
		if err != nil {
			return nil, fmt.Errorf("failed to get token ID at index %d: %w", i, err)
		}
		tokenIDs = append(tokenIDs, tokenID)
	}

	return tokenIDs, nil
}

// monitoringLoop continuously monitors pool price and detects out-of-range conditions (T035-T041)
// Returns true if out-of-range detected, false otherwise, or error
func (b *Blackhole) monitoringLoop(
	ctx context.Context,
	state *types.StrategyState,
	reportChan chan<- string,
) (bool, error) {
	// T034: Check context cancellation
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	// T036: Get current pool state
	// wavaxUsdcPairAddr, _ := b.GetAddress(wavaxUsdcPair)
	poolState, err := b.GetAMMState()
	if err != nil {
		return false, fmt.Errorf("failed to get pool state: %w", err)
	}

	// Update last observed price
	state.LastPrice = poolState.SqrtPrice

	// T037: Check if position is out of range
	positionRange := &types.PositionRange{
		TickLower: state.TickLower,
		TickUpper: state.TickUpper,
	}

	isOutOfRange := positionRange.IsOutOfRange(poolState.Tick)

	// T039: Send monitoring report
	// sendReport(b, reportChan, StrategyReport{
	// 	Timestamp: time.Now(),
	// 	EventType: "monitoring",
	// 	Message:   fmt.Sprintf("Price check: tick=%d, range=[%d, %d], out_of_range=%v", poolState.Tick, state.TickLower, state.TickUpper, isOutOfRange),
	// 	Phase:     &state.CurrentState,
	// }, false)
	log.Printf("[monitoring] Price check: tick=%d, range=[%d, %d], out_of_range=%v\n", poolState.Tick, state.TickLower, state.TickUpper, isOutOfRange)

	// T038: Transition to RebalancingRequired if out of range
	if isOutOfRange {
		state.CurrentState = types.RebalancingRequired
		sendReport(reportChan, types.StrategyReport{
			Timestamp:  time.Now(),
			EventType:  "out_of_range",
			Message:    fmt.Sprintf("Position out of range detected: current tick %d outside [%d, %d]", poolState.Tick, state.TickLower, state.TickUpper),
			Phase:      &state.CurrentState,
			NFTTokenID: state.NFTTokenID,
		}) // State changed to RebalancingRequired
		return true, nil
	}

	return false, nil
}

// GetPositionDetails retrieves the detailed information for a specific position NFT
// Returns a Position struct containing all position data
func (b *Blackhole) GetPositionDetails(tokenID *big.Int) (*types.Position, error) {
	if tokenID == nil || tokenID.Sign() <= 0 {
		return nil, fmt.Errorf("invalid token ID: must be positive")
	}

	nftManagerClient, err := b.registry.Client(nonfungiblePositionManager)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFT manager client: %w", err)
	}

	// Call positions(tokenId) function
	positionResult, err := nftManagerClient.Call(nil, "positions", tokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to get position details for token ID %s: %w", tokenID.String(), err)
	}

	// Parse the returned values according to the ABI
	// positions() returns: (nonce, operator, token0, token1, deployer, tickLower, tickUpper,
	//                       liquidity, feeGrowthInside0LastX128, feeGrowthInside1LastX128,
	//                       tokensOwed0, tokensOwed1)
	position := &types.Position{
		Nonce:                    positionResult[0].(*big.Int),
		Operator:                 positionResult[1].(common.Address),
		Token0:                   positionResult[2].(common.Address),
		Token1:                   positionResult[3].(common.Address),
		Deployer:                 positionResult[4].(common.Address),
		TickLower:                int32(positionResult[5].(*big.Int).Int64()),
		TickUpper:                int32(positionResult[6].(*big.Int).Int64()),
		Liquidity:                positionResult[7].(*big.Int),
		FeeGrowthInside0LastX128: positionResult[8].(*big.Int),
		FeeGrowthInside1LastX128: positionResult[9].(*big.Int),
		TokensOwed0:              positionResult[10].(*big.Int),
		TokensOwed1:              positionResult[11].(*big.Int),
	}

	return position, nil
}

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
				if eventName, ok := event["event"].(string); ok && eventName == "Transfer" {
					if params, ok := event["parameter"].(map[string]interface{}); ok {
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
