package main

import (
	"errors"
	"math/big"
	"testing"
)

func TestTokenBalanceQuantityPreservesExactDecimals(t *testing.T) {
	asset := Asset{Symbol: "USDC", Mint: "EPjFWdd5AufqSSqeM\u003d?", Decimals: 6}
	balance := Balance{Asset: asset, Amount: 1234567}

	want := new(big.Rat)
	if _, ok := want.SetString("1.234567"); !ok {
		t.Fatal("failed to parse expected quantity")
	}
	if got := balance.Quantity(); got.Cmp(want) != 0 {
		t.Fatalf("Balance.Quantity() = %s, want %s", got.RatString(), want.RatString())
	}
}

func TestPortfolioCanRepresentMultipleAssets(t *testing.T) {
	portfolio := Portfolio{Wallet: Wallet{Address: testWalletAddress}, Balances: []Balance{
		{Asset: NativeSOL, Amount: 2_000_000_000},
		{Asset: Asset{Symbol: "USDC", Mint: "EPjFWdd5AufqSSqeM\u003d?", Decimals: 6}, Amount: 250_000_000},
		{Asset: Asset{Symbol: "TOKENA", Mint: "TokenA1234567890", Decimals: 9}, Amount: 42_000_000_000},
	}}

	if len(portfolio.Balances) != 3 {
		t.Fatalf("len(portfolio.Balances) = %d, want 3", len(portfolio.Balances))
	}
	if _, ok := portfolio.BalanceOf(Asset{Symbol: "USDC", Mint: "EPjFWdd5AufqSSqeM\u003d?"}); !ok {
		t.Fatal("portfolio should contain USDC balance")
	}
}

func TestPortfolioValuationHandlesUnpricedAssets(t *testing.T) {
	provider := stubAssetPriceProvider{prices: map[string]*big.Rat{
		"SOL":  big.NewRat(100, 1),
		"USDC": big.NewRat(1, 1),
	}}
	portfolio := Portfolio{Balances: []Balance{
		{Asset: NativeSOL, Amount: 1_500_000_000},
		{Asset: Asset{Symbol: "USDC", Mint: "EPjFWdd5AufqSSqeM\u003d?", Decimals: 6}, Amount: 250_000_000},
		{Asset: Asset{Symbol: "UNKNOWN", Mint: "unknownMint", Decimals: 9}, Amount: 5000},
	}}

	valuation := portfolio.Evaluate(provider, "AUD")
	if valuation.Incomplete != true {
		t.Fatal("expected valuation to be marked incomplete when an asset has no price")
	}
	if len(valuation.Assets) != 3 {
		t.Fatalf("len(valuation.Assets) = %d, want 3", len(valuation.Assets))
	}
	if valuation.Assets[2].Err == nil {
		t.Fatal("expected unknown asset to carry an error status")
	}
	if valuation.Total == nil || valuation.Total.Cmp(big.NewRat(400, 1)) != 0 {
		t.Fatalf("valuation.Total = %s, want %s", valuation.Total.RatString(), big.NewRat(400, 1).RatString())
	}
}

func TestAssetProviderReturnsPriceForSupportedAsset(t *testing.T) {
	provider := stubAssetPriceProvider{prices: map[string]*big.Rat{"USDC": big.NewRat(2, 1)}}
	price, err := getAssetPrice(provider, Asset{Symbol: "USDC", Mint: "EPjFWdd5AufqSSqeM\u003d?", Decimals: 6}, "AUD")
	if err != nil {
		t.Fatalf("getAssetPrice() unexpected error: %v", err)
	}
	if price.Cmp(big.NewRat(2, 1)) != 0 {
		t.Fatalf("getAssetPrice() = %s, want %s", price.RatString(), big.NewRat(2, 1).RatString())
	}
}

func TestAssetProviderReturnsUnsupportedAssetError(t *testing.T) {
	provider := stubAssetPriceProvider{prices: map[string]*big.Rat{}}
	_, err := getAssetPrice(provider, Asset{Symbol: "UNKNOWN", Mint: "unknownMint", Decimals: 9}, "AUD")
	if !errors.Is(err, ErrAssetUnsupported) {
		t.Fatalf("getAssetPrice() error = %v, want %v", err, ErrAssetUnsupported)
	}
}

func TestAssetProviderPropagatesProviderFailure(t *testing.T) {
	wantErr := errors.New("provider network failure")
	provider := stubAssetPriceProvider{err: wantErr}
	_, err := getAssetPrice(provider, NativeSOL, "AUD")
	if !errors.Is(err, wantErr) {
		t.Fatalf("getAssetPrice() error = %v, want %v", err, wantErr)
	}
}

type stubAssetPriceProvider struct {
	prices map[string]*big.Rat
	err    error
}

func (s stubAssetPriceProvider) GetAssetPrice(asset Asset, currencyCode string) (*big.Rat, error) {
	if s.err != nil {
		return nil, s.err
	}
	if asset.Symbol == "" {
		return nil, ErrAssetUnsupported
	}
	if price, ok := s.prices[asset.Symbol]; ok {
		return price, nil
	}
	return nil, ErrAssetUnsupported
}

func (s stubAssetPriceProvider) GetSOLPrice(currencyCode string) (*big.Rat, error) {
	return s.GetAssetPrice(NativeSOL, currencyCode)
}
