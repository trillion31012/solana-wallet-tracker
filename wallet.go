package main

import "github.com/gagliardetto/solana-go"

// Wallet is a Solana wallet identified by its public address.
type Wallet struct {
	Address string
}

func NewSolanaWallet(address string) (Wallet, error) {
	pubkey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return Wallet{}, err
	}
	return Wallet{Address: pubkey.String()}, nil
}
