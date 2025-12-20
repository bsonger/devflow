package service

import "go.opentelemetry.io/otel"

const (
	ServiceName = "devflow"
)

var devflowTracer = otel.Tracer(ServiceName)
