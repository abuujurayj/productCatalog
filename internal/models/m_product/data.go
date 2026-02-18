package m_product

import (
	"time"

	"cloud.google.com/go/spanner"
)

type ProductRow struct {
	ProductID            string    `spanner:"product_id"`
	Name                 string    `spanner:"name"`
	Description          string    `spanner:"description"`
	Category             string    `spanner:"category"`
	BasePriceNumerator   int64     `spanner:"base_price_numerator"`
	BasePriceDenominator int64     `spanner:"base_price_denominator"`
	DiscountPercent      *spanner.NullNumeric `spanner:"discount_percent"`
	DiscountStartDate    *spanner.NullTime    `spanner:"discount_start_date"`
	DiscountEndDate      *spanner.NullTime    `spanner:"discount_end_date"`
	Status               string    `spanner:"status"`
	CreatedAt            time.Time `spanner:"created_at"`
	UpdatedAt            time.Time `spanner:"updated_at"`
	ArchivedAt           *spanner.NullTime    `spanner:"archived_at"`
}

type Model struct{}

func (m *Model) TableName() string {
	return "products"
}

func (m *Model) InsertMut(productID string, row *ProductRow) *spanner.Mutation {
	return spanner.Insert(m.TableName(), []string{
		"product_id", "name", "description", "category",
		"base_price_numerator", "base_price_denominator",
		"discount_percent", "discount_start_date", "discount_end_date",
		"status", "created_at", "updated_at", "archived_at",
	}, []interface{}{
		productID, row.Name, row.Description, row.Category,
		row.BasePriceNumerator, row.BasePriceDenominator,
		row.DiscountPercent, row.DiscountStartDate, row.DiscountEndDate,
		row.Status, row.CreatedAt, row.UpdatedAt, row.ArchivedAt,
	})
}

// func (m *Model) UpdateMut(productID string, updates map[string]interface{}) *spanner.Mutation {
// 	keys := make([]string, 0, len(updates))
// 	values := make([]interface{}, 0, len(updates))
	
// 	for k, v := range updates {
// 		keys = append(keys, k)
// 		values = append(values, v)
// 	}
	
// 	return spanner.Update(m.TableName(), keys, append([]interface{}{productID}, values...))
// }

func (m *Model) UpdateMut(productID string, updates map[string]interface{}) *spanner.Mutation {
    cols := []string{"product_id"}
    vals := []interface{}{productID}
    
    for k, v := range updates {
        cols = append(cols, k)
        vals = append(vals, v)
    }
    
    return spanner.Update(m.TableName(), cols, vals)
}

func (m *Model) DeleteMut(productID string) *spanner.Mutation {
	return spanner.Delete(m.TableName(), spanner.Key{productID})
}