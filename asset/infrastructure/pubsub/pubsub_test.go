package pubsub

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/asset/domain"
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
	m.published = append(m.published, mockPublishedEvent{topic: topic, msg: msg})
	return nil
}

func TestNewAssetPubSub(t *testing.T) {
	pub := &mockPublisher{}
	ps := NewAssetPubSub(pub, "test-topic")
	assert.NotNil(t, ps)
	assert.Equal(t, pub, ps.publisher)
	assert.Equal(t, "test-topic", ps.topic)
}

func TestAssetPubSub_PublishEvents(t *testing.T) {
	ctx := context.Background()
	pub := &mockPublisher{}
	ps := NewAssetPubSub(pub, "test-topic")

	// Create test asset
	asset := domain.NewAsset(
		domain.NewID(),
		"test.txt",
		100,
		"text/plain",
	)
	asset.MoveToWorkspace(domain.NewWorkspaceID())
	asset.MoveToProject(domain.NewProjectID())
	asset.UpdateStatus(domain.StatusActive, "")

	tests := []struct {
		name     string
		publish  func() error
		expected repository.AssetEvent
	}{
		{
			name: "publish created event",
			publish: func() error {
				return ps.PublishAssetCreated(ctx, asset)
			},
			expected: repository.AssetEvent{
				Type:        repository.EventTypeAssetCreated,
				AssetID:     asset.ID(),
				WorkspaceID: asset.WorkspaceID(),
				ProjectID:   asset.ProjectID(),
				Status:      asset.Status(),
				Error:       asset.Error(),
			},
		},
		{
			name: "publish updated event",
			publish: func() error {
				return ps.PublishAssetUpdated(ctx, asset)
			},
			expected: repository.AssetEvent{
				Type:        repository.EventTypeAssetUpdated,
				AssetID:     asset.ID(),
				WorkspaceID: asset.WorkspaceID(),
				ProjectID:   asset.ProjectID(),
				Status:      asset.Status(),
				Error:       asset.Error(),
			},
		},
		{
			name: "publish deleted event",
			publish: func() error {
				return ps.PublishAssetDeleted(ctx, asset.ID())
			},
			expected: repository.AssetEvent{
				Type:    repository.EventTypeAssetDeleted,
				AssetID: asset.ID(),
			},
		},
		{
			name: "publish uploaded event",
			publish: func() error {
				return ps.PublishAssetUploaded(ctx, asset)
			},
			expected: repository.AssetEvent{
				Type:        repository.EventTypeAssetUploaded,
				AssetID:     asset.ID(),
				WorkspaceID: asset.WorkspaceID(),
				ProjectID:   asset.ProjectID(),
				Status:      asset.Status(),
				Error:       asset.Error(),
			},
		},
		{
			name: "publish extracted event",
			publish: func() error {
				return ps.PublishAssetExtracted(ctx, asset)
			},
			expected: repository.AssetEvent{
				Type:        repository.EventTypeAssetExtracted,
				AssetID:     asset.ID(),
				WorkspaceID: asset.WorkspaceID(),
				ProjectID:   asset.ProjectID(),
				Status:      asset.Status(),
				Error:       asset.Error(),
			},
		},
		{
			name: "publish transferred event",
			publish: func() error {
				return ps.PublishAssetTransferred(ctx, asset)
			},
			expected: repository.AssetEvent{
				Type:        repository.EventTypeAssetTransferred,
				AssetID:     asset.ID(),
				WorkspaceID: asset.WorkspaceID(),
				ProjectID:   asset.ProjectID(),
				Status:      asset.Status(),
				Error:       asset.Error(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous events
			pub.published = nil

			// Publish event
			err := tt.publish()
			assert.NoError(t, err)

			// Check published event
			assert.Len(t, pub.published, 1)
			assert.Equal(t, "test-topic", pub.published[0].topic)
			assert.Equal(t, tt.expected, pub.published[0].msg)
		})
	}
}
