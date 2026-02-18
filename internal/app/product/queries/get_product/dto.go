package get_product

import (
	"math/big"
	"time"
)

type ProductDTO struct {
	ID              string
	Name            string
	Description     string
	Category        string
	BasePrice       *big.Rat
	DiscountedPrice *big.Rat // The amount saved
	EffectivePrice  *big.Rat // Final price after all rules
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ArchivedAt      *time.Time
}