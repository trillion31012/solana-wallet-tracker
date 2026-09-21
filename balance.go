package main

import "math/big"

// Balance is how many base units of an Asset a wallet holds.
// Amount is exact (lamports for native SOL) and is never stored as float64.
type Balance struct {
	Asset  Asset
	Amount uint64
}

func (b Balance) Quantity() *big.Rat {
	quantity := new(big.Rat).SetUint64(b.Amount)
	return quantity.Quo(quantity, b.Asset.baseUnitDivisor())
}

func (b Balance) FiatValue(pricePerWhole *big.Rat) *big.Rat {
	return new(big.Rat).Mul(b.Quantity(), pricePerWhole)
}
