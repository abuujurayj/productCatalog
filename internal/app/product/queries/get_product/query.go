package get_product

import (
	"context"
	"math/big"
	"time"

	"cloud.google.com/go/spanner"
)

type Query struct {
	client *spanner.Client
}

func NewQuery(client *spanner.Client) *Query {
	return &Query{
		client: client,
	}
}

type QueryParams struct {
	ID string
}

func (q *Query) Execute(ctx context.Context, params QueryParams) (*ProductDTO, error) {
	stmt := spanner.Statement{
		SQL: `SELECT 
                product_id, name, description, category,
                base_price_numerator, base_price_denominator,
                discount_percent, discount_start_date, discount_end_date,
                status, created_at, updated_at, archived_at
              FROM products 
              WHERE product_id = @id`,
		Params: map[string]interface{}{"id": params.ID},
	}

	iter := q.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err != nil {
		return nil, err 
	}

	var (
		productID            string
		name                 string
		description          string
		category             string
		basePriceNumerator   int64
		basePriceDenominator int64
		discountPercent      spanner.NullNumeric
		discountStartDate    spanner.NullTime
		discountEndDate      spanner.NullTime
		status               string
		createdAt            time.Time
		updatedAt            time.Time
		archivedAt           spanner.NullTime
	)

	err = row.Columns(
		&productID, &name, &description, &category,
		&basePriceNumerator, &basePriceDenominator,
		&discountPercent, &discountStartDate, &discountEndDate,
		&status, &createdAt, &updatedAt, &archivedAt,
	)
	if err != nil {
		return nil, err
	}

	// 1. Calculate Base Price
	basePrice := big.NewRat(basePriceNumerator, basePriceDenominator)

	// 2. Calculate Effective Price
	effectivePrice := new(big.Rat).Set(basePrice)
	var discountedPrice *big.Rat

	now := time.Now()
	// Logic: If discount is active (start <= now < end)
	if discountPercent.Valid && discountStartDate.Valid && discountEndDate.Valid {
		if now.After(discountStartDate.Time) && now.Before(discountEndDate.Time) {
			// percentage is big.Rat from Spanner NullNumeric
			pct := discountPercent.Numeric 
			
			// calculation: discount_amount = base * percentage
			discountAmount := new(big.Rat).Mul(basePrice, &pct)
			
			// calculation: effective = base - discount_amount
			effectivePrice.Sub(basePrice, discountAmount)
			discountedPrice = discountAmount
		}
	}

	var archAt *time.Time
	if archivedAt.Valid {
		archAt = &archivedAt.Time
	}

	return &ProductDTO{
		ID:              productID,
		Name:            name,
		Description:     description,
		Category:        category,
		BasePrice:       basePrice,
		DiscountedPrice: discountedPrice,
		EffectivePrice:  effectivePrice,
		Status:          status,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		ArchivedAt:      archAt,
	}, nil
}