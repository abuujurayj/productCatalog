package list_products

type QueryParams struct {
	Category string
	PageSize int
	Offset   int
}

type ListResult struct {
	Products []interface{} // Replace interface{} with your product model later
	Total    int
}