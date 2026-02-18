package domain

import (
	"math/big"
)

// Money value object for precise decimal arithmetic
type Money struct {
	amount *big.Rat
}

func NewMoneyFromRat(amount *big.Rat) *Money {
	if amount == nil {
		return nil
	}
	// Create a copy to prevent external modification
	amountCopy := new(big.Rat).Set(amount)
	return &Money{amount: amountCopy}
}

func NewMoneyFromInt64(numerator, denominator int64) *Money {
	return &Money{
		amount: big.NewRat(numerator, denominator),
	}
}

func (m *Money) Amount() *big.Rat {
	// Return a copy to maintain immutability
	return new(big.Rat).Set(m.amount)
}

func (m *Money) Numerator() int64 {
	return m.amount.Num().Int64()
}

func (m *Money) Denominator() int64 {
	return m.amount.Denom().Int64()
}

func (m *Money) Add(other *Money) *Money {
	sum := new(big.Rat).Add(m.amount, other.amount)
	return NewMoneyFromRat(sum)
}

func (m *Money) Sub(other *Money) *Money {
	diff := new(big.Rat).Sub(m.amount, other.amount)
	return NewMoneyFromRat(diff)
}

func (m *Money) Mul(percentage *big.Rat) *Money {
	result := new(big.Rat).Mul(m.amount, percentage)
	return NewMoneyFromRat(result)
}

func (m *Money) Equals(other *Money) bool {
	if other == nil {
		return false
	}
	return m.amount.Cmp(other.amount) == 0
}

func (m *Money) String() string {
	return m.amount.FloatString(2)
}