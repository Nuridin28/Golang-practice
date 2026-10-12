// Task 4: Bank account.
package main

import (
	"errors"
	"fmt"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

type BankAccount struct {
	Owner   string
	balance float64
}

// All methods use pointer receivers. Deposit and Withdraw modify the balance;
// Balance only reads, but follows the consistency rule from the Go FAQ.

func (a *BankAccount) Deposit(amount float64) {
	a.balance += amount
}

func (a *BankAccount) Withdraw(amount float64) error {
	if amount > a.balance {
		return fmt.Errorf("withdraw %.2f from %s (balance %.2f): %w",
			amount, a.Owner, a.balance, ErrInsufficientFunds)
	}
	a.balance -= amount
	return nil
}

func (a *BankAccount) Balance() float64 {
	return a.balance
}

func main() {
	accounts := []BankAccount{
		{Owner: "Ali", balance: 100},
		{Owner: "Bota", balance: 200},
		{Owner: "Dias", balance: 300},
	}

	if err := accounts[0].Withdraw(1000); err != nil {
		fmt.Println("error:", err)
	}

	// acc is a copy, so the interest is lost.
	for _, acc := range accounts {
		acc.Deposit(acc.Balance() * 0.05)
	}
	fmt.Println("after range-copy loop:")
	for i := range accounts {
		fmt.Printf("  %s: %.2f\n", accounts[i].Owner, accounts[i].Balance())
	}

	for i := range accounts {
		accounts[i].Deposit(accounts[i].Balance() * 0.05)
	}
	fmt.Println("after index loop (+5%):")
	for i := range accounts {
		fmt.Printf("  %s: %.2f\n", accounts[i].Owner, accounts[i].Balance())
	}
}
