package main

import (
	"errors"
	"math/big"
)

// Portfolio is the set of balances held by a wallet.
// It is domain state only: it does not call RPC or price APIs.
type Portfolio struct {
	Wallet   Wallet
	Balances []Balance
}

type AssetValuation struct {
	Asset  Asset
	Price  *big.Rat
	Value  *big.Rat
	Amount *big.Rat
	Err    error
}

type PortfolioValuation struct {
	Assets     []AssetValuation
	Total      *big.Rat
	Incomplete bool
}

func (p Portfolio) BalanceOf(asset Asset) (Balance, bool) {
	for _, balance := range p.Balances {
		if balance.Asset.Equal(asset) {
			return balance, true
		}
	}
	return Balance{}, false
}

func (p Portfolio) FiatValue(pricePerWhole *big.Rat) *big.Rat {
	total := new(big.Rat)
	for _, balance := range p.Balances {
		total.Add(total, balance.FiatValue(pricePerWhole))
	}
	return total
}

func (p Portfolio) Evaluate(provider AssetPriceProvider, currencyCode string) PortfolioValuation {
	valuations := make([]AssetValuation, 0, len(p.Balances))
	total := new(big.Rat)
	incomplete := false

	for _, balance := range p.Balances {
		valuation := AssetValuation{
			Asset:  balance.Asset,
			Amount: balance.Quantity(),
			Price:  nil,
			Value:  nil,
		}

		price, err := getAssetPrice(provider, balance.Asset, currencyCode)
		if err != nil {
			if errors.Is(err, ErrAssetUnsupported) {
				valuation.Err = ErrAssetUnsupported
				incomplete = true
				valuations = append(valuations, valuation)
				continue
			}
			valuation.Err = err
			incomplete = true
			valuations = append(valuations, valuation)
			continue
		}

		valuation.Price = price
		valuation.Value = new(big.Rat).Mul(balance.Quantity(), price)
		total.Add(total, valuation.Value)
		valuations = append(valuations, valuation)
	}

	return PortfolioValuation{Assets: valuations, Total: total, Incomplete: incomplete}
}
