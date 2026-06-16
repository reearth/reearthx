package pgxtest

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ApplyFS executes every *.sql file in fsys (lexical order) against pool. It is
// for test setup: point it at an embedded or on-disk migrations directory
// (os.DirFS / embed.FS) to materialize the schema on a fresh test database.
// Non-".sql" files (e.g. atlas.sum) are ignored. Each file is run as a single
// Exec, so multi-statement files work via the simple protocol.
func ApplyFS(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS) error {
	var files []string
	if err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".sql") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("pgxtest: walk migrations: %w", err)
	}
	sort.Strings(files)

	for _, f := range files {
		b, err := fs.ReadFile(fsys, f)
		if err != nil {
			return fmt.Errorf("pgxtest: read %s: %w", f, err)
		}
		if _, err := pool.Exec(ctx, string(b)); err != nil {
			return fmt.Errorf("pgxtest: apply %s: %w", f, err)
		}
	}
	return nil
}
