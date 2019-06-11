package luatable

import (
	"testing"

	"encoding/json"
	"github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	//t:=test.T
	assert := assert.New(t)
	var result interface{}
	for _, tc := range tests {
		// tc := tests[0]
		var err error
		err = Unmarshal(tc.data, &result)
		if tc.err != "" {
			if assert.Error(err, tc.name) {
				assert.Equal(tc.err, err.Error(), tc.name)
			}
		} else {
			if assert.NoError(err, tc.name) {
				str, err := json.Marshal(result)
				if assert.NoError(err, tc.name) {
					t.Logf("result is %v", result)
					assert.Equal(tc.want, string(str), tc.name)
				}
			}
		}
	}

}

func xTestDecodeStruct(t *testing.T) {
	assert := assert.New(t)
	var result testStruct
	tc := tests[0]
	var err error
	err = Unmarshal(tc.data, &result)
	if tc.err != "" {
		if assert.Error(err, tc.name) {
			assert.Equal(tc.err, err.Error(), tc.name)
		}
	} else {
		if assert.NoError(err, tc.name) {
			str, err := json.Marshal(result)
			if assert.NoError(err, tc.name) {
				t.Logf("result is %v", result)
				assert.Equal(tc.want, string(str), tc.name)
			}
		}
	}
}

func BenchmarkUmarshalLuaToInterface(b *testing.B) {
	var v interface{}
	var err error
	src := tests[0].data
	for i := 0; i < b.N; i++ {
		if err = Unmarshal(src, &v); err != nil {
			b.Error(err.Error())
		}
	}
}

type testData struct {
	ID   string
	Name string
}
type testStruct struct {
	Config1 []testData
	Config2 map[string]testData
}

func xBenchmarkUmarshalLuaToStruct(b *testing.B) {
	var v testStruct
	var err error
	src := tests[0].data
	for i := 0; i < b.N; i++ {
		if err = Unmarshal(src, &v); err != nil {
			b.Error(err.Error())
		}
	}
}

func BenchmarkUmarshalJSONIter(b *testing.B) {
	var v interface{}
	var err error
	src := []byte(tests[0].want)
	jsondec := jsoniter.ConfigCompatibleWithStandardLibrary
	for i := 0; i < b.N; i++ {
		if err = jsondec.Unmarshal(src, &v); err != nil {
			b.Error(err.Error())
		}
	}
}

func BenchmarkUmarshalJSON(b *testing.B) {
	var v interface{}
	var err error
	src := []byte(tests[0].want)
	for i := 0; i < b.N; i++ {
		if err = json.Unmarshal(src, &v); err != nil {
			b.Error(err.Error())
		}
	}
}

