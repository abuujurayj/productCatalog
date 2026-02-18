package domain

import "time"

type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
}

type BaseEvent struct {
	aggregateID string
	timestamp   time.Time
}

func (e BaseEvent) AggregateID() string { return e.aggregateID }
func (e BaseEvent) OccurredAt() time.Time { return e.timestamp }

type ProductCreatedEvent struct {
	BaseEvent
	Name        string
	Description string
	Category    string
	BasePrice   *Money
}

func (e ProductCreatedEvent) EventType() string { return "product.created" }

type ProductUpdatedEvent struct {
	BaseEvent
	Name        *string
	Description *string
	Category    *string
}

func (e ProductUpdatedEvent) EventType() string { return "product.updated" }

type ProductActivatedEvent struct {
	BaseEvent
}

func (e ProductActivatedEvent) EventType() string { return "product.activated" }

type ProductDeactivatedEvent struct {
	BaseEvent
}

func (e ProductDeactivatedEvent) EventType() string { return "product.deactivated" }

type DiscountAppliedEvent struct {
	BaseEvent
	Percentage string
	StartDate  time.Time
	EndDate    time.Time
}

func (e DiscountAppliedEvent) EventType() string { return "discount.applied" }

type DiscountRemovedEvent struct {
	BaseEvent
}

func (e DiscountRemovedEvent) EventType() string { return "discount.removed" }