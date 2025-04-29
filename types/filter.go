package types

type ActivityFilter struct {
	CategoryID *uint `form:"category_id"`
	StartDate  *string `form:"start_date"` // YYYY-MM-DD
	EndDate    *string `form:"end_date"` // YYYY-MM-DD
}