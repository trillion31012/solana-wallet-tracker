package main

import (
	"math/big"
	"testing"
)

const testWalletAddress = "11111111111111111111111111111111"

func TestWalletCanRepresentAValidSolanaAddress(t *testing.T) {
	wallet, err := NewSolanaWallet(testWalletAddress)
	if err != nil {
		t.Fatalf("NewSolanaWallet() unexpected error: %v", err)
	}
	if wallet.Address != testWalletAddress {
		t.Fatalf("Wallet.Address = %q, want %q", wallet.Address, testWalletAddress)
	}
}

func TestNewSolanaWalletRejectsInvalidAddress(t *testing.T) {
	_, err := NewSolanaWallet("not-a-solana-address")
	if err == nil {
		t.Fatal("expected an error for an invalid Solana address")
	}
}

func TestAssetCanRepresentSOL(t *testing.T) {
	if NativeSOL.Symbol != "SOL" {
		t.Fatalf("NativeSOL.Symbol = %q, want %q", NativeSOL.Symbol, "SOL")
	}
	if NativeSOL.Mint != "" {
		t.Fatalf("NativeSOL.Mint = %q, want empty mint for native SOL", NativeSOL.Mint)
	}
	if NativeSOL.Decimals != 9 {
		t.Fatalf("NativeSOL.Decimals = %d, want 9", NativeSOL.Decimals)
	}
	if !NativeSOL.Equal(Asset{Symbol: "SOL", Mint: ""}) {
		t.Fatal("NativeSOL should equal another SOL asset with an empty mint")
	}
}

func TestBalanceAssociatesSOLWithAnExactAmount(t *testing.T) {
	const lamports uint64 = 1_500_000_000
	balance := Balance{Asset: NativeSOL, Amount: lamports}

	if !balance.Asset.Equal(NativeSOL) {
		t.Fatal("Balance.Asset should be NativeSOL")
	}
	if balance.Amount != lamports {
		t.Fatalf("Balance.Amount = %d, want %d", balance.Amount, lamports)
	}

	quantity := balance.Quantity()
	wantQuantity := big.NewRat(3, 2)
	if quantity.Cmp(wantQuantity) != 0 {
		t.Fatalf("Balance.Quantity() = %s, want %s", quantity.RatString(), wantQuantity.RatString())
	}
}

func TestPortfolioContainsTheWalletSOLBalance(t *testing.T) {
	wallet, err := NewSolanaWallet(testWalletAddress)
	if err != nil {
		t.Fatalf("NewSolanaWallet() unexpected error: %v", err)
	}

	const lamports uint64 = 2_000_000_000
	portfolio := Portfolio{
		Wallet:   wallet,
		Balances: []Balance{{Asset: NativeSOL, Amount: lamports}},
	}

	if portfolio.Wallet.Address != testWalletAddress {
		t.Fatalf("Portfolio.Wallet.Address = %q, want %q", portfolio.Wallet.Address, testWalletAddress)
	}

	solBalance, ok := portfolio.BalanceOf(NativeSOL)
	if !ok {
		t.Fatal("portfolio should contain a NativeSOL balance")
	}
	if solBalance.Amount != lamports {
		t.Fatalf("SOL Balance.Amount = %d, want %d", solBalance.Amount, lamports)
	}
}

func TestPortfolioFiatValueMatchesExistingSOLValuation(t *testing.T) {
	price := new(big.Rat)
	if _, ok := price.SetString("162.101796551575652452"); !ok {
		t.Fatal("failed to parse price")
	}

	const lamports uint64 = 1_500_000_000
	portfolio := Portfolio{
		Wallet:   Wallet{Address: testWalletAddress},
		Balances: []Balance{{Asset: NativeSOL, Amount: lamports}},
	}

	got := portfolio.FiatValue(price)
	want := calculateValue(lamports, price)
	if got.Cmp(want) != 0 {
		t.Fatalf("Portfolio.FiatValue() = %s, want %s", got.RatString(), want.RatString())
	}
}
