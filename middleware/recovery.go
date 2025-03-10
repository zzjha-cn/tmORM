package middleware

import (
	"fmt"
	"log"
	"tm_orm"
)

func Recovery(next tmorm.MHandlerFunc) tmorm.MHandlerFunc {
	return func(mctx *tmorm.MiddleCtx) (res tmorm.MResult) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Default().Println(rec)
				res.Err = fmt.Errorf("got a panic %s", rec)
				return
			}
		}()
		res = next(mctx)
		return
	}
}
