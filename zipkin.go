package zipkin

import (
	"context"
	"errors"

	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/server"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin-contrib/zipkin-go-opentracing"
)

// NewHandlerWrapper accepts an opentracing Tracer and returns a Handler Wrapper
func NewHandlerWrapper() server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			name := req.Service() + "." + req.Endpoint()
			ctx, span, err := StartSpanFromContext(ctx, name)
			if err != nil {
				return err
			}
			defer span.Finish()
			return h(ctx, req, rsp)
		}
	}
}

// StartSpanFromContext returns a new span with the given operation name and options. If a span
// is found in the context, it will be used as the parent of the resulting span.
func StartSpanFromContext(ctx context.Context, name string, opts ...opentracing.StartSpanOption) (context.Context, opentracing.Span, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(map[string]string)
	}

	// copy the metadata to prevent race
	md = metadata.Copy(md)

	// find trace in go-micro metadata
	if spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, opentracing.TextMapCarrier(md)); err == nil {
		opts = append(opts, opentracing.ChildOf(spanCtx))
	}

	// find span context in opentracing library
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
	}

	sp := opentracing.GlobalTracer().StartSpan(name, opts...)

	if err := sp.Tracer().Inject(sp.Context(), opentracing.TextMap, opentracing.TextMapCarrier(md)); err != nil {
		return nil, nil, err
	}

	ctx = opentracing.ContextWithSpan(ctx, sp)
	ctx = metadata.NewContext(ctx, md)
	return ctx, sp, nil
}

// SetGlobalTracer sets the [singleton] opentracing.Tracer returned by
// GlobalTracer(). Those who use GlobalTracer (rather than directly manage an
// opentracing.Tracer instance) should call SetGlobalTracer as early as
// possible in main(), prior to calling the `StartSpan` global func below.
// Prior to calling `SetGlobalTracer`, any Spans started via the `StartSpan`
// (etc) globals are noops.
func SetGlobalTracer(serviceName string, serviceAddress string, zipkinURL string) error {
	collector, err := zipkin.NewHTTPCollector(zipkinURL)
	if err != nil {
		return errors.New("Failed to create zipkin collecter: " + err.Error())
	}
	recorder := zipkin.NewRecorder(collector, false, serviceAddress, serviceName)
	tracer, err := zipkin.NewTracer(recorder, zipkin.ClientServerSameSpan(true), zipkin.TraceID128Bit(true))
	if err != nil {
		return errors.New("Failed to crate zipkin tracer: " + err.Error())
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}
