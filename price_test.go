package main

import (
	"errors"
	"math/big"
	"testing"
)

func TestCalculateValueUsesExactRationalArithmetic(t *testing.T) {
	price := new(big.Rat)
	if _, ok := price.SetString("162.101796551575652452"); !ok {
		t.Fatal("failed to parse price")
	}

	const lamports uint64 = 1_500_000_000 // 1.5 SOL
	got := calculateValue(lamports, price)
	want := new(big.Rat)
	want.SetString("243.152694827363478678")

	if got.Cmp(want) != 0 {
		t.Fatalf("calculateValue() = %s, want %s", got.RatString(), want.RatString())
	}
}

func TestFormatFiatUsesMinorUnits(t *testing.T) {
	value := big.NewRat(123456, 1000) // 123.456
	if got := formatFiat(value, 2); got != "123.46" {
		t.Fatalf("formatFiat(..., 2) = %q, want %q", got, "123.46")
	}
}

func TestCoinbasePriceProviderSatisfiesPriceProvider(t *testing.T) {
	var provider PriceProvider = CoinbasePriceProvider{}
	if _, ok := provider.(CoinbasePriceProvider); !ok {
		t.Fatal("expected CoinbasePriceProvider to satisfy PriceProvider")
	}
}

func TestNewPriceProviderReturnsPriceProvider(t *testing.T) {
	provider := newPriceProvider()
	if provider == nil {
		t.Fatal("newPriceProvider() returned nil")
	}
}

func TestGetSOLPriceUsesTheProvidedPriceProvider(t *testing.T) {
	want := big.NewRat(162, 1)
	provider := stubPriceProvider{price: want}

	got, err := getSOLPrice(provider, "AUD")
	if err != nil {
		t.Fatalf("getSOLPrice() unexpected error: %v", err)
	}
	if got.Cmp(want) != 0 {
		t.Fatalf("getSOLPrice() = %s, want %s", got.RatString(), want.RatString())
	}
}

func TestGetSOLPriceReturnsProviderError(t *testing.T) {
	wantErr := errors.New("price lookup failed")
	provider := stubPriceProvider{err: wantErr}

	_, err := getSOLPrice(provider, "USD")
	if !errors.Is(err, wantErr) {
		t.Fatalf("getSOLPrice() error = %v, want %v", err, wantErr)
	}
}

func TestCoinbaseGetSOLPriceRejectsEmptyCurrency(t *testing.T) {
	provider := CoinbasePriceProvider{}
	_, err := provider.GetSOLPrice("  ")
	if err == nil {
		t.Fatal("expected an error for an empty currency code")
	}
}

type stubPriceProvider struct {
	price *big.Rat
	err   error
}

func (s stubPriceProvider) GetSOLPrice(string) (*big.Rat, error) {
	return s.price, s.err
}
