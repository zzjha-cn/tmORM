package tmorm

import (
	"tm_orm/impl"
)

// 调用流程中间件，负责状态过滤与结果集
// 因为框架没有统一函数调用签名，所以中间件的输入需要构建中间状态

// 需要中间件的流程：
// 获取连接的前后、发起查询之前，完成查询之后

type (
	MethodTyp string
)
type (
	// MiddleHandler struct {
	// 	Name string
	// 	Fn   MHandlerFunc
	// 	ind  int
	// }
	//MiddleChain struct {
	//	chain       []*MiddleHandler
	//	beforeConn  []int
	//	afterConn   []int
	//	beforeQuery []int
	//	afterQuery  []int
	//}

	// 核心中间件逻辑
	MHandlerFunc func(mctx *MiddleCtx) MResult

	// 中间件构造器，辅助链式构造
	MHandlerBuilder func(next MHandlerFunc) MHandlerFunc

	// 负责所用中间件的管理，内聚功能
	MiddleChain struct {
		chain []MHandlerBuilder
	}

	// 中间流程上下文
	MiddleCtx struct {
		Typ     MethodTyp
		Session *MSession
		Query   impl.IBsonQuery
		Update  impl.IUpdateBuilder
		Upsert  impl.IUpsertBuilder
	}

	MResult struct {
		Val any
		Err error
	}
)

func NewMiddleChainAdapt() *MiddleChain {
	return &MiddleChain{}
}

func NewMiddleCtx(s *MSession, typ MethodTyp) *MiddleCtx {
	c := &MiddleCtx{}
	c.Typ = typ
	c.Session = s
	return c
}

func (c *MiddleChain) Use(msList ...MHandlerBuilder) *MiddleChain {
	c.chain = msList
	return c
}

func (c *MiddleChain) build(root MHandlerFunc, sessBuilder ...MHandlerBuilder) MHandlerFunc {
	res := root
	for _, m := range c.chain {
		res = m(res)
	}

	for _, m := range sessBuilder {
		res = m(res)
	}

	return res
}

//func (c *MiddleChain) addMs(msList ...MHandlerFunc) (*MiddleChain, []int) {
//	var (
//		ind     = len(c.chain)
//		indList = make([]int, 0, len(msList))
//	)
//
//	for _, m := range msList {
//		c.chain = append(c.chain, &MiddleHandler{
//			ind: ind,
//			Fn:  m,
//		})
//		indList = append(indList, ind)
//		ind++
//	}
//	return c, indList
//}
