package middleware

import (
	"tm_orm"
)

func Recovery(next tmorm.MHandlerFunc) tmorm.MHandlerFunc {
	return func(mctx *tmorm.MiddleCtx) tmorm.MResult {
		defer func() {
			if rec := recover(); rec != nil {
				return
			}
		}()
		return next(mctx)
	}
}
