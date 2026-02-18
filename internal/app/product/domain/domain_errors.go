package domain

import "errors"

var (
	ErrProductNotFound        = errors.New("product not found")
	ErrProductAlreadyExists   = errors.New("product already exists")
	ErrInvalidProductName     = errors.New("invalid product name")
	ErrInvalidProductPrice    = errors.New("invalid product price")
	ErrProductNotActive       = errors.New("product is not active")
	ErrProductAlreadyActive   = errors.New("product is already active")
	ErrProductAlreadyInactive = errors.New("product is already inactive")
	ErrProductArchived        = errors.New("product is archived")
	
	// Discount errors
	ErrDiscountAlreadyExists  = errors.New("discount already exists for this product")
	ErrNoDiscountToRemove     = errors.New("no discount to remove")
	ErrInvalidDiscountPeriod  = errors.New("invalid discount period")
	ErrDiscountPercentage     = errors.New("discount percentage must be between 0 and 100")
)