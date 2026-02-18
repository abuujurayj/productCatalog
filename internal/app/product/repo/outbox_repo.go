package repo

import (
	"encoding/json"

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
	// This ensures the consumer always has a consistent JSON structure.
	envelope := struct {
		Event    domain.DomainEvent     `json:"event"`
		Metadata map[string]interface{} `json:"metadata"`
	}{
		Event:    event,
		Metadata: metadata,
	}

	payloadBytes, _ := json.Marshal(envelope)

	row := &m_outbox.OutboxRow{
		EventID:     uuid.New().String(),
		EventType:   event.EventType(),
		AggregateID: event.AggregateID(),
		Payload:     string(payloadBytes),
		Status:      "pending", // Matches initial_schema.sql logic
		CreatedAt:   event.OccurredAt(),
		ProcessedAt: &spanner.NullTime{Valid: false},
	}

	return r.dbModel.InsertMut(row.EventID, row)
}