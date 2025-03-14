package appx

import (
	"context"
	"io"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/reearth/reearthx/log"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Tracer string

const TRACER_GCP = Tracer("gcp")
const TRACER_JAEGER = Tracer("jaeger")

type TracerConfig struct {
	Name         string
	Tracer       Tracer
	TracerSample float64
}

func InitTracer(ctx context.Context, conf *TracerConfig) io.Closer {
	if conf.Tracer == TRACER_GCP {
		initGCPTracer(ctx, conf)
	} else if conf.Tracer == TRACER_JAEGER {
		return initJaegerTracer(conf)
	}
	return nil
}

func initGCPTracer(ctx context.Context, conf *TracerConfig) {
	exporter, err := texporter.New()
	if err != nil {
		log.Fatalc(ctx, err)
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter), sdktrace.WithSampler(sdktrace.TraceIDRatioBased(conf.TracerSample)))
	defer func() {
		_ = tp.ForceFlush(ctx)
	}()

	otel.SetTracerProvider(tp)

	log.Infofc(ctx, "tracer: initialized cloud trace with sample fraction: %g", conf.TracerSample)
}

func initJaegerTracer(conf *TracerConfig) io.Closer {
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: conf.TracerSample,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	closer, err := cfg.InitGlobalTracer(
		conf.Name,
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)

	if err != nil {
		log.Fatalf("Could not initialize jaeger tracer: %s\n", err.Error())
	}

	log.Infof("tracer: initialized jaeger tracer with sample fraction: %g\n", conf.TracerSample)
	return closer
}
