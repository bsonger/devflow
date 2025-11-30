package otel

import (
	"context"
	"github.com/bsonger/devflow/pkg/logging"

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

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(tp)
	logging.Logger.Info("init otel tracer client")
	return nil
}

func Start(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer("devflow").Start(ctx, name)
}
