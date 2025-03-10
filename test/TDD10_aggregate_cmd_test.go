package test

import (
	"testing"
	"tm_orm/aggregator"
)

func TestAggregateCmd(t *testing.T) {
	testCase := []struct {
		name string

		bd *aggregator.Aggregator[TestUser]

		before func(bd *aggregator.Aggregator[TestUser])
		check  func(bd *aggregator.Aggregator[TestUser]) error
	}{
		{
			name: "测试SetObj",
		},
	}

	for _, t1 := range testCase {
		t.Run(t1.name, func(t *testing.T) {
			if t1.before != nil {
				t1.before(t1.bd)
			}

			if t1.check != nil {
				if err := t1.check(t1.bd); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
