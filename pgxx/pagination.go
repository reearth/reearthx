package pgxx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// CountAndList runs an offset-paginated query: it executes countSQL to get the
// total, then runs listSQL with `LIMIT $n OFFSET $m` appended (the caller must
// NOT include LIMIT/OFFSET in listSQL), scanning rows into []T via scan
// (e.g. pgx.RowToStructByPos[gen.Foo]). countSQL and listSQL share the same
// leading args. Callers map total into their own page-info type
// (interfaces.PageBasedInfo, usecasex.PageInfo, …).
func CountAndList[T any](
	ctx context.Context,
	db DBTX,
	countSQL, listSQL string,
	args []any,
	limit, offset int64,
	scan pgx.RowToFunc[T],
) ([]T, int64, error) {
	var total int64
	if err := db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, WrapError(err)
	}

	pageArgs := append(append([]any{}, args...), limit, offset)
	q := fmt.Sprintf("%s LIMIT $%d OFFSET $%d", listSQL, len(args)+1, len(args)+2)
	rows, err := db.Query(ctx, q, pageArgs...)
	if err != nil {
		return nil, 0, WrapError(err)
	}
	defer rows.Close()

	list, err := pgx.CollectRows(rows, scan)
	if err != nil {
		return nil, 0, WrapError(err)
	}
	return list, total, nil
}

// List runs a non-paginated query and scans rows into []T via scan.
func List[T any](ctx context.Context, db DBTX, sql string, args []any, scan pgx.RowToFunc[T]) ([]T, error) {
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, WrapError(err)
	}
	defer rows.Close()

	list, err := pgx.CollectRows(rows, scan)
	if err != nil {
		return nil, WrapError(err)
	}
	return list, nil
}
