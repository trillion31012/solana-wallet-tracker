package main

import "math/big"

const lamportsPerSOL = 1_000_000_000

// PriceProvider looks up the current SOL price in a fiat currency.
// The rest of the app depends on this interface, not on a specific exchange.
type PriceProvider interface {
	GetSOLPrice(currencyCode string) (*big.Rat, error)
}

func getSOLPrice(provider PriceProvider, currencyCode string) (*big.Rat, error) {
	return provider.GetSOLPrice(currencyCode)
}

func calculateValue(lamports uint64, solPrice *big.Rat) *big.Rat {
	value := new(big.Rat).SetUint64(lamports)
	value.Mul(value, solPrice)
	value.Quo(value, new(big.Rat).SetUint64(lamportsPerSOL))
	return value
}

func formatFiat(value *big.Rat, minorUnits int) string {
	digits := minorUnits
	if digits < 0 {
		digits = 2
	}
	return value.FloatString(digits)
}
