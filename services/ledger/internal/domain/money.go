package domain

import "fmt"

type Money struct {
	AmountMinor int64
	Currency    string
}

func NewMoney(amountMinor int64, currency string) (Money, error) {
	if amountMinor <= 0 {
		return Money{}, fmt.Errorf("amount_minor must be positive")
	}
	if len(currency) != 3 {
		return Money{}, fmt.Errorf("currency must be 3 chars")
	}
	return Money{AmountMinor: amountMinor, Currency: currency}, nil
}
