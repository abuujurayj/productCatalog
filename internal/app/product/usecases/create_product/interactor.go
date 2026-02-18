package create_product

import (
	"context"
	"math/big"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/pkg/clock"

	"github.com/Vektor-AI/commitplan"
)

type Request struct {
	Name        string
	Description string
	Category    string
	BasePrice   *big.Rat
}

type Interactor struct {
	productRepo contracts.ProductRepository
	outboxRepo  contracts.OutboxRepository
	committer   commitplan.Committer // Interface from Vektor-AI lib
	clock       clock.Clock
}

func NewInteractor(
	productRepo contracts.ProductRepository,
	outboxRepo contracts.OutboxRepository,
	committer commitplan.Committer,
	clock clock.Clock,
) *Interactor {
	return &Interactor{
		productRepo: productRepo,
		outboxRepo:  outboxRepo,
		committer:   committer,
		clock:       clock,
	}
}

func (it *Interactor) Execute(ctx context.Context, req Request) (string, error) {
	// 1. Create or load aggregate
	now := it.clock.Now()
	basePrice := domain.NewMoneyFromRat(req.BasePrice)
	
	product, err := domain.NewProduct(
		req.Name,
		req.Description,
		req.Category,
		basePrice,
		now,
	)
	if err != nil {
		return "", err
	}

	// 2. Build commit plan
	plan := commitplan.NewPlan()

	// 3. Get mutations from repository (repo returns, doesn't apply)
	if mut := it.productRepo.InsertMut(product); mut != nil {
		plan.Add(mut)
	}

	// 4. Add outbox events (Requirement 6: Enrich with metadata)
	metadata := map[string]interface{}{
		"user_id":    "system", 
		"company_id": "default",
	}
	
	for _, event := range product.DomainEvents() {
		if mut := it.outboxRepo.InsertMut(event, metadata); mut != nil {
			plan.Add(mut)
		}
	}

	// 5. Apply plan (Usecase applies, NOT handler!)
	if err := it.committer.Apply(ctx, plan); err != nil {
		return "", err
	}

	return product.ID(), nil
}