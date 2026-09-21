package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/JohannesJHN/iso4217"
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

func loadCurrencies() map[string]iso4217.Currency {
	return iso4217.AllActive()
}

func readLine(reader *bufio.Reader) (string, bool) {
	line, err := reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", false
	}

	return strings.TrimSpace(line), true
}

func getCurrency(reader *bufio.Reader, currencies map[string]iso4217.Currency) (iso4217.Currency, bool) {
	for {
		fmt.Print("Enter currency ticker (e.g. AUD, USD, CAD): ")
		ticker, ok := readLine(reader)
		if !ok {
			return iso4217.Currency{}, false
		}

		currency, exists := currencies[strings.ToUpper(ticker)]
		if exists {
			return currency, true
		}

		fmt.Println("Unknown currency. Please enter a valid ISO-4217 currency code.")
	}
}

func getWalletAddress(reader *bufio.Reader) (string, bool) {
	for {
		fmt.Print("Enter Solana wallet address: ")
		wallet, ok := readLine(reader)
		if !ok {
			return "", false
		}

		pubkey, err := solana.PublicKeyFromBase58(wallet)
		if err == nil {
			return pubkey.String(), true
		}

		fmt.Println("Invalid Solana wallet address. Please try again.")
	}
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

func shouldRetry(reader *bufio.Reader) bool {
	for {
		fmt.Print("Try again? (y/n): ")
		answer, ok := readLine(reader)
		if !ok {
			return false
		}

		switch strings.ToLower(answer) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("Please enter y or n.")
		}
	}
}

func displayResult(currency iso4217.Currency, wallet string, lamports uint64, solPrice *big.Rat, fiatValue *big.Rat) {
	solBalance := new(big.Rat).SetUint64(lamports)
	solBalance.Quo(solBalance, new(big.Rat).SetUint64(lamportsPerSOL))

	fmt.Println("Wallet address accepted:", wallet)
	fmt.Println("Selected currency:", currency.Alpha3)
	fmt.Printf("SOL balance: %s SOL\n", solBalance.FloatString(9))
	fmt.Printf("SOL price: %s %s\n", formatFiat(solPrice, currency.MinorUnits), currency.Alpha3)
	fmt.Printf("Estimated value: %s %s\n", formatFiat(fiatValue, currency.MinorUnits), currency.Alpha3)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Solana Wallet Tracker")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("Loading supported currencies...")

	var prices PriceProvider = CoinbasePriceProvider{}
	currencies := loadCurrencies()
	currency, ok := getCurrency(reader, currencies)
	if !ok {
		return
	}

	wallet, ok := getWalletAddress(reader)
	if !ok {
		return
	}

	for {
		lamports, err := getBalance(wallet)
		if err != nil {
			fmt.Println("Could not retrieve wallet balance:", err)
			if !shouldRetry(reader) {
				return
			}
			continue
		}

		solPrice, err := getSOLPrice(prices, currency.Alpha3)
		if err != nil {
			fmt.Println("Could not retrieve SOL price:", err)
			if !shouldRetry(reader) {
				return
			}
			continue
		}

		fiatValue := calculateValue(lamports, solPrice)
		displayResult(currency, wallet, lamports, solPrice, fiatValue)
		break
	}

	fmt.Println()
	fmt.Print("Press Enter to exit.")
	readLine(reader)
}
