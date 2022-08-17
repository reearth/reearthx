package usecasex

// Pagination is a struct for Relay-Style Cursor Pagination
// ref: https://www.apollographql.com/docs/react/features/pagination/#relay-style-cursor-pagination
type Pagination struct {
	Before *Cursor
	After  *Cursor
	First  *int
	Last   *int
}

func NewPagination(first *int, last *int, before *Cursor, after *Cursor) *Pagination {
	return &Pagination{
		Before: before,
		After:  after,
		First:  first,
		Last:   last,
	}
}
