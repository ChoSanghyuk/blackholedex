package contractclient

import (
	"blackholego/internal/util"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func TestDecodeTransaction(t *testing.T) {
	// Load .env.test.local file
	err := godotenv.Load("env/.env.test.local")
	if err != nil {
		t.Fatalf("Failed to load .env.test.local: %v", err)
	}

	// Get configuration from env
	contractAddr := os.Getenv("CONTRACT_ADDR")
	if contractAddr == "" {
		t.Fatal("CONTRACT_ADDR not set in .env.test.local")
	}

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		t.Fatal("RPC_URL not set in .env.test.local")
	}

	txHash := os.Getenv("TX_HASH")
	txData := os.Getenv("TX_DATA")
	if txHash == "" && txData == "" {
		t.Fatal("Either TX_HASH or TX_DATA not set in .env.test.local")
	}

	path := os.Getenv("ABI_PATH")
	if path == "" {
		t.Fatal("ABI_PATH not set in .env.test.local")
	}

	t.Logf("Loaded test config - Contract: %s, RPC: %s, TxHash: %s, TxData: %s\n", contractAddr, rpcURL, txHash, txData)

	abi, err := util.LoadABIFromHardhatArtifact(path)
	if err != nil {
		t.Fatal(err)
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		t.Fatal(err)
	}
	cc := NewContractClient(client, common.HexToAddress(contractAddr), abi)

	t.Run("decode_tx", func(t *testing.T) {

		var txDataBytes []byte
		if txData != "" {
			txDataBytes = util.Hex2Bytes(txData)
		} else {
			txDataBytes, err = cc.TransactionData(common.HexToHash(txHash))
		}
		// Decode it
		decoded, err := cc.DecodeTransaction(txDataBytes)
		if err != nil {
			t.Fatal(err)
		}

		// data is your struct instance
		jsonData, err := json.MarshalIndent(decoded, "", "  ")
		if err != nil {
			// Handle error
			fmt.Println("Error marshalling to JSON:", err)
			return
		}

		// Print the JSON byte slice as a string
		t.Logf("Decoded transaction:\n%s", string(jsonData))
	})

	// t.Run("decode_hex_string", func(t *testing.T) {

	// 	codec := NewContractClient(nil, common.HexToAddress(contractAddr), &abi)

	// 	// transfer(address,uint256) - with 0x prefix
	// 	hexData := "0xa9059cbb0000000000000000000000006e4141d33021b52c91c28608403db4a0ffb50ec600000000000000000000000000000000000000000000000000000000000f4240"

	// 	decoded, err := codec.DecodeTransactionHex(hexData)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	if decoded.MethodName != "transfer" {
	// 		t.Errorf("expected method name 'transfer', got '%s'", decoded.MethodName)
	// 	}

	// 	decodedJSON, err := json.MarshalIndent(decoded, "", "  ")
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	t.Logf("Decoded from hex:\n%s", string(decodedJSON))
	// })

}

func TestCallTransaction(t *testing.T) {
	// Load .env.test.local file
	err := godotenv.Load("env/.env.globalstate.local")
	if err != nil {
		t.Fatalf("Failed to load .env.test.local: %v", err)
	}

	// Get configuration from env
	contractAddr := os.Getenv("CONTRACT_ADDR")
	if contractAddr == "" {
		t.Fatal("CONTRACT_ADDR not set in .env.test.local")
	}

	// callerAddr := os.Getenv("CALLER_ADDR")
	// if contractAddr == "" {
	// 	t.Fatal("CALLER_ADDR not set in .env.test.local")
	// }

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		t.Fatal("RPC_URL not set in .env.test.local")
	}

	path := os.Getenv("ABI_PATH")
	if path == "" {
		t.Fatal("ABI_PATH not set in .env.test.local")
	}

	t.Logf("Loaded test config - Contract: %s, RPC: %s\n", contractAddr, rpcURL)

	abi, err := util.LoadABIFromHardhatArtifact(path)
	if err != nil {
		t.Fatal(err)
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		t.Fatal(err)
	}
	cc := NewContractClient(client, common.HexToAddress(contractAddr), abi)

	t.Run("safelyGetStateOfAMM", func(t *testing.T) {

		// temp := common.HexToAddress(callerAddr)
		outputs, err := cc.Call(nil, "safelyGetStateOfAMM")
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("safelyGetStateOfAMM outputs:%v", outputs)
		// [304014154377809408260091 -249428 500 2 1514349024952878554 -249398 -249433]
	})

	t.Run("tickSpacing", func(t *testing.T) {

		// temp := common.HexToAddress(callerAddr)
		outputs, err := cc.Call(nil, "tickSpacing")
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("safelyGetStateOfAMM outputs:%v", outputs)
		// [304014154377809408260091 -249428 500 2 1514349024952878554 -249398 -249433]
	})

}
