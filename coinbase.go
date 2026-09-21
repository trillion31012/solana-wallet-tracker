package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"
)

const (
	coinbaseSpotURLFmt = "https://api.coinbase.com/v2/prices/%s-%s/spot"
	priceHTTPTimeout   = 10 * time.Second
)

var coinbaseHTTPClient = &http.Client{Timeout: priceHTTPTimeout}

// CoinbasePriceProvider loads live SOL/fiat prices from Coinbase.
type CoinbasePriceProvider struct{}

// newPriceProvider is the only place the app chooses which PriceProvider to use.
func newPriceProvider() PriceProvider {
	return CoinbasePriceProvider{}
}

type coinbaseSpotPriceResponse struct {
	Data coinbaseSpotPrice `json:"data"`
}

type coinbaseSpotPrice struct {
	Amount   string `json:"amount"`
	Base     string `json:"base"`
	Currency string `json:"currency"`
}

type coinbaseErrorResponse struct {
	Errors []coinbaseAPIError `json:"errors"`
}

type coinbaseAPIError struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func (p CoinbasePriceProvider) GetSOLPrice(currencyCode string) (*big.Rat, error) {
	return p.GetAssetPrice(NativeSOL, currencyCode)
}

func (p CoinbasePriceProvider) GetAssetPrice(asset Asset, currencyCode string) (*big.Rat, error) {
	code := strings.ToUpper(strings.TrimSpace(currencyCode))
	if code == "" {
		return nil, fmt.Errorf("currency code is required")
	}

	baseSymbol := asset.Symbol
	if baseSymbol == "" {
		baseSymbol = NativeSOL.Symbol
	}
	requestURL := fmt.Sprintf(coinbaseSpotURLFmt, baseSymbol, code)
	request, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create price request: %w", err)
	}
	request.Header.Set("Accept", "application/json")

	response, err := coinbaseHTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch asset price: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read price response: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("%w: %s/%s", ErrAssetUnsupported, baseSymbol, code)
		}
		return nil, p.httpError(code, response.StatusCode, body)
	}

	var result coinbaseSpotPriceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode price response: %w", err)
	}

	if result.Data.Amount == "" || result.Data.Base == "" || result.Data.Currency == "" {
		return nil, fmt.Errorf("invalid price response: missing amount, base, or currency")
	}
	if !strings.EqualFold(result.Data.Base, baseSymbol) {
		return nil, fmt.Errorf("invalid price response: expected base %s, got %s", baseSymbol, result.Data.Base)
	}
	if !strings.EqualFold(result.Data.Currency, code) {
		return nil, fmt.Errorf("invalid price response: expected currency %s, got %s", code, result.Data.Currency)
	}

	price := new(big.Rat)
	if _, ok := price.SetString(result.Data.Amount); !ok {
		return nil, fmt.Errorf("invalid price response: could not parse amount %q", result.Data.Amount)
	}
	if price.Sign() <= 0 {
		return nil, fmt.Errorf("invalid price response: non-positive amount %s", result.Data.Amount)
	}

	return price, nil
}

func (p CoinbasePriceProvider) httpError(currencyCode string, statusCode int, body []byte) error {
	var apiError coinbaseErrorResponse
	if err := json.Unmarshal(body, &apiError); err == nil && len(apiError.Errors) > 0 {
		message := apiError.Errors[0].Message
		if p.isUnsupportedCurrencyError(statusCode, apiError.Errors[0]) {
			return fmt.Errorf("unsupported currency %s: %s", currencyCode, message)
		}
		return fmt.Errorf("price API HTTP %d: %s", statusCode, message)
	}

	if statusCode == http.StatusNotFound {
		return fmt.Errorf("unsupported currency %s", currencyCode)
	}

	return fmt.Errorf("price API returned HTTP status %d", statusCode)
}

func (CoinbasePriceProvider) isUnsupportedCurrencyError(statusCode int, apiError coinbaseAPIError) bool {
	if statusCode == http.StatusNotFound {
		return true
	}

	id := strings.ToLower(apiError.ID)
	message := strings.ToLower(apiError.Message)
	return id == "not_found" || (strings.Contains(message, "invalid") && strings.Contains(message, "currency"))
}
