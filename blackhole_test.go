package blackholedex

import (
	"blackholego/internal/util"
	"blackholego/pkg/contractclient"
	"blackholego/pkg/txlistener"
	"crypto/ecdsa"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func TestBlackhole(t *testing.T) {
	// Load environment variables
	err := godotenv.Load(".env.test.local")
	if err != nil {
		t.Fatalf("Failed to load .env.test.local: %v", err)
	}

	// Get private key
	// pk := os.Getenv("PK")
	// if pk == "" {
	// 	t.Fatal("PK not set")
	// }
	// privateKey, err := crypto.HexToECDSA(pk)
	// if err != nil {
	// 	t.Fatalf("Failed to parse private key: %v", err)
	// }
	// publicKey := privateKey.Public()
	// publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	// if !ok {
	// 	t.Fatal("error casting public key to ECDSA")
	// }
	// address := crypto.PubkeyToAddress(*publicKeyECDSA)

	var privateKey *ecdsa.PrivateKey = nil
	addrHex := os.Getenv("ADDRESS")
	address := common.HexToAddress(addrHex)

	// Connect to RPC
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		t.Fatal("RPC_URL not set in .env.test.local")
	}
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		t.Fatalf("Failed to connect to RPC: %v", err)
	}

	// Setup Router contract
	// routerAddr := os.Getenv("ROUTERV2_ADDR")
	// if routerAddr == "" {
	// 	t.Fatal("ROUTERV2_ADDR not set in .env.test.local")
	// }
	routerABIPath := os.Getenv("ROUTERV2_ABI_PATH")
	if routerABIPath == "" {
		t.Fatal("ROUTERV2_ABI_PATH not set in .env.test.local")
	}
	routerABI, err := util.LoadABIFromHardhatArtifact(routerABIPath)
	if err != nil {
		t.Fatalf("Failed to load router ABI: %v", err)
	}
	swapClient := contractclient.NewContractClient(client, common.HexToAddress(routerv2), routerABI)

	// Setup ERC20 ABI (assume ERC20 standard ABI is available)
	erc20ABIPath := os.Getenv("ERC20_ABI_PATH")
	if erc20ABIPath == "" {
		t.Fatal("ERC20_ABI_PATH not set in .env.test.local")
	}
	erc20ABI, err := util.LoadABI(erc20ABIPath)
	if err != nil {
		t.Fatalf("Failed to load ERC20 ABI: %v", err)
	}

	// Get token addresses
	// usdcAddr := os.Getenv("USDC_ADDR")
	// if usdcAddr == "" {
	// 	t.Fatal("USDC_ADDR not set in .env.test.local")
	// }
	// wavaxAddr := os.Getenv("WAVAX_ADDR")
	// if wavaxAddr == "" {
	// 	t.Fatal("WAVAX_ADDR not set in .env.test.local")
	// }

	// Create ERC20 clients
	usdcClient := contractclient.NewContractClient(client, common.HexToAddress(usdc), erc20ABI)
	wavaxClient := contractclient.NewContractClient(client, common.HexToAddress(wavax), erc20ABI)

	poolStateABIPath := os.Getenv("POOLSTATE_ABI_PATH")
	if erc20ABIPath == "" {
		t.Fatal("ERC20_ABI_PATH not set in .env.test.local")
	}
	poolStateABI, err := util.LoadABI(poolStateABIPath)
	if err != nil {
		t.Fatalf("Failed to load ERC20 ABI: %v", err)
	}

	// Create Wavax/usdc pool clients
	wausPoolClient := contractclient.NewContractClient(client, common.HexToAddress(wavaxUsdcPair), poolStateABI)

	// Setup TxListener
	listener := txlistener.NewTxListener(
		client,
		txlistener.WithPollInterval(2*time.Second),
		txlistener.WithTimeout(5*time.Minute),
	)

	// Create Blackhole instance
	b := &Blackhole{
		privateKey: privateKey,
		myAddr:     address,
		tl:         listener,
		ccm: map[string]ContractClient{
			routerv2:      swapClient,
			usdc:          usdcClient,
			wavax:         wavaxClient,
			wavaxUsdcPair: wausPoolClient,
		},
	}

	t.Run("SwapTokens", func(t *testing.T) {
		// Get swap parameters from environment
		amountIn := os.Getenv("SWAP_AMOUNT_IN")
		if amountIn == "" {
			t.Skip("SWAP_AMOUNT_IN not set, skipping swap test")
		}
		amountInBig := new(big.Int)
		amountInBig.SetString(amountIn, 10)

		amountOutMin := os.Getenv("SWAP_AMOUNT_OUT_MIN")
		if amountOutMin == "" {
			t.Skip("SWAP_AMOUNT_OUT_MIN not set, skipping swap test")
		}
		amountOutMinBig := new(big.Int)
		amountOutMinBig.SetString(amountOutMin, 10)

		// Create swap parameters
		deadline := big.NewInt(time.Now().Add(20 * time.Minute).Unix())
		params := &SWAPExactTokensForTokensParams{
			AmountIn:     amountInBig,
			AmountOutMin: amountOutMinBig,
			Routes: []Route{
				{
					Pair:         common.HexToAddress(wavaxUsdcPair),
					From:         common.HexToAddress(wavax),
					To:           common.HexToAddress(usdc),
					Stable:       true,
					Concentrated: true,
					Receiver:     b.myAddr,
				},
			},
			To:       address,
			Deadline: deadline,
		}

		// Execute swap
		t.Logf("Executing swap: %s WAVX -> USDC", amountIn)
		t.Logf("From address: %s", address.Hex())
		t.Logf("Router address: %s", routerv2)

		txHash, err := b.Swap(params)
		if err != nil {
			t.Fatalf("Swap failed: %v", err)
		}

		t.Logf("Swap transaction submitted: %s", txHash.Hex())

		// Wait for transaction confirmation
		receipt, err := listener.WaitForTransaction(txHash)
		if err != nil {
			t.Fatalf("Failed to wait for transaction: %v", err)
		}

		t.Logf("Swap confirmed in block %s", receipt.BlockNumber)
		t.Logf("Gas used: %s", receipt.GasUsed)
		t.Logf("Status: %s", receipt.Status)

		if receipt.Status != "0x1" {
			t.Fatalf("Swap transaction failed with status: %s", receipt.Status)
		}
	})

	t.Run("Mint", func(t *testing.T) {

		clg := 200 // CL Gap
		lpg := 3   // Liquidity Providing Gap
		// maxP := 0.8 // max portion
		var slippage int64 = 10

		state, err := b.GetAMMState(common.HexToAddress(wavaxUsdcPair))
		if err != nil {
			t.Fatalf("Failed to call GetAMMState: %v", err)
		}

		tickLower := (int(state.Tick)/clg - lpg) * 200
		tickUpper := (int(state.Tick)/clg + lpg) * 200

		wavaxClient, err := b.Client(wavax)
		outputs, err := wavaxClient.Call(&b.myAddr, "balanceOf", b.myAddr)
		wavaxBalace := outputs[0].(*big.Int)
		// wavaxMax := wavaxBalace.Sub()

		usdcClient, err := b.Client(usdc)
		outputs, err = usdcClient.Call(&b.myAddr, "balanceOf", b.myAddr)
		usdcBalace := outputs[0].(*big.Int)

		amount0Desired, amount1Desired, l := util.ComputeAmounts(state.SqrtPrice, int(state.Tick), tickLower, tickUpper, wavaxBalace, usdcBalace)
		t.Logf("liquidity : %v\n", l)

		deadline := big.NewInt(time.Now().Add(20 * time.Minute).Unix())
		params := &MintParams{
			Token0:         common.HexToAddress(wavax),
			Token1:         common.HexToAddress(usdc),
			Deployer:       common.HexToAddress(deployer),
			TickLower:      big.NewInt(int64(tickLower)),
			TickUpper:      big.NewInt(int64(tickUpper)),
			Amount0Desired: amount0Desired,
			Amount1Desired: amount1Desired,
			Amount0Min:     amount0Desired.Mul(amount0Desired, big.NewInt(100-slippage)).Div(amount0Desired, big.NewInt(100)),
			Amount1Min:     amount1Desired.Mul(amount1Desired, big.NewInt(100-slippage)).Div(amount0Desired, big.NewInt(100)),
			Recipient:      b.myAddr,
			Deadline:       deadline,
		}
		t.Logf("MintParams : %v\n", params)

	})
}
