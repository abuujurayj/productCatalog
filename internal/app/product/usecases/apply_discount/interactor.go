package apply_discount

import (
	"context"
	"math/big"
	"time"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/pkg/clock"

	"github.com/Vektor-AI/commitplan"
)

type Request struct {
	ProductID  string
	Percentage *big.Rat
	StartDate  time.Time
	EndDate    time.Time
}

type Interactor struct {
	productRepo contracts.ProductRepository
	outboxRepo  contracts.OutboxRepository
	committer   commitplan.Committer
	clock       clock.Clock
}

func NewInteractor(
	productRepo contracts.ProductRepository,
	outboxRepo contracts.OutboxRepository,
	committer commitplan.Committer,
	clk clock.Clock,
) *Interactor {
	return &Interactor{
		productRepo: productRepo,
		outboxRepo:  outboxRepo,
		committer:   committer,
		clock:       clk,
	}
}

func (it *Interactor) Execute(ctx context.Context, req Request) error {
	// 1. Load aggregate (Rehydration)
	product, err := it.productRepo.Load(ctx, req.ProductID)
	if err != nil {
		return err
	}

	// 2. Create discount value object (Encapsulated validation)
	discount, err := domain.NewDiscount(req.Percentage, req.StartDate, req.EndDate)
	if err != nil {
		return err
	}

	// 3. Apply domain logic (Pure Logic)
	now := it.clock.Now()
	if err := product.ApplyDiscount(discount, now); err != nil {
		return err
	}

	// 4. Build commit plan
	plan := commitplan.NewPlan()

	// 5. Get update mutation (Optimized via ChangeTracker)
	if mut := it.productRepo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	// 6. Add outbox events (Enriched with metadata)
	metadata := map[string]interface{}{
		"user_id":    "system",
		"company_id": "default",
	}

	for _, event := range product.DomainEvents() {
		if mut := it.outboxRepo.InsertMut(event, metadata); mut != nil {
			plan.Add(mut)
		}
	}

	// 7. Apply plan atomically
	if err := it.committer.Apply(ctx, plan); err != nil {
		return err
	}

	return nil
}