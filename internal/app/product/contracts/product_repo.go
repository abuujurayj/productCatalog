package contracts

import (
	"context"
	"product-catalog-service/internal/app/product/domain"

	"cloud.google.com/go/spanner"
)

type ProductRepository interface {
	// Load rehydrates the Product aggregate from Spanner
	Load(ctx context.Context, id string) (*domain.Product, error)
	
	// Exists checks for product existence (useful for Create validation)
	Exists(ctx context.Context, id string) (bool, error)

	// InsertMut returns the mutation for a new product
	InsertMut(product *domain.Product) *spanner.Mutation
	
	// UpdateMut returns the mutation for an existing product (partial update)
	UpdateMut(product *domain.Product) *spanner.Mutation
}

// type OutboxRepository interface {
// 	// CreateEventMut generates a mutation for a single domain event
// 	CreateEventMut(event domain.DomainEvent) *spanner.Mutation
// }