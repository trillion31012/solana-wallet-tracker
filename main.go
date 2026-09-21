package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gagliardetto/solana-go"
)

const solanaRPCURL = "https://api.mainnet-beta.solana.com"

type rpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type rpcResponse struct {
	Result *struct {
		Value uint64 `json:"value"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func getBalance(wallet string) (uint64, error) {
	requestBody := rpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "getBalance",
		Params:  []interface{}{wallet},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return 0, fmt.Errorf("encode RPC request: %w", err)
	}

	response, err := http.Post(solanaRPCURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return 0, fmt.Errorf("send RPC request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("RPC request returned HTTP status %s", response.Status)
	}

	var result rpcResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode RPC response: %w", err)
	}

	if result.Error != nil {
		return 0, fmt.Errorf("RPC error %d: %s", result.Error.Code, result.Error.Message)
	}
	if result.Result == nil {
		return 0, fmt.Errorf("RPC response did not contain a balance")
	}

	return result.Result.Value, nil
}

func main() {
	fmt.Println("Solana Wallet Tracker")
	fmt.Println("======================")

	fmt.Print("Enter Solana wallet address: ")

	var wallet string
	fmt.Scanln(&wallet)

	wallet = strings.TrimSpace(wallet)

	pubkey, err := solana.PublicKeyFromBase58(wallet)
	if err != nil {
		fmt.Println("Invalid Solana wallet address.")
		return
	}

	fmt.Println("Wallet address accepted:", pubkey.String())

	lamports, err := getBalance(pubkey.String())
	if err != nil {
		fmt.Println("Could not retrieve wallet balance:", err)
		return
	}

	const lamportsPerSOL = 1_000_000_000
	solBalance := float64(lamports) / lamportsPerSOL
	fmt.Printf("SOL balance: %.9f SOL\n", solBalance)
}
