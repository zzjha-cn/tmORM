package middleware

import (
	"time"
	"tm_orm"
)

type SLowQueryMiddleware struct {
	// 慢查询的阈值，毫秒单位
	Threshold int64
	Fn        func(ctx *tmorm.MiddleCtx)
}

func (m SLowQueryMiddleware) Build() tmorm.MHandlerBuilder {
	return func(next tmorm.MHandlerFunc) tmorm.MHandlerFunc {
		return func(mctx *tmorm.MiddleCtx) tmorm.MResult {
			var (
				start = time.Now()
			)

			defer func() {
				duration := time.Now().Sub(start)
				if m.Threshold > 0 && duration.Milliseconds() > m.Threshold {
					if m.Fn != nil {
						m.Fn(mctx)
					}
				}
				//var (
				//	q          = mctx.Query.GetBsonD()
				//	t          = mctx.Typ
				//	collection = mctx.Session.Collection
				//)
				//log.Default().Println("emongo_orm slow query %s", map[string]any{
				//	"query":   q,
				//	"method":  t,
				//	"collect": collection,
				//	"dur":     duration.Milliseconds(),
				//})
			}()
			return next(mctx)
		}
	}
}
