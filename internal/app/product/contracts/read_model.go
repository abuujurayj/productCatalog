package contracts

import (
	"context"
	"product-catalog-service/internal/app/product/queries/get_product"
	"product-catalog-service/internal/app/product/queries/list_products"
)

type ProductReadModel interface {
	GetProduct(ctx context.Context, id string) (*get_product.ProductDTO, error)
	ListProducts(ctx context.Context, params list_products.QueryParams) (*list_products.ListResult, error)
}