package e2e

import (
	"context"
	"math/big"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/queries/get_product"
	"product-catalog-service/internal/app/product/repo"
	"product-catalog-service/internal/app/product/usecases/activate_product"
	"product-catalog-service/internal/app/product/usecases/apply_discount"
	"product-catalog-service/internal/app/product/usecases/create_product"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
)

func setupTest(t *testing.T) (context.Context, *spanner.Client, func()) {
	ctx := context.Background()

	// SPANNER_EMULATOR_HOST must be set in your environment/docker-compose
	database := "projects/test-project/instances/test-instance/databases/test-database"
	client, err := spanner.NewClient(ctx, database)
	require.NoError(t, err)

	cleanup := func() {
		_, err := client.Apply(ctx, []*spanner.Mutation{
			spanner.Delete("products", spanner.AllKeys()),
			spanner.Delete("outbox_events", spanner.AllKeys()),
		})
		assert.NoError(t, err)
		client.Close()
	}

	return ctx, client, cleanup
}

func TestProductCreationFlow(t *testing.T) {
	ctx, client, cleanup := setupTest(t)
	defer cleanup()

	productRepo := repo.NewProductRepository(client)
	outboxRepo := repo.NewOutboxRepository()
	cm := committer.NewCommitter(client)
	cl := clock.RealClock{}

	createUC := create_product.NewInteractor(productRepo, outboxRepo, cm, cl)

	req := create_product.Request{
		Name:        "Test Product",
		Description: "Test Description",
		Category:    "electronics",
		BasePrice:   big.NewRat(1999, 100),
	}

	productID, err := createUC.Execute(ctx, req)
	require.NoError(t, err)
	require.NotEmpty(t, productID)

	getQuery := get_product.NewQuery(client)
	product, err := getQuery.Execute(ctx, get_product.QueryParams{ID: productID})
	require.NoError(t, err)

	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "inactive", product.Status)
	assert.Equal(t, 0, product.BasePrice.Cmp(big.NewRat(1999, 100)))
}

func TestDiscountApplicationFlow(t *testing.T) {
	ctx, client, cleanup := setupTest(t)
	defer cleanup()

	// Use fixed time to ensure discount is "active" during query
	fixedTime := time.Date(2026, 2, 18, 12, 0, 0, 0, time.UTC)
	testClock := clock.NewFakeClock(fixedTime)

	productRepo := repo.NewProductRepository(client)
	outboxRepo := repo.NewOutboxRepository()
	cm := committer.NewCommitter(client)

	// 1. Create
	createUC := create_product.NewInteractor(productRepo, outboxRepo, cm, testClock)
	productID, _ := createUC.Execute(ctx, create_product.Request{
		Name:      "Sale Item",
		Category:  "test",
		BasePrice: big.NewRat(100, 1), // $100.00
	})

	// 2. Activate
	activateUC := activate_product.NewInteractor(productRepo, outboxRepo, cm, testClock)
	_ = activateUC.Execute(ctx, activate_product.Request{ProductID: productID})

	// 3. Apply Discount
	applyUC := apply_discount.NewInteractor(productRepo, outboxRepo, cm, testClock)
	err := applyUC.Execute(ctx, apply_discount.Request{
		ProductID:  productID,
		Percentage: big.NewRat(25, 100), // 25% OFF
		StartDate:  fixedTime.Add(-1 * time.Hour),
		EndDate:    fixedTime.Add(1 * time.Hour),
	})
	require.NoError(t, err)

	// 4. Verify Pricing Calculation via Read Model
	getQuery := get_product.NewQuery(client)
	product, err := getQuery.Execute(ctx, get_product.QueryParams{ID: productID})
	require.NoError(t, err)

	// $100 - 25% = $75
	expected := big.NewRat(75, 1)
	assert.Equal(t, 0, product.EffectivePrice.Cmp(expected), "Effective price should be $75.00")
}

func TestBusinessRule_CannotDiscountInactiveProduct(t *testing.T) {
	ctx, client, cleanup := setupTest(t)
	defer cleanup()

	testClock := clock.NewFakeClock(time.Now())
	productRepo := repo.NewProductRepository(client)
	outboxRepo := repo.NewOutboxRepository()
	cm := committer.NewCommitter(client)

	createUC := create_product.NewInteractor(productRepo, outboxRepo, cm, testClock)
	productID, _ := createUC.Execute(ctx, create_product.Request{
		Name:      "Inactive Item",
		BasePrice: big.NewRat(10, 1),
	})

	applyUC := apply_discount.NewInteractor(productRepo, outboxRepo, cm, testClock)
	err := applyUC.Execute(ctx, apply_discount.Request{
		ProductID:  productID,
		Percentage: big.NewRat(1, 10),
		StartDate:  time.Now(),
		EndDate:    time.Now().Add(time.Hour),
	})

	assert.ErrorIs(t, err, domain.ErrProductNotActive)
}