package repo

import (
	"context"
	"time"

	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/models/m_product"

	"cloud.google.com/go/spanner"
)

type productRepo struct {
	client  *spanner.Client
	dbModel *m_product.Model
}

func NewProductRepository(client *spanner.Client) *productRepo {
	return &productRepo{
		client:  client,
		dbModel: &m_product.Model{},
	}
}

func (r *productRepo) Load(ctx context.Context, id string) (*domain.Product, error) {
	row, err := r.readRow(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.mapToDomain(row), nil
}

func (r *productRepo) readRow(ctx context.Context, id string) (*m_product.ProductRow, error) {
	stmt := spanner.Statement{
		SQL: `SELECT product_id, name, description, category, 
                     base_price_numerator, base_price_denominator,
                     discount_percent, discount_start_date, discount_end_date,
                     status, created_at, updated_at, archived_at
              FROM products WHERE product_id = @id`,
		Params: map[string]interface{}{"id": id},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err != nil {
		if err == spanner.ErrRowNotFound {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	var productRow m_product.ProductRow
	if err := row.ToStruct(&productRow); err != nil {
		return nil, err
	}

	return &productRow, nil
}

func (r *productRepo) mapToDomain(row *m_product.ProductRow) *domain.Product {
	// Map base price
	basePrice := domain.NewMoneyFromInt64(
		row.BasePriceNumerator,
		row.BasePriceDenominator,
	)

	// Map discount if exists
	var discount *domain.Discount
	if row.DiscountPercent != nil && row.DiscountPercent.Valid {
		// row.DiscountPercent.Numeric is already a big.Rat
		discount, _ = domain.NewDiscount(
			&row.DiscountPercent.Numeric,
			row.DiscountStartDate.Time,
			row.DiscountEndDate.Time,
		)
	}

	// Map archived at
	var archivedAt *time.Time
	if row.ArchivedAt != nil && row.ArchivedAt.Valid {
		archivedAt = &row.ArchivedAt.Time
	}

	return domain.RebuildProduct(
		row.ProductID,
		row.Name,
		row.Description,
		row.Category,
		basePrice,
		discount,
		domain.ProductStatus(row.Status),
		row.CreatedAt,
		row.UpdatedAt,
		archivedAt,
	)
}

func (r *productRepo) InsertMut(product *domain.Product) *spanner.Mutation {
	row := r.mapToRow(product)
	return r.dbModel.InsertMut(product.ID(), row)
}

func (r *productRepo) UpdateMut(product *domain.Product) *spanner.Mutation {
	if !product.Changes().HasChanges() {
		return nil
	}

	updates := make(map[string]interface{})
	ct := product.Changes()

	if ct.Dirty(domain.FieldName) {
		updates[m_product.Name] = product.Name()
	}

	if ct.Dirty(domain.FieldDescription) {
		updates[m_product.Description] = product.Description()
	}

	if ct.Dirty(domain.FieldCategory) {
		updates[m_product.Category] = product.Category()
	}

	if ct.Dirty(domain.FieldBasePrice) {
		updates[m_product.BasePriceNumerator] = product.BasePrice().Numerator()
		updates[m_product.BasePriceDenominator] = product.BasePrice().Denominator()
	}

	if ct.Dirty(domain.FieldDiscount) {
		if d := product.Discount(); d != nil {
			// Set discount fields using high-precision big.Rat
			updates[m_product.DiscountPercent] = d.Percentage()
			updates[m_product.DiscountStartDate] = d.StartDate()
			updates[m_product.DiscountEndDate] = d.EndDate()
		} else {
			// Explicitly clear discount by setting columns to NULL
			updates[m_product.DiscountPercent] = spanner.NullNumeric{Valid: false}
			updates[m_product.DiscountStartDate] = spanner.NullTime{Valid: false}
			updates[m_product.DiscountEndDate] = spanner.NullTime{Valid: false}
		}
	}

	if ct.Dirty(domain.FieldStatus) {
		updates[m_product.Status] = string(product.Status())
	}

	// Always apply the timestamp from the aggregate
	updates[m_product.UpdatedAt] = product.UpdatedAt()

	return r.dbModel.UpdateMut(product.ID(), updates)
}

func (r *productRepo) mapToRow(product *domain.Product) *m_product.ProductRow {
	row := &m_product.ProductRow{
		ProductID:            product.ID(),
		Name:                 product.Name(),
		Description:          product.Description(),
		Category:             product.Category(),
		BasePriceNumerator:   product.BasePrice().Numerator(),
		BasePriceDenominator: product.BasePrice().Denominator(),
		Status:               string(product.Status()),
		CreatedAt:            product.CreatedAt(),
		UpdatedAt:            product.UpdatedAt(),
	}

	if d := product.Discount(); d != nil {
		row.DiscountPercent = &spanner.NullNumeric{Numeric: *d.Percentage(), Valid: true}
		row.DiscountStartDate = &spanner.NullTime{Time: d.StartDate(), Valid: true}
		row.DiscountEndDate = &spanner.NullTime{Time: d.EndDate(), Valid: true}
	} else {
		row.DiscountPercent = &spanner.NullNumeric{Valid: false}
		row.DiscountStartDate = &spanner.NullTime{Valid: false}
		row.DiscountEndDate = &spanner.NullTime{Valid: false}
	}

	if at := product.ArchivedAt(); at != nil {
		row.ArchivedAt = &spanner.NullTime{Time: *at, Valid: true}
	} else {
		row.ArchivedAt = &spanner.NullTime{Valid: false}
	}

	return row
}

func (r *productRepo) Exists(ctx context.Context, id string) (bool, error) {
	stmt := spanner.Statement{
		SQL:    "SELECT 1 FROM products WHERE product_id = @id",
		Params: map[string]interface{}{"id": id},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	_, err := iter.Next()
	if err == spanner.ErrRowNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}