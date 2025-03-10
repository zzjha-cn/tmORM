package test

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"tm_orm/query"
)

func TestGroupCmd(t *testing.T) {
	testCase := []struct {
		name   string
		gc     *query.GroupCmd
		before func(gc *query.GroupCmd)
		check  func(gc *query.GroupCmd) error
	}{
		{
			name: "测试Raw方法",
			gc:   query.NewGroupCmd(),
			before: func(gc *query.GroupCmd) {
				*gc = gc.Id(nil).Raw(bson.D{
					{"customField", bson.D{{"$sum", bson.A{1}}}},
					{"otherField", "value"},
				})
			},
			check: func(gc *query.GroupCmd) error {
				want := bson.D{
					{"_id", nil},
					{"customField", bson.D{{"$sum", bson.A{1}}}},
					{"otherField", "value"},
				}
				d := gc.Build().GetData()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试空字段名和空值",
			gc:   query.NewGroupCmd(),
			before: func(gc *query.GroupCmd) {
				*gc = gc.IdWithField("").Key("").Sum(gc.ToFd(""))
			},
			check: func(gc *query.GroupCmd) error {
				want := bson.D{
					{"_id", "$"},
					{"", bson.D{{"$sum", bson.A{"$"}}}},
				}
				d := gc.Build().GetData()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试多个参数的聚合函数",
			gc:   query.NewGroupCmd(),
			before: func(gc *query.GroupCmd) {
				*gc = gc.Id(nil).Key("total").Sum(
					gc.ToFd("value1"),
					gc.ToFd("value2"),
					gc.AnyVal(10),
				)
			},
			check: func(gc *query.GroupCmd) error {
				want := bson.D{
					{"_id", nil},
					{"total", bson.D{{"$sum", bson.A{"$value1", "$value2", 10}}}},
				}
				d := gc.Build().GetData()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试基本分组操作",
			gc:   query.NewGroupCmd(),
			before: func(gc *query.GroupCmd) {
				// 测试基本的分组和聚合操作
				*gc = gc.IdWithField("department").
					Key("totalSalary").Sum(gc.ToFd("salary")).
					Key("avgAge").Avg(gc.ToFd("age")).
					Key("employeeCount").Sum(gc.AnyVal(1))
			},
			check: func(gc *query.GroupCmd) error {
				want := bson.D{
					{"_id", "$department"},
					{"totalSalary", bson.D{{"$sum", bson.A{"$salary"}}}},
					{"avgAge", bson.D{{"$avg", bson.A{"$age"}}}},
					{"employeeCount", bson.D{{"$sum", bson.A{1}}}},
				}
				d := gc.Build().GetData()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试复杂聚合函数组合",
			gc:   query.NewGroupCmd(),
			before: func(gc *query.GroupCmd) {
				// 测试多个聚合函数的组合使用
				*gc = gc.Id(nil).
					Key("maxSalary").Max(gc.ToFd("salary")).
					Key("minSalary").Min(gc.ToFd("salary")).
					Key("uniqueDepts").AddToSet(gc.ToFd("department")).
					Key("firstEmployee").First(gc.ToFd("name")).
					Key("lastEmployee").Last(gc.ToFd("name")).
					Key("allNames").Push(gc.ToFd("name"))
			},
			check: func(gc *query.GroupCmd) error {
				want := bson.D{
					{"_id", nil},
					{"maxSalary", bson.D{{"$max", bson.A{"$salary"}}}},
					{"minSalary", bson.D{{"$min", bson.A{"$salary"}}}},
					{"uniqueDepts", bson.D{{"$addToSet", "$department"}}},
					{"firstEmployee", bson.D{{"$first", "$name"}}},
					{"lastEmployee", bson.D{{"$last", "$name"}}},
					{"allNames", bson.D{{"$push", "$name"}}},
				}
				d := gc.Build().GetData()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试复杂Id构建",
			gc:   query.NewGroupCmd(),
			before: func(gc *query.GroupCmd) {
				// 测试多个聚合函数的组合使用
				*gc = gc.Id(
					gc.IdBuilder().Key("year").Year(query.F("created_at")).
						Key("month").Month(query.F("created_at")).
						Key("date").DateToString(
						query.F("created_at"),
						"%Y-%m-%d",
						"Asia/Shanghai",
						nil,
					)).
					Key("allNames").Push(gc.ToFd("name"))
			},
			check: func(gc *query.GroupCmd) error {
				want := bson.D{
					{"_id", bson.D{
						{"year", bson.D{{"$year", "$created_at"}}},
						{"month", bson.D{{"$month", "$created_at"}}},
						{"date", bson.D{{"$dateToString", bson.D{
							{"date", "$created_at"},
							{"format", "%Y-%m-%d"},
							{"timezone", "Asia/Shanghai"},
						}}}}}},
					{"allNames", bson.D{{"$push", "$name"}}},
				}
				d := gc.Build().GetData()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试复杂Id构建 Cond与数组",
			gc:   query.NewGroupCmd(),
			before: func(gc *query.GroupCmd) {
				// 测试多个聚合函数的组合使用
				*gc = gc.Id(
					gc.IdBuilder().Key("firstTag").ArrayElemAt(
						query.F("tags"),
						query.V(0),
					).Key("status").Cond(
						query.F("is_active"),
						query.V("active"),
						query.V("inactive"),
					)).
					Key("allNames").Push(gc.ToFd("name"))
			},
			check: func(gc *query.GroupCmd) error {
				want := bson.D{
					{"_id", bson.D{
						{"firstTag", bson.D{{"$arrayElemAt", bson.A{"$tags", 0}}}},
						{"status", bson.D{{"$cond", bson.A{"$is_active", "active", "inactive"}}}},
					}},
					{"allNames", bson.D{{"$push", "$name"}}},
				}
				d := gc.Build().GetData()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "",
			gc:   query.NewGroupCmd(),
			before: func(gc *query.GroupCmd) {
				// 测试多个聚合函数的组合使用
				*gc = gc.Id(
					gc.IdBuilder().Key("year").Year(query.F("created_at")).
						Key("month").Month(query.F("created_at")).
						Key("date").DateToString(
						query.F("created_at"),
						"%Y-%m-%d",
						"Asia/Shanghai",
						nil,
					)).
					Key("allNames").Push(gc.ToFd("name"))
			},
			check: func(gc *query.GroupCmd) error {
				want := bson.D{
					{"_id", bson.D{
						{"year", bson.D{{"$year", "$created_at"}}},
						{"month", bson.D{{"$month", "$created_at"}}},
						{"date", bson.D{{"$dateToString", bson.D{
							{"date", "$created_at"},
							{"format", "%Y-%m-%d"},
							{"timezone", "Asia/Shanghai"},
						}}}}}},
					{"allNames", bson.D{{"$push", "$name"}}},
				}
				d := gc.Build().GetData()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试复杂Id构建 Cond与数组",
			gc:   query.NewGroupCmd(),
			before: func(gc *query.GroupCmd) {
				// 测试多个聚合函数的组合使用
				*gc = gc.Id(
					gc.IdBuilder().Key("firstTag").ArrayElemAt(
						query.F("tags"),
						query.V(0),
					).Key("status").Cond(
						query.F("is_active"),
						query.V("active"),
						query.V("inactive"),
					)).
					Key("allNames").Push(gc.ToFd("name"))
			},
			check: func(gc *query.GroupCmd) error {
				want := bson.D{
					{"_id", bson.D{
						{"firstTag", bson.D{{"$arrayElemAt", bson.A{"$tags", 0}}}},
						{"status", bson.D{{"$cond", bson.A{"$is_active", "active", "inactive"}}}},
					}},
					{"allNames", bson.D{{"$push", "$name"}}},
				}
				d := gc.Build().GetData()
				assert.Equal(t, want, d)
				return nil
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			if tc.before != nil {
				tc.before(tc.gc)
			}

			if tc.check != nil {
				if err := tc.check(tc.gc); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
