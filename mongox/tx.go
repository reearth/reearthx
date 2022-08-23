package mongox

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type Tx struct {
	session mongo.Session
	commit  bool
}

func (t *Tx) Commit() {
	if t == nil {
		return
	}
	t.commit = true
}

func (t *Tx) End(ctx context.Context) error {
	if t == nil {
		return nil
	}

	if t.commit {
		if err := t.session.CommitTransaction(ctx); err != nil {
			return err
		}
	} else if err := t.session.AbortTransaction(ctx); err != nil {
		return err
	}

	t.session.EndSession(ctx)
	return nil
}

func (t *Tx) IsCommitted() bool {
	return t.commit
}
