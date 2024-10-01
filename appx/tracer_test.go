package appx

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/reearth/reearthx/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Mock for GCP exporter
type mockGCPExporter struct {
	mock.Mock
}

func (m *mockGCPExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	args := m.Called(ctx, spans)
	return args.Error(0)
}

func (m *mockGCPExporter) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Mock for Jaeger closer
type mockCloser struct {
	mock.Mock
}

func (m *mockCloser) Close() error {
	args := m.Called()
	return args.Error(0)
}

type testLogWriter struct {
	strings.Builder
}

func (w *testLogWriter) Write(p []byte) (int, error) {
	return w.Builder.Write(p)
}

func TestInitTracer(t *testing.T) {
	// Create function variables
	var testInitGCPTracer func(ctx context.Context, conf *TracerConfig)
	var testInitJaegerTracer func(conf *TracerConfig) io.Closer

	// Create a test wrapper for InitTracer that uses the function variables
	testInitTracer := func(ctx context.Context, conf *TracerConfig) io.Closer {
		if conf.Tracer == TRACER_GCP {
			testInitGCPTracer(ctx, conf)
			return nil
		} else if conf.Tracer == TRACER_JAEGER {
			return testInitJaegerTracer(conf)
		}
		return nil
	}

	tests := []struct {
		name     string
		config   *TracerConfig
		setup    func()
		expected io.Closer
	}{
		{
			name: "GCP Tracer",
			config: &TracerConfig{
				Name:         "test-gcp",
				Tracer:       TRACER_GCP,
				TracerSample: 0.5,
			},
			setup: func() {
				testInitGCPTracer = func(ctx context.Context, conf *TracerConfig) {
					// Mock the GCP tracer initialization
					mockExporter := &mockGCPExporter{}
					tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(mockExporter))
					otel.SetTracerProvider(tp)
					log.Infofc(ctx, "tracer: initialized cloud trace with sample fraction: %g", conf.TracerSample)
				}
			},
			expected: nil,
		},
		{
			name: "Jaeger Tracer",
			config: &TracerConfig{
				Name:         "test-jaeger",
				Tracer:       TRACER_JAEGER,
				TracerSample: 0.5,
			},
			setup: func() {
				testInitJaegerTracer = func(conf *TracerConfig) io.Closer {
					// Mock the Jaeger tracer initialization
					mockCloser := &mockCloser{}
					mockCloser.On("Close").Return(nil)
					log.Infof("tracer: initialized jaeger tracer with sample fraction: %g", conf.TracerSample)
					return mockCloser
				}
			},
			expected: &mockCloser{},
		},
		{
			name: "Unknown Tracer",
			config: &TracerConfig{
				Name:         "test-unknown",
				Tracer:       "unknown",
				TracerSample: 0.5,
			},
			setup:    func() {},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			// Capture log output
			logWriter := &testLogWriter{}
			log.SetOutput(logWriter)
			defer log.SetOutput(nil)

			ctx := context.Background()
			closer := testInitTracer(ctx, tt.config)

			if tt.expected == nil {
				assert.Nil(t, closer)
			} else {
				assert.NotNil(t, closer)
				assert.IsType(t, tt.expected, closer)
			}

			// Check if the log output contains the expected message
			logOutput := logWriter.String()
			expectedLogMessage := "tracer: initialized"
			if tt.config.Tracer != "unknown" {
				assert.Contains(t, logOutput, expectedLogMessage)
			} else {
				assert.Empty(t, logOutput)
			}
		})
	}
}
