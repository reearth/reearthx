package migration

import (
	"context"
	"fmt"
	"slices"
	"sort"

	"github.com/reearth/reearthx/log"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/samber/lo"
)

type Key = int64
type MigrationFunc[C DBClient] func(context.Context, C) error
type Migrations[C DBClient] map[Key]MigrationFunc[C]

type DBClient interface {
	Transaction() usecasex.Transaction
}

type ConfigRepo interface {
	Current(ctx context.Context) (Key, error)
	Save(ctx context.Context, m Key) error
	Begin(ctx context.Context) error
	End(ctx context.Context) error
}

type Client[C DBClient] struct {
	client     C
	config     ConfigRepo
	migrations Migrations[C]
	retry      int
}

func NewClient[C DBClient](c C, config ConfigRepo, migrations Migrations[C], retry int) *Client[C] {
	return &Client[C]{
		client:     c,
		config:     config,
		migrations: migrations,
		retry:      retry,
	}
}

func (c Client[C]) Migrate(ctx context.Context) (err error) {
	if err := c.config.Begin(ctx); err != nil {
		return err
	}

	defer func() {
		if err2 := c.config.End(ctx); err == nil && err2 != nil {
			err = err2
		}
	}()

	current, err := c.config.Current(ctx)
	if err != nil {
		return fmt.Errorf("Failed to load config: %w", rerror.UnwrapErrInternal(err))
	}

	nextMigrations := nextMigration(lo.Keys(c.migrations), current)
	if len(nextMigrations) == 0 {
		return nil
	}

	tr := c.client.Transaction()
	for _, m := range nextMigrations {
		log.Infofc(ctx, "DB migration: %d\n", m)
		if err := usecasex.DoTransaction(ctx, tr, c.retry, func(ctx context.Context) error {
			if err := c.migrations[m](ctx, c.client); err != nil {
				return err
			}

			return c.config.Save(ctx, m)
		}); err != nil {
			return fmt.Errorf("Failed to exec migration %d: %w", m, rerror.UnwrapErrInternalOr(err))
		}
	}

	return nil
}

func nextMigration(migrations []Key, current Key) []Key {
	migrations2 := slices.Clone(migrations)
	sort.SliceStable(migrations2, func(i, j int) bool {
		return migrations2[i] < migrations2[j]
	})

	for i, m := range migrations2 {
		if len(migrations2) <= i {
			return nil
		}
		if current < m {
			return migrations2[i:]
		}
	}

	return nil
}
