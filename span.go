package dgotel

import (
	"context"

	"github.com/darwinOrg/go-common/constants"
	dgctx "github.com/darwinOrg/go-common/context"
	dgsys "github.com/darwinOrg/go-common/sys"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func DoActionWithinNewSpan(ctx *dgctx.DgContext, spanName string, opts []trace.SpanStartOption, action func(c context.Context, span trace.Span)) {
	if Tracer == nil {
		return
	}

	var c context.Context
	if ctx.GetInnerContext() != nil {
		c = ctx.GetInnerContext()
	} else {
		c = context.Background()
	}

	if opts == nil {
		opts = []trace.SpanStartOption{}
	}
	nc, span := Tracer.Start(c, spanName, opts...)
	defer span.End()

	action(nc, span)
}

func GetSpanByDgContext(ctx *dgctx.DgContext) trace.Span {
	if ctx.GetInnerContext() == nil {
		return nil
	}

	spanContext := trace.SpanContextFromContext(ctx.GetInnerContext())
	if !spanContext.IsValid() || !spanContext.HasSpanID() {
		return nil
	}

	return trace.SpanFromContext(ctx.GetInnerContext())
}

func SetSpanAttributesByDgContext(ctx *dgctx.DgContext) {
	span := GetSpanByDgContext(ctx)
	if span == nil {
		return
	}

	var attrs []attribute.KeyValue

	profile := dgsys.GetProfile()
	if profile != "" {
		attrs = append(attrs, attribute.String(constants.Profile, profile))
	}
	if ctx.UserId > 0 {
		attrs = append(attrs, attribute.Int64(constants.UID, ctx.UserId))
	}
	if ctx.OpId > 0 {
		attrs = append(attrs, attribute.Int64(constants.OpId, ctx.OpId))
	}
	if ctx.RunAs > 0 {
		attrs = append(attrs, attribute.Int64(constants.RunAs, ctx.RunAs))
	}
	if ctx.Roles != "" {
		attrs = append(attrs, attribute.String(constants.Roles, ctx.Roles))
	}
	if ctx.BizTypes > 0 {
		attrs = append(attrs, attribute.Int(constants.BizTypes, ctx.BizTypes))
	}
	if ctx.GroupId > 0 {
		attrs = append(attrs, attribute.Int64(constants.GroupId, ctx.GroupId))
	}
	if ctx.Platform != "" {
		attrs = append(attrs, attribute.String(constants.Platform, ctx.Platform))
	}
	if ctx.CompanyId > 0 {
		attrs = append(attrs, attribute.Int64(constants.CompanyId, ctx.CompanyId))
	}
	if ctx.Product > 0 {
		attrs = append(attrs, attribute.Int(constants.Product, ctx.Product))
	}
	if len(ctx.Products) > 0 {
		attrs = append(attrs, attribute.IntSlice(constants.Products, ctx.Products))
	}
	if len(ctx.DepartmentIds) > 0 {
		attrs = append(attrs, attribute.Int64Slice(constants.Products, ctx.DepartmentIds))
	}

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}
}

func SetSpanAttributesByMap(ctx *dgctx.DgContext, mp map[string]string) {
	if len(mp) == 0 {
		return
	}

	span := GetSpanByDgContext(ctx)
	if span == nil {
		return
	}

	var attrs []attribute.KeyValue
	for k, v := range mp {
		attrs = append(attrs, attribute.String(k, v))
	}

	span.SetAttributes(attrs...)
}
