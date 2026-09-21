package main

import (
	"errors"
	"fmt"
	"math/big"
)

var ErrAssetUnsupported = errors.New("asset is unsupported by the price provider")

// PriceProvider looks up the current SOL price in a fiat currency.
// The rest of the app depends on this interface, not on a specific exchange.
type PriceProvider interface {
	GetSOLPrice(currencyCode string) (*big.Rat, error)
}

// AssetPriceProvider looks up the current price for a particular asset in a fiat currency.
type AssetPriceProvider interface {
	GetAssetPrice(asset Asset, currencyCode string) (*big.Rat, error)
}

func getSOLPrice(provider PriceProvider, currencyCode string) (*big.Rat, error) {
	return provider.GetSOLPrice(currencyCode)
}

func getAssetPrice(provider AssetPriceProvider, asset Asset, currencyCode string) (*big.Rat, error) {
	if provider == nil {
		return nil, fmt.Errorf("nil price provider")
	}
	price, err := provider.GetAssetPrice(asset, currencyCode)
	if err != nil {
		if errors.Is(err, ErrAssetUnsupported) {
			return nil, ErrAssetUnsupported
		}
		return nil, err
	}
	return price, nil
}

func calculateValue(lamports uint64, solPrice *big.Rat) *big.Rat {
	return Balance{Asset: NativeSOL, Amount: lamports}.FiatValue(solPrice)
}

func formatFiat(value *big.Rat, minorUnits int) string {
	digits := minorUnits
	if digits < 0 {
		digits = 2
	}
	return value.FloatString(digits)
}
