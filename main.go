package main

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
)

func main() {
	fmt.Println("Solana Wallet Tracker")
	fmt.Println("======================")

	fmt.Print("Enter Solana wallet address: ")

	var wallet string
	fmt.Scanln(&wallet)

	pubkey, err := solana.PublicKeyFromBase58(wallet)
	if err != nil {
		fmt.Println("Invalid Solana address.")
		return
	}

	fmt.Println("Valid address:", pubkey.String())
}
