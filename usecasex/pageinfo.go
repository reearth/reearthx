package usecasex

type PageInfo struct {
	TotalCount      int
	StartCursor     *Cursor
	EndCursor       *Cursor
	HasNextPage     bool
	HasPreviousPage bool
}

func NewPageInfo(totalCount int, startCursor, endCursor *Cursor, hasNextPage, hasPreviousPage bool) *PageInfo {
	return &PageInfo{
		TotalCount:      totalCount,
		StartCursor:     startCursor.CopyRef(),
		EndCursor:       endCursor.CopyRef(),
		HasNextPage:     hasNextPage,
		HasPreviousPage: hasPreviousPage,
	}
}

func EmptyPageInfo() *PageInfo {
	return &PageInfo{
		TotalCount:      0,
		StartCursor:     nil,
		EndCursor:       nil,
		HasNextPage:     false,
		HasPreviousPage: false,
	}
}

func (p *PageInfo) OrEmpty() *PageInfo {
	if p == nil {
		return EmptyPageInfo()
	}
	return p
}

func (p *PageInfo) Clone() *PageInfo {
	if p == nil {
		return nil
	}
	return &PageInfo{
		TotalCount:      p.TotalCount,
		StartCursor:     p.StartCursor.CopyRef(),
		EndCursor:       p.EndCursor.CopyRef(),
		HasNextPage:     p.HasNextPage,
		HasPreviousPage: p.HasPreviousPage,
	}
}
