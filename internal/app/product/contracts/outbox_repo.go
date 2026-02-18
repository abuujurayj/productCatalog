package contracts

import (
	"product-catalog-service/internal/app/product/domain"

	"cloud.google.com/go/spanner"
)

type OutboxRepository interface {
	// InsertMut returns mutation for inserting an outbox event
	InsertMut(event domain.DomainEvent, metadata map[string]interface{}) *spanner.Mutation
}