package repo

import (
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/models/m_outbox"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
)

type outboxRepo struct {
	dbModel *m_outbox.Model
}

func NewOutboxRepository() *outboxRepo {
	return &outboxRepo{
		dbModel: &m_outbox.Model{},
	}
}

func (r *outboxRepo) InsertMut(event domain.DomainEvent, metadata map[string]interface{}) *spanner.Mutation {
	// Create an "Envelope" for the event. 
	envelope := struct {
		Event    domain.DomainEvent     `json:"event"`
		Metadata map[string]interface{} `json:"metadata"`
	}{
		Event:    event,
		Metadata: metadata,
	}

	// Spanner's JSON type can accept structs directly if using the spanner.NullJSON
	// or it can be handled via the row mapping if m_outbox.OutboxRow is configured correctly.
	row := &m_outbox.OutboxRow{
		EventID:     uuid.New().String(),
		EventType:   event.EventType(),
		AggregateID: event.AggregateID(),
		// Use NullJSON to correctly map to the JSON column type in your SQL
		Payload: spanner.NullJSON{
			Value: envelope,
			Valid: true,
		},
		Status:      "pending", 
		CreatedAt:   event.OccurredAt(),
		ProcessedAt: spanner.NullTime{Valid: false},
	}

	return r.dbModel.InsertMut(row.EventID, row)
}