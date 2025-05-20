package mongotest

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Env      = ""
	Database = "test"
)

func Connect(t *testing.T) func(*testing.T) *mongo.Database {
	t.Helper()

	// Skip unit testing if the env var is not configured
	var db string
	if Env != "" {
		db = os.Getenv(Env)
	}
	if db == "" {
		t.Skip("mongotest: no db uri was provided")
		return nil
	}

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			fmt.Printf("[MongoDB Command] %s: %v\n", evt.CommandName, evt.Command)
		},
	}

	c, _ := mongo.Connect(
		context.Background(),
		options.Client().
			ApplyURI(db).
			SetConnectTimeout(time.Second*10).
			SetMonitor(monitor),
	)

	return func(t *testing.T) *mongo.Database {
		t.Helper()

		databaseName := Database + "_" + uuid.NewString()
		t.Cleanup(func() {
			_ = c.Database(databaseName).Drop(context.Background())
		})
		return c.Database(databaseName)
	}
}
