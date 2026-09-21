package main

import (
	"fmt"
	"strings"

	"github.com/gagliardetto/solana-go"
)

func main() {
	fmt.Println("Solana Wallet Tracker")
	fmt.Println("======================")

	fmt.Print("Enter Solana wallet address: ")

	var wallet string
	fmt.Scanln(&wallet)

	wallet = strings.TrimSpace(wallet)

	pubkey, err := solana.PublicKeyFromBase58(wallet)
	if err != nil {
		fmt.Println("Invalid Solana wallet address.")
		return
	}

	fmt.Println("Wallet address accepted:", pubkey.String())
}
