package otel

import (
	"context"
	"fmt"
	"github.com/bsonger/devflow/pkg/logging"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"os"

	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func Init(endpoint, service string) error {
	exporter, _ := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)

	res, _ := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(service),
			semconv.ServiceNamespace("app"),
			semconv.DeploymentEnvironmentName(os.Getenv("env")),
		),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	logging.Logger.Info(fmt.Sprintf("OpenTelemetry tracing enabled, service name: %s", service))
	return nil
}

func Start(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer("devflow").Start(ctx, name)
}
