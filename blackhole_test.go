package blackholedex

import (
	"blackholego/pkg/contractclient"
	"blackholego/pkg/txlistener"
	"blackholego/pkg/types"
	"blackholego/pkg/util"
	"crypto/ecdsa"

	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func TestBlackhole(t *testing.T) {
	// Load environment variables
	env := ".env.test.local"
	err := godotenv.Load(env)
	if err != nil {
		t.Fatalf("Failed to load .env.test.local: %v", err)
	}

	// Get private key
	// pk := os.Getenv("PK")
	// if pk == "" {
	// 	t.Fatal("PK not set")
	// }

	// Get private key
	var address common.Address
	var privateKey *ecdsa.PrivateKey

	encryptedPk := os.Getenv("ENC_PK")
	if encryptedPk != "" {
		key := os.Getenv("KEY")
		if key == "" {
			panic("PK not set")
		}

		pk, err := util.Decrypt([]byte(key), encryptedPk)
		if err != nil {
			panic(err)
		}
		privateKey, err = crypto.HexToECDSA(pk)
		if err != nil {
			t.Fatalf("Failed to parse private key: %v", err)
		}
		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			t.Fatal("error casting public key to ECDSA")
		}
		address = crypto.PubkeyToAddress(*publicKeyECDSA)
	} else {
		privateKey = nil
		addrHex := os.Getenv("CALLER_ADDRESS")
		address = common.HexToAddress(addrHex)
	}

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
	routerAddr := os.Getenv("ROUTERV2_ADDR")
	if routerAddr == "" {
		t.Fatal("ROUTERV2_ADDR not set in .env.test.local")
	}
	routerABIPath := os.Getenv("ROUTERV2_ABI_PATH")
	if routerABIPath == "" {
		t.Fatal("ROUTERV2_ABI_PATH not set in .env.test.local")
	}
	routerABI, err := util.LoadABIFromHardhatArtifact(routerABIPath)
	if err != nil {
		t.Fatalf("Failed to load router ABI: %v", err)
	}
	swapClient := contractclient.NewContractClient(client, common.HexToAddress(routerAddr), routerABI)

	// Create ERC20 clients
	usdcAddr := os.Getenv("USDC_ADDR")
	if usdcAddr == "" {
		t.Fatal("USDC_ADDR not set in .env.test.local")
	}
	wavaxAddr := os.Getenv("WAVAX_ADDR")
	if wavaxAddr == "" {
		t.Fatal("WAVAX_ADDR not set in .env.test.local")
	}
	blackAddr := os.Getenv("BLACK_ADDR")
	if blackAddr == "" {
		t.Fatal("BLACK_ADDR not set in .env.test.local")
	}
	erc20ABIPath := os.Getenv("ERC20_ABI_PATH")
	if erc20ABIPath == "" {
		t.Fatal("ERC20_ABI_PATH not set in .env.test.local")
	}
	erc20ABI, err := util.LoadABI(erc20ABIPath)
	if err != nil {
		t.Fatalf("Failed to load ERC20 ABI: %v", err)
	}
	usdcClient := contractclient.NewContractClient(client, common.HexToAddress(usdcAddr), erc20ABI)
	wavaxClient := contractclient.NewContractClient(client, common.HexToAddress(wavaxAddr), erc20ABI)
	blackClient := contractclient.NewContractClient(client, common.HexToAddress(blackAddr), erc20ABI)

	// Create Wavax/usdc pool clients
	poolAddr := os.Getenv("POOL_ADDR")
	if poolAddr == "" {
		t.Fatal("POOL_ADDR not set in .env.test.local")
	}
	poolStateABIPath := os.Getenv("POOL_ABI_PATH")
	if poolStateABIPath == "" {
		t.Fatal("POOL_ABI_PATH not set in .env.test.local")
	}
	poolStateABI, err := util.LoadABI(poolStateABIPath)
	if err != nil {
		t.Fatalf("Failed to load pool state ABI: %v", err)
	}
	wausPoolClient := contractclient.NewContractClient(client, common.HexToAddress(poolAddr), poolStateABI)

	// Create NFTPositionManager clients
	nftManagerAddr := os.Getenv("NFTMANAGER_ADDR")
	if nftManagerAddr == "" {
		t.Fatal("NFTMANAGER_ADDR not set in .env.test.local")
	}
	nftPositionManagerABIPath := os.Getenv("NFTMANAGER_ABI_PATH")
	if nftPositionManagerABIPath == "" {
		t.Fatal("NFTMANAGER_ABI_PATH not set in .env.test.local")
	}
	nftPositionManagerABI, err := util.LoadABI(nftPositionManagerABIPath)
	if err != nil {
		t.Fatalf("Failed to load NFT manager ABI: %v", err)
	}
	nftPositionManagerClient := contractclient.NewContractClient(client, common.HexToAddress(nftManagerAddr), nftPositionManagerABI)

	// Create Gauge clients
	gaugeAddr := os.Getenv("GAUGE_ADDR")
	if gaugeAddr == "" {
		t.Fatal("GAUGE_ADDR not set in .env.test.local")
	}
	gaugeABIPath := os.Getenv("GAUGE_ABI_PATH")
	if gaugeABIPath == "" {
		t.Fatal("GAUGE_ABI_PATH not set in .env.test.local")
	}
	gaugeABI, err := util.LoadABI(gaugeABIPath)
	if err != nil {
		t.Fatalf("Failed to load gauge ABI: %v", err)
	}
	gaugeClient := contractclient.NewContractClient(client, common.HexToAddress(gaugeAddr), gaugeABI)

	// Create FarmingCenter clients
	farmingCenterAddr := os.Getenv("FARMING_CENTER_ADDR")
	if farmingCenterAddr == "" {
		t.Fatal("FARMING_CENTER_ADDR not set in .env.test.local")
	}
	farmingCenterABIPath := os.Getenv("FARMING_CENTER_ABI_PATH")
	if farmingCenterABIPath == "" {
		t.Fatal("FARMING_CENTER_ABI_PATH not set in .env.test.local")
	}
	farmingCenterABI, err := util.LoadABI(farmingCenterABIPath)
	if err != nil {
		t.Fatalf("Failed to load farming center ABI: %v", err)
	}
	farmingCenterClient := contractclient.NewContractClient(client, common.HexToAddress(farmingCenterAddr), farmingCenterABI)

	// Create Deployer client
	deployerAddr := os.Getenv("DEPLOYER_ADDR")
	if deployerAddr == "" {
		t.Fatal("DEPLOYER_ADDR not set in .env.test.local")
	}

	deployerClient := contractclient.NewContractClient(client, common.HexToAddress(deployerAddr), nil)

	// Setup TxListener
	listener := txlistener.NewTxListener(
		client,
		txlistener.WithPollInterval(2*time.Second),
		txlistener.WithTimeout(5*time.Minute),
	)
	ccm := map[string]ContractClient{
		routerv2:                   swapClient,
		usdc:                       usdcClient,
		wavax:                      wavaxClient,
		black:                      blackClient,
		wavaxUsdcPair:              wausPoolClient,
		nonfungiblePositionManager: nftPositionManagerClient,
		gauge:                      gaugeClient,
		farmingCenter:              farmingCenterClient,
		deployer:                   deployerClient,
	}
	// Create Blackhole instance
	b := &Blackhole{
		privateKey: privateKey,
		myAddr:     address,
		tl:         listener,
		registry:   NewContractRegistry(ccm),
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
		params := &types.SWAPExactTokensForTokensParams{
			AmountIn:     amountInBig,
			AmountOutMin: amountOutMinBig,
			Routes: []types.Route{
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

	// t.Run("GetAmountOut", func(t *testing.T) {

	// 	amount, err := b.GetAmountOut(common.HexToAddress(wavaxUsdcPair), big.NewInt(5023780141629851102), common.HexToAddress(wavax))
	// 	if err != nil {
	// 		t.Fatalf("Failed to call GetAmountOut: %v", err)
	// 	}
	// 	t.Logf("GetAmountOut Result %v", amount)
	// })

	t.Run("Mint", func(t *testing.T) {
		b.poolType = types.CL1
		maxWAVAX := big.NewInt(410565038267351832) //0.390070889896271532
		maxUSDC := big.NewInt(160879)
		rangeWidth := 200
		slippagePct := 5

		rtn, err := b.Mint(maxWAVAX, maxUSDC, rangeWidth, slippagePct)
		if err != nil {
			t.Fatalf("Mint failed: %v", err)
		}

		t.Logf("Mint Result %v", rtn)
	})

	t.Run("Stake", func(t *testing.T) {

		nftId := big.NewInt(2519306)
		rtn, err := b.Stake(nftId)
		if err != nil {
			t.Fatalf("Stake failed: %v", err)
		}

		t.Logf("Stake Result %v", rtn)
	})

	t.Run("Unstake", func(t *testing.T) {

		nftId := big.NewInt(2519306)
		rtn, err := b.Unstake(nftId, big.NewInt(1)) // todo Nonce 구하는 법
		if err != nil {
			t.Fatalf("Stake failed: %v", err)
		}

		t.Logf("Stake Result %v", rtn)
	})

	t.Run("Withdraw", func(t *testing.T) {
		nftId := big.NewInt(2519306)
		rtn, err := b.Withdraw(nftId) // todo Nonce 구하는 법
		if err != nil {
			t.Fatalf("Withdraw failed: %v", err)
		}

		t.Logf("Withdraw Result %v", rtn)
	})

	t.Run("GetAMMState", func(t *testing.T) {
		// if !strings.Contains(env, "IFarmingCenter") {
		// 	t.Fatal("wrong env")
		// }

		state, err := b.GetAMMState()
		if err != nil {
			t.Fatalf("Failed to call GetAMMState: %v", err)
		}
		t.Logf("GetAMMState Result %v", state)
	})
}