var (
	tests = []struct {
		name string
		data string
		want string
		err  string
	}{
		{ // 简单测试
			name: "简单测试",
			data: `{Config1 = {
				{
					id=1,
					name="name1",
				},
				{
					id=2,
					name="name2",
				}
			},
			Config2 = {
				name1 = {
					id=1,
					name="name1",
				},
				name2 = {
					id=2,
					name="name2",
				}
			}}
`,
			want: `{"Config1":[{"id":1,"name":"name1"},{"id":2,"name":"name2"}],"Config2":{"name1":{"id":1,"name":"name1"},"name2":{"id":2,"name":"name2"}}}`,
		},
		{ // 双key
			name: "双key测试",
			data: `{test={1001={name1={/*1001*/id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},name2={/*1002*/id=1001,name="name2",type=2,icon="icon/head1002",attack=11,life=101}},1003={name3={/*1003*/id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102}},1004={name4={/*1004*/id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}}}
`,
			want: `{"test":{"1001":{"name1":{"attack":10,"icon":"icon/head1001","id":1001,"life":100,"name":"name1","type":1},"name2":{"attack":11,"icon":"icon/head1002","id":1001,"life":101,"name":"name2","type":2}},"1003":{"name3":{"attack":12,"icon":"icon/head1003","id":1003,"life":102,"name":"name3","type":3}},"1004":{"name4":{"attack":13,"icon":"icon/head1004","id":1004,"life":103,"name":"name4","type":4}}}}`,
		},
		{ // 双key
			name: "双key测试中间插注释",
			data: `{test={1001={1={/*1001*/id=1001,/*name1*/type=1,icon="icon/head1001",attack=10,life=100}}}}
`,
			want: `{"test":{"1001":{"1":{"attack":10,"icon":"icon/head1001","id":1001,"life":100,"type":1}}}}`,
		},
		{ // 单key
			name: "单key测试",
			data: `{test={1001={/*1001*/id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},1002={/*1002*/id=1002,name="name2",type=2,icon="icon/head1002",attack=11,life=101},1003={/*1003*/id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102},1004={/*1004*/id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}}
`,
			want: `{"test":{"1001":{"attack":10,"icon":"icon/head1001","id":1001,"life":100,"name":"name1","type":1},"1002":{"attack":11,"icon":"icon/head1002","id":1002,"life":101,"name":"name2","type":2},"1003":{"attack":12,"icon":"icon/head1003","id":1003,"life":102,"name":"name3","type":3},"1004":{"attack":13,"icon":"icon/head1004","id":1004,"life":103,"name":"name4","type":4}}}`,
		},
		{ // 单key
			name: "数字为key测试",
			data: `{test={[1001]={/*1001*/id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},[1002]={/*1002*/id=1002,name="name2",type=2,icon="icon/head1002",attack=11,life=101},[1003]={/*1003*/id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102},[1004]={/*1004*/id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}}
`,
			want: `{"test":{"1001":{"attack":10,"icon":"icon/head1001","id":1001,"life":100,"name":"name1","type":1},"1002":{"attack":11,"icon":"icon/head1002","id":1002,"life":101,"name":"name2","type":2},"1003":{"attack":12,"icon":"icon/head1003","id":1003,"life":102,"name":"name3","type":3},"1004":{"attack":13,"icon":"icon/head1004","id":1004,"life":103,"name":"name4","type":4}}}`,
		},
		{ // 单key
			name: "数字为key测试",
			data: `{[222]={[1001]={/*1001*/id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},[1002]={/*1002*/id=1002,name="name2",type=2,icon="icon/head1002",attack=11,life=101},[1003]={/*1003*/id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102},[1004]={/*1004*/id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}}
`,
			want: `{"222":{"1001":{"attack":10,"icon":"icon/head1001","id":1001,"life":100,"name":"name1","type":1},"1002":{"attack":11,"icon":"icon/head1002","id":1002,"life":101,"name":"name2","type":2},"1003":{"attack":12,"icon":"icon/head1003","id":1003,"life":102,"name":"name3","type":3},"1004":{"attack":13,"icon":"icon/head1004","id":1004,"life":103,"name":"name4","type":4}}}`,
		},
		{ // 数组
			name: "数组测试",
			data: `{test={{/*1001*/id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},{/*1002*/id=1001,name="name2",type=2,icon="icon/head1002",attack=11,life=101},{/*1003*/id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102},{/*1004*/id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}}
`,
			want: `{"test":[{"attack":10,"icon":"icon/head1001","id":1001,"life":100,"name":"name1","type":1},{"attack":11,"icon":"icon/head1002","id":1001,"life":101,"name":"name2","type":2},{"attack":12,"icon":"icon/head1003","id":1003,"life":102,"name":"name3","type":3},{"attack":13,"icon":"icon/head1004","id":1004,"life":103,"name":"name4","type":4}]}`,
		},
		{ // 数组
			name: "纯数组测试",
			data: `{{/*1001*/id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},{/*1002*/id=1001,name="name2",type=2,icon="icon/head1002",attack=11,life=101},{/*1003*/id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102}{/*1004*/id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}
`,
			want: `[{"attack":10,"icon":"icon/head1001","id":1001,"life":100,"name":"name1","type":1},{"attack":11,"icon":"icon/head1002","id":1001,"life":101,"name":"name2","type":2},{"attack":12,"icon":"icon/head1003","id":1003,"life":102,"name":"name3","type":3},{"attack":13,"icon":"icon/head1004","id":1004,"life":103,"name":"name4","type":4}]`,
		},
		{ // 数组
			name: "数组少}测试",
			data: `{test={{/*1001*/id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},{/*1002*/id=1001,name="name2",type=2,icon="icon/head1002",attack=11,life=101},{/*1003*/id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102,{/*1004*/id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}}
`,
			want: ``,
			err:  `位于{id=1004name=前面:缺少}号`,
		},
		{ // 数组
			name: "数组少,测试",
			data: `{test={{/*1001*/id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},{/*1002*/id=1001,name="name2",type=2,icon="icon/head1002",attack=11,life=101},{/*1003*/id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102}{/*1004*/id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}}
`,
			want: `{"test":[{"attack":10,"icon":"icon/head1001","id":1001,"life":100,"name":"name1","type":1},{"attack":11,"icon":"icon/head1002","id":1001,"life":101,"name":"name2","type":2},{"attack":12,"icon":"icon/head1003","id":1003,"life":102,"name":"name3","type":3},{"attack":13,"icon":"icon/head1004","id":1004,"life":103,"name":"name4","type":4}]}`,
		},
		{ // 双key
			name: "双key测试括号封闭不正常",
			data: `{test={}1001={1={/*1001*/id=1001,/*name1*/type=1,icon="icon/head1001",attack=10,life=100}}}}
`,
			err: `位于}前面:结尾字符过多，检查前面的括号封闭`,
		},
	}
)
