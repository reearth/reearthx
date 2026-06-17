package pgxx

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/reearth/reearthx/usecasex"
)

// KeysetPaginate runs single-column keyset (Relay-style) pagination driven by
// usecasex.CursorPagination — the SQL analogue of mongox cursor pagination.
//
// keyCol must be a unique, sortable column (typically the primary key); a row's
// cursor IS that column's value, which cursorOf returns. table is the FROM
// target (e.g. "triggers"), columns the SELECT list (e.g. "*" or "id, name"),
// where an optional predicate WITHOUT the "WHERE" keyword (may be ""), and args
// the placeholders referenced by where ($1..$len(args)).
//
// It runs a COUNT over the filtered set (TotalCount), then a keyset page query,
// and returns rows in ascending keyCol order with a usecasex.PageInfo:
//   - First (+optional After) → forward page; Last (+optional Before) → backward.
//   - hasMore is detected with a LIMIT n+1 probe; the extra row is trimmed.
func KeysetPaginate[T any](
	ctx context.Context,
	db DBTX,
	table, columns, where string,
	args []any,
	keyCol string,
	p *usecasex.CursorPagination,
	scan pgx.RowToFunc[T],
	cursorOf func(T) usecasex.Cursor,
) ([]T, *usecasex.PageInfo, error) {
	whereClause := func(extra ...string) string {
		conds := make([]string, 0, len(extra)+1)
		if where != "" {
			conds = append(conds, where)
		}
		for _, e := range extra {
			if e != "" {
				conds = append(conds, e)
			}
		}
		if len(conds) == 0 {
			return ""
		}
		return " WHERE " + strings.Join(conds, " AND ")
	}

	var total int64
	if err := db.QueryRow(ctx, "SELECT count(*) FROM "+table+whereClause(), args...).Scan(&total); err != nil {
		return nil, nil, WrapError(err)
	}

	// Forward unless Last is the only bound set.
	forward := p == nil || p.First != nil || p.Last == nil
	pageArgs := append([]any{}, args...)

	var boundary, order string
	var limited bool
	var limit int64
	if forward {
		if p != nil && p.After != nil {
			pageArgs = append(pageArgs, string(*p.After))
			boundary = fmt.Sprintf("%s > $%d", keyCol, len(pageArgs))
		}
		order = " ORDER BY " + keyCol + " ASC"
		if p != nil && p.First != nil {
			limited, limit = true, *p.First
		}
	} else {
		if p.Before != nil {
			pageArgs = append(pageArgs, string(*p.Before))
			boundary = fmt.Sprintf("%s < $%d", keyCol, len(pageArgs))
		}
		order = " ORDER BY " + keyCol + " DESC"
		if p.Last != nil {
			limited, limit = true, *p.Last
		}
	}

	q := "SELECT " + columns + " FROM " + table + whereClause(boundary) + order
	if limited {
		pageArgs = append(pageArgs, limit+1) // +1 to detect hasMore
		q += fmt.Sprintf(" LIMIT $%d", len(pageArgs))
	}

	list, err := List[T](ctx, db, q, pageArgs, scan)
	if err != nil {
		return nil, nil, err
	}

	hasMore := limited && int64(len(list)) > limit
	if hasMore {
		list = list[:limit]
	}
	if !forward { // queried DESC; restore ascending keyCol order
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
	}
	hasNext := (forward && hasMore) || (!forward && p != nil && p.Before != nil)
	hasPrev := (!forward && hasMore) || (forward && p != nil && p.After != nil)

	var start, end *usecasex.Cursor
	if len(list) > 0 {
		s := cursorOf(list[0])
		e := cursorOf(list[len(list)-1])
		start, end = &s, &e
	}
	return list, usecasex.NewPageInfo(total, start, end, hasNext, hasPrev), nil
}
