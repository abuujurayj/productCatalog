package m_outbox

import (
	"time"

	"cloud.google.com/go/spanner"
)

type OutboxRow struct {
	EventID     string    `spanner:"event_id"`
	EventType   string    `spanner:"event_type"`
	AggregateID string    `spanner:"aggregate_id"`
	Payload     spanner.NullJSON    `spanner:"payload"` // Must be NullJSON for JSON columns
	Status      string    `spanner:"status"`
	CreatedAt   time.Time `spanner:"created_at"`
	ProcessedAt spanner.NullTime `spanner:"processed_at"`
}

type Model struct{}

func (m *Model) TableName() string {
	return "outbox_events"
}

func (m *Model) InsertMut(eventID string, row *OutboxRow) *spanner.Mutation {
	return spanner.Insert(m.TableName(), []string{
		"event_id", "event_type", "aggregate_id", "payload",
		"status", "created_at", "processed_at",
	}, []interface{}{
		eventID, row.EventType, row.AggregateID, row.Payload,
		row.Status, row.CreatedAt, row.ProcessedAt,
	})
}

func (m *Model) UpdateStatusMut(eventID string, status string, processedAt *time.Time) *spanner.Mutation {
	updates := map[string]interface{}{
		"status": status,
	}
	if processedAt != nil {
		updates["processed_at"] = processedAt
	}
	
	keys := make([]string, 0, len(updates))
	values := make([]interface{}, 0, len(updates))
	for k, v := range updates {
		keys = append(keys, k)
		values = append(values, v)
	}
	
	return spanner.Update(m.TableName(), keys, append([]interface{}{eventID}, values...))
}