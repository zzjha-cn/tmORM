package query

//panic: (BadValue) $and must be an array [recovered]

type (
	QueryAnd struct {
		mc mongoCmd
	}

	QueryOr struct {
		mc mongoCmd
	}
)

func newQueryAnd() *QueryAnd {
	return &QueryAnd{}
}

func (q *QueryAnd) K(k string) mongoCmd {
	q.mc.e.SetKey(k)
	return q.mc
}

func (q *QueryAnd) Expr(f func(m MExpr) Builder) Builder {
	return q.mc.b.Expr(f)
}

// =============================================

func newQueryOr() *QueryOr {
	return &QueryOr{}
}

func (q *QueryOr) K(k string) mongoCmd {
	q.mc.e.SetKey(k)
	return q.mc
}

func (q *QueryOr) Expr(f func(m MExpr) Builder) Builder {
	return q.mc.b.Expr(f)
}
