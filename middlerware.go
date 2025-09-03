package tmorm

import (
	"context"
	"tm_orm/impl"
)

type (
	MethodTyp string
)
type (
	MiddlewareFunc func(mctx *MiddleCtx, next func(m *MiddleCtx))

	// 中间流程上下文
	MiddleCtx struct {
		UserCtx context.Context
		Typ     MethodTyp
		Query   impl.IBsonQuery
		Update  impl.IUpdateBuilder
		Upsert  impl.IUpsertBuilder
		Result  *MResult
	}

	MResult struct {
		Val any
		Err error
	}
)

func NewMiddleContext(ctx context.Context, typ MethodTyp) *MiddleCtx {
	return &MiddleCtx{
		UserCtx: ctx,
		Typ:     typ,
		Result:  &MResult{},
	}
}

func Executor(mctx *MiddleCtx, list []MiddlewareFunc) {
	var et func(i int, ctx *MiddleCtx)
	et = func(i int, ctx *MiddleCtx) {
		if i >= len(list) {
			return
		}
		md := list[i]
		md(mctx, func(m *MiddleCtx) {
			et(i+1, m)
		})
	}
	et(0, mctx)
}
