package collection

import (
	"context"
	"errors"
	"testing"

	tmorm "github.com/zzjha-cn/tm_orm"
)

// 中间件短路（例如熔断器打开）时 Count 不能 panic
func TestCountDoesNotPanicWhenShortCircuited(t *testing.T) {
	shortCircuit := errors.New("circuit breaker is open")

	cli := (&tmorm.ORMClient{}).Use(func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
		mctx.Result.SetError(shortCircuit, tmorm.ErrorTypeConnection, "CIRCUIT_BREAKER_OPEN", "rejected")
		// 故意不调用 next，模拟中间件短路
	})

	count, err := NewCollection[filterTestDoc](cli, "testdb", "testcoll").Count(context.Background())
	if !errors.Is(err, shortCircuit) {
		t.Fatalf("期望返回短路错误，实际 %v", err)
	}
	if count != 0 {
		t.Fatalf("短路时计数应为 0，实际 %d", count)
	}
}
