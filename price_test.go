package main

import (
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

func TestCoinbasePriceProviderSatisfiesPriceProvider(t *testing.T) {
	// This assignment only compiles if CoinbasePriceProvider implements PriceProvider.
	var provider PriceProvider = CoinbasePriceProvider{}
	if _, ok := provider.(CoinbasePriceProvider); !ok {
		t.Fatal("expected CoinbasePriceProvider to be used as a PriceProvider")
	}
}

func TestFormatFiatUsesMinorUnits(t *testing.T) {
	value := big.NewRat(123456, 1000) // 123.456
	if got := formatFiat(value, 2); got != "123.46" {
		t.Fatalf("formatFiat(..., 2) = %q, want %q", got, "123.46")
	}
}
