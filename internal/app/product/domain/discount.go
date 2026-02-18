package domain

import (
	"math/big"
	"time"
)

// Discount value object
type Discount struct {
	percentage *big.Rat
	startDate  time.Time
	endDate    time.Time
}

func NewDiscount(percentage *big.Rat, startDate, endDate time.Time) (*Discount, error) {
	if percentage == nil {
		return nil, ErrInvalidDiscountPeriod
	}

	// Validate percentage is between 0 and 100
	if percentage.Cmp(big.NewRat(0, 1)) < 0 || percentage.Cmp(big.NewRat(100, 1)) > 0 {
		return nil, ErrDiscountPercentage
	}

	// Validate dates
	if startDate.After(endDate) || startDate.Equal(endDate) {
		return nil, ErrInvalidDiscountPeriod
	}

	return &Discount{
		percentage: new(big.Rat).Set(percentage),
		startDate:  startDate,
		endDate:    endDate,
	}, nil
}

func (d *Discount) Percentage() *big.Rat {
	return new(big.Rat).Set(d.percentage)
}

func (d *Discount) StartDate() time.Time {
	return d.startDate
}

func (d *Discount) EndDate() time.Time {
	return d.endDate
}

func (d *Discount) IsValidAt(t time.Time) bool {
	return (t.Equal(d.startDate) || t.After(d.startDate)) && 
		   (t.Before(d.endDate) || t.Equal(d.endDate))
}

func (d *Discount) CalculateDiscountedPrice(originalPrice *Money) *Money {
	if originalPrice == nil {
		return nil
	}

	// discountAmount = originalPrice * (percentage/100)
	discountFactor := new(big.Rat).Quo(d.percentage, big.NewRat(100, 1))
	discountAmount := originalPrice.Mul(discountFactor)
	
	// finalPrice = originalPrice - discountAmount
	return originalPrice.Sub(discountAmount)
}

func (d *Discount) Equals(other *Discount) bool {
	if other == nil {
		return false
	}
	return d.percentage.Cmp(other.percentage) == 0 &&
		d.startDate.Equal(other.startDate) &&
		d.endDate.Equal(other.endDate)
}