package main

import "math/big"

// Asset is a fungible asset that a wallet can hold.
// Native SOL uses an empty Mint. Future SPL tokens can use the same
// fields with a mint address; they do not need a separate domain type.
type Asset struct {
	Symbol   string
	Mint     string
	Decimals int
}

// NativeSOL is the identity of Solana's native asset.
var NativeSOL = Asset{
	Symbol:   "SOL",
	Mint:     "",
	Decimals: 9,
}

func (a Asset) Equal(other Asset) bool {
	return a.Symbol == other.Symbol && a.Mint == other.Mint
}

func (a Asset) baseUnitDivisor() *big.Rat {
	decimals := a.Decimals
	if decimals < 0 {
		decimals = 0
	}
	denom := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	return new(big.Rat).SetInt(denom)
}
