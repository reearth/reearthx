package pubsub

import (
	"context"
	"sync"
	"testing"

	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/repository"
	"github.com/stretchr/testify/assert"
)

type mockPublisher struct {
	published []mockPublishedEvent
}

type mockPublishedEvent struct {
	topic string
	msg   interface{}
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, msg interface{}) error {
	// Make a copy of the event to ensure it's not modified after storage
	if event, ok := msg.(repository.AssetEvent); ok {
		eventCopy := repository.AssetEvent{
			Type:        event.Type,
			AssetID:     event.AssetID,
			WorkspaceID: event.WorkspaceID,
			ProjectID:   event.ProjectID,
			Status:      event.Status,
			Error:       event.Error,
		}
		m.published = append(m.published, mockPublishedEvent{topic: topic, msg: eventCopy})
	} else {
		m.published = append(m.published, mockPublishedEvent{topic: topic, msg: msg})
	}
	return nil
}

func TestNewAssetPubSub(t *testing.T) {
	pub := &mockPublisher{}
	ps := NewAssetPubSub(pub, "test-topic")
	assert.NotNil(t, ps)
	assert.Equal(t, pub, ps.publisher)
	assert.Equal(t, "test-topic", ps.topic)
}

func TestAssetPubSub_Subscribe(t *testing.T) {
	ps := NewAssetPubSub(&mockPublisher{}, "test-topic")

	var receivedEvents []repository.AssetEvent
	var mu sync.Mutex

	// Subscribe to all events
	ps.Subscribe("*", func(ctx context.Context, event repository.AssetEvent) {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
	})

	// Create test asset
	asset := entity.NewAsset(
		id.NewID(),
		"test.txt",
		100,
		"text/plain",
	)
	asset.MoveToWorkspace(id.NewWorkspaceID())
	asset.MoveToProject(id.NewProjectID())
	asset.UpdateStatus(entity.StatusActive, "")

	// Publish events
	ctx := context.Background()
	assert.NoError(t, ps.PublishAssetCreated(ctx, asset))
	assert.NoError(t, ps.PublishAssetUpdated(ctx, asset))
	assert.NoError(t, ps.PublishAssetUploaded(ctx, asset))

	// Check received events
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 3, len(receivedEvents))
	assert.Equal(t, repository.EventTypeAssetCreated, receivedEvents[0].Type)
	assert.Equal(t, repository.EventTypeAssetUpdated, receivedEvents[1].Type)
	assert.Equal(t, repository.EventTypeAssetUploaded, receivedEvents[2].Type)
}

func TestAssetPubSub_SubscribeSpecificEvent(t *testing.T) {
	ps := NewAssetPubSub(&mockPublisher{}, "test-topic")

	var receivedEvents []repository.AssetEvent
	var mu sync.Mutex

	// Subscribe only to created events
	ps.Subscribe(repository.EventTypeAssetCreated, func(ctx context.Context, event repository.AssetEvent) {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
	})

	// Create test asset
	asset := entity.NewAsset(
		id.NewID(),
		"test.txt",
		100,
		"text/plain",
	)

	// Publish different events
	ctx := context.Background()
	assert.NoError(t, ps.PublishAssetCreated(ctx, asset))  // Should be received
	assert.NoError(t, ps.PublishAssetUpdated(ctx, asset))  // Should be ignored
	assert.NoError(t, ps.PublishAssetUploaded(ctx, asset)) // Should be ignored

	// Check received events
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 1, len(receivedEvents))
	assert.Equal(t, repository.EventTypeAssetCreated, receivedEvents[0].Type)
}

func TestAssetPubSub_Unsubscribe(t *testing.T) {
	ps := NewAssetPubSub(&mockPublisher{}, "test-topic")

	var receivedEvents []repository.AssetEvent
	var mu sync.Mutex

	handler := func(ctx context.Context, event repository.AssetEvent) {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
	}

	// Subscribe and then unsubscribe
	ps.Subscribe(repository.EventTypeAssetCreated, handler)
	ps.Unsubscribe(repository.EventTypeAssetCreated, handler)

	// Create test asset
	asset := entity.NewAsset(
		id.NewID(),
		"test.txt",
		100,
		"text/plain",
	)

	// Publish event
	ctx := context.Background()
	assert.NoError(t, ps.PublishAssetCreated(ctx, asset))

	// Check that no events were received
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 0, len(receivedEvents))
}

func TestAssetPubSub_PublishEvents(t *testing.T) {
	pub := &mockPublisher{}
	ps := NewAssetPubSub(pub, "test-topic")

	// Create test asset
	asset := entity.NewAsset(
		id.NewID(),
		"test.txt",
		100,
		"text/plain",
	)
	workspaceID := id.NewWorkspaceID()
	projectID := id.NewProjectID()
	asset.MoveToWorkspace(workspaceID)
	asset.MoveToProject(projectID)

	// Set status and error before publishing
	asset.UpdateStatus(entity.StatusActive, "test error")

	// Test all publish methods
	ctx := context.Background()
	assert.NoError(t, ps.PublishAssetCreated(ctx, asset))
	assert.NoError(t, ps.PublishAssetUpdated(ctx, asset))
	assert.NoError(t, ps.PublishAssetDeleted(ctx, asset.ID()))
	assert.NoError(t, ps.PublishAssetUploaded(ctx, asset))
	assert.NoError(t, ps.PublishAssetExtracted(ctx, asset))
	assert.NoError(t, ps.PublishAssetTransferred(ctx, asset))

	// Verify published events
	assert.Equal(t, 6, len(pub.published))

	// Verify event details
	for i, event := range pub.published {
		assert.Equal(t, "test-topic", event.topic)
		assetEvent, ok := event.msg.(repository.AssetEvent)
		assert.True(t, ok, "Event message should be of type AssetEvent")
		assert.Equal(t, asset.ID(), assetEvent.AssetID)

		// For deleted event, we don't expect other fields
		if i == 2 { // deleted event
			assert.Equal(t, repository.EventTypeAssetDeleted, assetEvent.Type)
			assert.Empty(t, assetEvent.WorkspaceID)
			assert.Empty(t, assetEvent.ProjectID)
			assert.Empty(t, assetEvent.Status)
			assert.Empty(t, assetEvent.Error)
			continue
		}

		assert.Equal(t, workspaceID, assetEvent.WorkspaceID, "Event %d: WorkspaceID mismatch", i)
		assert.Equal(t, projectID, assetEvent.ProjectID, "Event %d: ProjectID mismatch", i)
		assert.Equal(t, string(asset.Status()), assetEvent.Status, "Event %d: Status mismatch", i)
		assert.Equal(t, asset.Error(), assetEvent.Error, "Event %d: Error mismatch", i)

		// Verify event types
		switch i {
		case 0:
			assert.Equal(t, repository.EventTypeAssetCreated, assetEvent.Type, "Event 0 should be Created")
		case 1:
			assert.Equal(t, repository.EventTypeAssetUpdated, assetEvent.Type, "Event 1 should be Updated")
		case 3:
			assert.Equal(t, repository.EventTypeAssetUploaded, assetEvent.Type, "Event 3 should be Uploaded")
		case 4:
			assert.Equal(t, repository.EventTypeAssetExtracted, assetEvent.Type, "Event 4 should be Extracted")
		case 5:
			assert.Equal(t, repository.EventTypeAssetTransferred, assetEvent.Type, "Event 5 should be Transferred")
		}
	}
}
