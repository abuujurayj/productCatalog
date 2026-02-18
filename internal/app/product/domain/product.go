package domain

import (
	"math/big"
	"time"

	"github.com/google/uuid"
)

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
	ProductStatusArchived ProductStatus = "archived"
)

// Field constants for change tracking
const (
	FieldName        = "name"
	FieldDescription = "description"
	FieldCategory    = "category"
	FieldBasePrice   = "base_price"
	FieldDiscount    = "discount"
	FieldStatus      = "status"
)

// Product aggregate root
type Product struct {
	id          string
	name        string
	description string
	category    string
	basePrice   *Money
	discount    *Discount
	status      ProductStatus
	createdAt   time.Time
	updatedAt   time.Time
	archivedAt  *time.Time
	changes     *ChangeTracker
	events      []DomainEvent
}

// Constructor
func NewProduct(name, description, category string, basePrice *Money, now time.Time) (*Product, error) {
	if name == "" {
		return nil, ErrInvalidProductName
	}
	if basePrice == nil || basePrice.Amount().Cmp(big.NewRat(0, 1)) <= 0 {
		return nil, ErrInvalidProductPrice
	}
	if category == "" {
		category = "uncategorized"
	}

	product := &Product{
		id:          uuid.New().String(),
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		status:      ProductStatusInactive, // Start as inactive
		createdAt:   now,
		updatedAt:   now,
		changes:     NewChangeTracker(),
		events:      []DomainEvent{},
	}

	// Record creation event
	product.events = append(product.events, ProductCreatedEvent{
		BaseEvent:   BaseEvent{aggregateID: product.id, timestamp: now},
		Name:        name,
		Description: description,
		Category:    category,
		BasePrice:   basePrice,
	})

	return product, nil
}

// Getters
func (p *Product) ID() string                    { return p.id }
func (p *Product) Name() string                   { return p.name }
func (p *Product) Description() string             { return p.description }
func (p *Product) Category() string                { return p.category }
func (p *Product) BasePrice() *Money               { return p.basePrice }
func (p *Product) Discount() *Discount              { return p.discount }
func (p *Product) Status() ProductStatus            { return p.status }
func (p *Product) CreatedAt() time.Time             { return p.createdAt }
func (p *Product) UpdatedAt() time.Time             { return p.updatedAt }
func (p *Product) ArchivedAt() *time.Time           { return p.archivedAt }
func (p *Product) Changes() *ChangeTracker          { return p.changes }
func (p *Product) DomainEvents() []DomainEvent      { return p.events }

// Business methods
func (p *Product) Update(name, description, category *string, now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}

	if name != nil && *name != "" && *name != p.name {
		p.name = *name
		p.changes.MarkDirty(FieldName)
	}
	if description != nil && *description != p.description {
		p.description = *description
		p.changes.MarkDirty(FieldDescription)
	}
	if category != nil && *category != "" && *category != p.category {
		p.category = *category
		p.changes.MarkDirty(FieldCategory)
	}

	if p.changes.HasChanges() {
		p.updatedAt = now
		
		// Record update event
		var namePtr, descPtr, catPtr *string
		if p.changes.Dirty(FieldName) {
			namePtr = &p.name
		}
		if p.changes.Dirty(FieldDescription) {
			descPtr = &p.description
		}
		if p.changes.Dirty(FieldCategory) {
			catPtr = &p.category
		}
		
		p.events = append(p.events, ProductUpdatedEvent{
			BaseEvent:   BaseEvent{aggregateID: p.id, timestamp: now},
			Name:        namePtr,
			Description: descPtr,
			Category:    catPtr,
		})
	}

	return nil
}

func (p *Product) Activate(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}
	if p.status == ProductStatusActive {
		return ErrProductAlreadyActive
	}

	p.status = ProductStatusActive
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)
	
	p.events = append(p.events, ProductActivatedEvent{
		BaseEvent: BaseEvent{aggregateID: p.id, timestamp: now},
	})
	
	return nil
}

func (p *Product) Deactivate(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}
	if p.status == ProductStatusInactive {
		return ErrProductAlreadyInactive
	}

	p.status = ProductStatusInactive
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)
	
	p.events = append(p.events, ProductDeactivatedEvent{
		BaseEvent: BaseEvent{aggregateID: p.id, timestamp: now},
	})
	
	return nil
}

func (p *Product) Archive(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}

	p.status = ProductStatusArchived
	p.archivedAt = &now
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)
	
	return nil
}

func (p *Product) ApplyDiscount(discount *Discount, now time.Time) error {
	if p.status != ProductStatusActive {
		return ErrProductNotActive
	}
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}
	if p.discount != nil {
		return ErrDiscountAlreadyExists
	}
	if !discount.IsValidAt(now) {
		return ErrInvalidDiscountPeriod
	}

	p.discount = discount
	p.updatedAt = now
	p.changes.MarkDirty(FieldDiscount)
	
	p.events = append(p.events, DiscountAppliedEvent{
		BaseEvent:  BaseEvent{aggregateID: p.id, timestamp: now},
		Percentage: discount.percentage.FloatString(4),
		StartDate:  discount.startDate,
		EndDate:    discount.endDate,
	})
	
	return nil
}

func (p *Product) RemoveDiscount(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}
	if p.discount == nil {
		return ErrNoDiscountToRemove
	}

	p.discount = nil
	p.updatedAt = now
	p.changes.MarkDirty(FieldDiscount)
	
	p.events = append(p.events, DiscountRemovedEvent{
		BaseEvent: BaseEvent{aggregateID: p.id, timestamp: now},
	})
	
	return nil
}

func (p *Product) EffectivePrice(now time.Time) *Money {
	if p.discount != nil && p.discount.IsValidAt(now) {
		return p.discount.CalculateDiscountedPrice(p.basePrice)
	}
	return p.basePrice
}

// Rebuild from database (factory method for repository)
func RebuildProduct(
	id string,
	name string,
	description string,
	category string,
	basePrice *Money,
	discount *Discount,
	status ProductStatus,
	createdAt time.Time,
	updatedAt time.Time,
	archivedAt *time.Time,
) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		discount:    discount,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		archivedAt:  archivedAt,
		changes:     NewChangeTracker(), // Fresh aggregate, no changes
		events:      []DomainEvent{},
	}
}

// ClearEvents clears domain events (call after persisting)
func (p *Product) ClearEvents() {
	p.events = []DomainEvent{}
}