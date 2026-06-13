package money

import (
	"fmt"

	kiterrors "github.com/mesob-wallet/go-kit/errors"
)

type Money struct {
	AmountMinor int64
	Currency    string
}

func New(amountMinor int64, currency string) (Money, error) {
	if amountMinor < 0 {
		return Money{}, fmt.Errorf("amount_minor must be non-negative")
	}
	if len(currency) != 3 {
		return Money{}, fmt.Errorf("currency must be ISO 4217 3-char code")
	}
	return Money{AmountMinor: amountMinor, Currency: currency}, nil
}

func MustNew(amountMinor int64, currency string) Money {
	m, err := New(amountMinor, currency)
	if err != nil {
		panic(err)
	}
	return m
}

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("currency mismatch: %s vs %s", m.Currency, other.Currency)
	}
	return Money{AmountMinor: m.AmountMinor + other.AmountMinor, Currency: m.Currency}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("currency mismatch: %s vs %s", m.Currency, other.Currency)
	}
	result := m.AmountMinor - other.AmountMinor
	if result < 0 {
		return Money{}, kiterrors.ErrInsufficientBalance
	}
	return Money{AmountMinor: result, Currency: m.Currency}, nil
}

func (m Money) IsZero() bool { return m.AmountMinor == 0 }
func (m Money) String() string {
	return fmt.Sprintf("%d %s", m.AmountMinor, m.Currency)
}
