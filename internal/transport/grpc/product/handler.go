package product

import (
	"context"
	"math/big"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "product-catalog-service/gen/go/product/v1"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/queries/get_product"
	"product-catalog-service/internal/app/product/queries/list_products"
	"product-catalog-service/internal/app/product/usecases/activate_product"
	"product-catalog-service/internal/app/product/usecases/apply_discount"
	"product-catalog-service/internal/app/product/usecases/create_product"
	"product-catalog-service/internal/app/product/usecases/deactivate_product"
	"product-catalog-service/internal/app/product/usecases/remove_discount"
	"product-catalog-service/internal/app/product/usecases/update_product"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
)

type ProductHandler struct {
	pb.UnimplementedProductServiceServer

	client      *spanner.Client
	productRepo contracts.ProductRepository
	outboxRepo  contracts.OutboxRepository
	committer   *committer.Committer
	clock       clock.Clock

	// Commands
	createProduct     *create_product.Interactor
	updateProduct     *update_product.Interactor
	activateProduct   *activate_product.Interactor
	deactivateProduct *deactivate_product.Interactor
	applyDiscount     *apply_discount.Interactor
	removeDiscount    *remove_discount.Interactor

	// Queries
	getProduct   *get_product.Query
	listProducts *list_products.Query
}

func NewHandler(
	client *spanner.Client,
	productRepo contracts.ProductRepository,
	outboxRepo contracts.OutboxRepository,
	committer *committer.Committer,
	clock clock.Clock,
) *ProductHandler {
	return &ProductHandler{
		client:            client,
		productRepo:       productRepo,
		outboxRepo:        outboxRepo,
		committer:         committer,
		clock:             clock,
		createProduct:     create_product.NewInteractor(productRepo, outboxRepo, committer, clock),
		updateProduct:     update_product.NewInteractor(productRepo, outboxRepo, committer, clock),
		activateProduct:   activate_product.NewInteractor(productRepo, outboxRepo, committer, clock),
		deactivateProduct: deactivate_product.NewInteractor(productRepo, outboxRepo, committer, clock),
		applyDiscount:     apply_discount.NewInteractor(productRepo, outboxRepo, committer, clock),
		removeDiscount:    remove_discount.NewInteractor(productRepo, outboxRepo, committer, clock),
		getProduct:        get_product.NewQuery(client),
		listProducts:      list_products.NewQuery(client),
	}
}

// CreateProduct implements the gRPC command to create a new product.
func (h *ProductHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductReply, error) {
	// 1. Basic gRPC layer validation
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	// 2. Map Proto to UseCase Request
	appReq := create_product.Request{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Category:    req.GetCategory(),
		BasePrice:   mapProtoValueToRat(req.GetBasePrice()),
	}

	// 3. Execute Interactor
	id, err := h.createProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// 4. Return Response
	return &pb.CreateProductReply{
		ProductId: id,
	}, nil
}

// ApplyDiscount implements the gRPC command to apply a discount to an existing product.
func (h *ProductHandler) ApplyDiscount(ctx context.Context, req *pb.ApplyDiscountRequest) (*pb.ApplyDiscountReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := apply_discount.Request{
		ProductID:  req.GetProductId(),
		Percentage: mapProtoValueToRat(req.GetPercentage()),
		StartDate:  req.GetStartDate().AsTime(),
		EndDate:    req.GetEndDate().AsTime(),
	}

	err := h.applyDiscount.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.ApplyDiscountReply{
		Success: true,
	}, nil
}

// GetProduct implements the gRPC query to retrieve a single product by ID.
func (h *ProductHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	params := get_product.QueryParams{
		ID: req.GetProductId(),
	}

	dto, err := h.getProduct.Execute(ctx, params)
	if err != nil {
		// Spanner row not found check
		if err == spanner.ErrRowNotFound {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.GetProductReply{
		Product: &pb.Product{
			Id:              dto.ID,
			Name:            dto.Name,
			Description:     dto.Description,
			Category:        dto.Category,
			BasePrice:       dto.BasePrice.FloatString(2),      // Precision mapping
			EffectivePrice:  dto.EffectivePrice.FloatString(2), // Precision mapping
			Status:          dto.Status,
		},
	}, nil
}

// --- Helper Functions ---

// mapProtoValueToRat converts a string representation of a decimal to a *big.Rat.
// This ensures we maintain the precision required by the business logic.
func mapProtoValueToRat(val string) *big.Rat {
	r, ok := new(big.Rat).SetString(val)
	if !ok {
		return big.NewRat(0, 1)
	}
	return r
}

// mapDomainErrorToGRPC translates domain-specific sentinel errors into standard gRPC status codes.
func mapDomainErrorToGRPC(err error) error {
	switch err {
	case domain.ErrProductNotFound:
		return status.Error(codes.NotFound, err.Error())
	case domain.ErrProductNotActive:
		return status.Error(codes.FailedPrecondition, err.Error())
	case domain.ErrInvalidDiscountPeriod:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		// Log the actual error internally here
		return status.Error(codes.Internal, "an internal error occurred")
	}
}