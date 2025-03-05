package utils

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

type TestStruct struct {
	Name    string  `bson:"name"`
	Age     int     `bson:"age"`
	Score   float64 `bson:"score"`
	Ignored string  `bson:"-"`
	Empty   string  `bson:"empty"`
}

type NestedStruct struct {
	Info    TestStruct `bson:"info"`
	Address string     `bson:"address"`
}

type PointerStruct struct {
	Info    *TestStruct `bson:"info"`
	Address *string     `bson:"address"`
}

func TestMakeBsonDByReflect(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		omitZero bool
		want     bson.D
		wantErr  bool
	}{
		{
			name: "基本结构体测试",
			input: TestStruct{
				Name:    "张三",
				Age:     25,
				Score:   98.5,
				Ignored: "ignored",
				Empty:   "",
			},
			omitZero: false,
			want: bson.D{
				{Key: "name", Value: "张三"},
				{Key: "age", Value: 25},
				{Key: "score", Value: 98.5},
				{Key: "empty", Value: ""},
			},
			wantErr: false,
		},
		{
			name: "omitZero为true时忽略零值",
			input: TestStruct{
				Name:    "张三",
				Age:     0,
				Score:   0,
				Ignored: "ignored",
				Empty:   "",
			},
			omitZero: true,
			want: bson.D{
				{Key: "name", Value: "张三"},
			},
			wantErr: false,
		},
		{
			name: "嵌套结构体测试",
			input: NestedStruct{
				Info: TestStruct{
					Name:  "张三",
					Age:   25,
					Score: 98.5,
				},
				Address: "北京",
			},
			omitZero: false,
			want: bson.D{
				{Key: "info", Value: bson.D{
					{Key: "name", Value: "张三"},
					{Key: "age", Value: 25},
					{Key: "score", Value: 98.5},
					{Key: "empty", Value: ""},
				}},
				{Key: "address", Value: "北京"},
			},
			wantErr: false,
		},
		{
			name: "指针结构体测试",
			input: &TestStruct{
				Name:  "张三",
				Age:   25,
				Score: 98.5,
			},
			omitZero: false,
			want: bson.D{
				{Key: "name", Value: "张三"},
				{Key: "age", Value: 25},
				{Key: "score", Value: 98.5},
				{Key: "empty", Value: ""},
			},
			wantErr: false,
		},
		{
			name:     "非结构体测试",
			input:    "string",
			omitZero: false,
			want:     nil,
			wantErr:  true,
		},
		{
			name: "带指针字段的结构体测试",
			input: func() PointerStruct {
				name := "张三"
				addr := "北京"
				return PointerStruct{
					Info: &TestStruct{
						Name:  name,
						Age:   25,
						Score: 98.5,
					},
					Address: &addr,
				}
			}(),
			omitZero: false,
			want: bson.D{
				{Key: "info", Value: bson.D{
					{Key: "name", Value: "张三"},
					{Key: "age", Value: 25},
					{Key: "score", Value: 98.5},
					{Key: "empty", Value: ""},
				}},
				{Key: "address", Value: "北京"},
			},
			wantErr: false,
		},
		{
			name: "带指针字段的结构体并存在nil的情况测试",
			input: func() PointerStruct {
				name := "张三"
				return PointerStruct{
					Info: &TestStruct{
						Name:  name,
						Age:   25,
						Score: 98.5,
					},
					Address: nil,
				}
			}(),
			omitZero: false,
			want: bson.D{
				{Key: "info", Value: bson.D{
					{Key: "name", Value: "张三"},
					{Key: "age", Value: 25},
					{Key: "score", Value: 98.5},
					{Key: "empty", Value: ""},
				}},
			},
			wantErr: false,
		},
		{
			name: "带指针字段的结构体并存在nil的情况测试",
			input: func() PointerStruct {
				name := "张三"
				return PointerStruct{
					Info:    nil,
					Address: &name,
				}
			}(),
			omitZero: false,
			want: bson.D{
				{Key: "address", Value: "张三"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeBsonDByReflect(tt.input, tt.omitZero)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeBsonDByReflect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeBsonDByReflect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTagString(t *testing.T) {
	type TestStruct struct {
		Normal   string `bson:"normal"`
		Empty    string `bson:""`
		Ignored  string `bson:"-"`
		Multiple string `bson:"multiple,omitempty"`
		NoTag    string
	}

	tests := []struct {
		name      string
		fieldName string
		want      string
		wantOk    bool
	}{
		{
			name:      "正常标签",
			fieldName: "Normal",
			want:      "normal",
			wantOk:    true,
		},
		{
			name:      "空标签",
			fieldName: "Empty",
			want:      "",
			wantOk:    false,
		},
		{
			name:      "忽略标签",
			fieldName: "Ignored",
			want:      "",
			wantOk:    false,
		},
		{
			name:      "多选项标签",
			fieldName: "Multiple",
			want:      "multiple",
			wantOk:    true,
		},
		{
			name:      "无标签",
			fieldName: "NoTag",
			want:      "",
			wantOk:    false,
		},
	}

	typ := reflect.TypeOf(TestStruct{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, _ := typ.FieldByName(tt.fieldName)
			got, ok := GetTagString(field, "bson")
			if ok != tt.wantOk {
				t.Errorf("GetTagString() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if got != tt.want {
				t.Errorf("GetTagString() = %v, want %v", got, tt.want)
			}
		})
	}
}
