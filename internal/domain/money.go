package domain

import (
	"errors"
	"strconv"
	"strings"
)

type Money struct {
	amount   int64
	currency string
}

func NewMoneyFromString(amountStr string, currency string) (Money, error) {
	if amountStr == "" || currency == "" {
		return Money{}, errors.New("amount or currency cannot be blank")
	}

	if strings.HasPrefix(amountStr, "-") {
		return Money{}, errors.New("invalid value cannot be negative")
	}

	parts := strings.Split(amountStr, ".")

	if len(parts) != 2 {
		return Money{}, errors.New("the value must contain a decimal point")
	}

	if len(parts[1]) != 2 {
		return Money{}, errors.New("the value must have exactly two decimal places")
	}

	cleanStr := parts[0] + parts[1]

	amount, err := strconv.ParseInt(cleanStr, 10, 64)
	if err != nil {
		return Money{}, errors.New("invalid number format")
	}

	return Money{
		amount:   amount,
		currency: currency,
	}, nil
}

func (m Money) Add(other Money) (Money, error) {

	if m.currency != other.currency {
		return Money{}, errors.New("currency mismatch")
	}

	amountSum := m.amount + other.amount

	return Money{
		amount:   amountSum,
		currency: m.currency,
	}, nil
}

func (m Money) Subtract(other Money) (Money, error) {

	if m.currency != other.currency {
		return Money{}, errors.New("currency mismatch")
	}

	amountSub := m.amount - other.amount

	return Money{
		amount:   amountSub,
		currency: m.currency,
	}, nil
}
