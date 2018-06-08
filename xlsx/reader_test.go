package xlsx

import (
	"testing"

	excel "github.com/360EntSecGroup-Skylar/excelize"
	"github.com/stretchr/testify/assert"
)

var (
	tests = []struct {
		name     string
		data     [][]string
		keyCount int
		want     string
	}{
		{ // 双key
			name:     "双key测试",
			data:     [][]string{{"备注", "id", "名称", "类型", "头像", "攻击", "生命"}, {"", "sc", "sc", "sc", "c", "s", "s"}, {"comment", "int", "string", "int", "string", "int", "int"}, {"", "id", "name", "type", "icon", "attack", "life"}, {"1001", "1001", "名字1", "1", "icon/head1001", "10", "100"}, {"1002", "1001", "名字2", "2", "icon/head1002", "11", "101"}, {"1003", "1003", "名字3", "3", "icon/head1003", "12", "102"}, {"1004", "1004", "名字4", "4", "icon/head1004", "13", "103"}},
			keyCount: 2,
			want: `test={1001={名字1={/*1001*/id=1001,name="名字1",type=1,icon="icon/head1001",attack=10,life=100},名字2={/*1002*/id=1001,name="名字2",type=2,icon="icon/head1002",attack=11,life=101}},1003={名字3={/*1003*/id=1003,name="名字3",type=3,icon="icon/head1003",attack=12,life=102}},1004={名字4={/*1004*/id=1004,name="名字4",type=4,icon="icon/head1004",attack=13,life=103}}}
`,
		},
		{ // 双key
			name:     "双key测试中间插注释",
			data:     [][]string{{"备注", "id", "名称", "类型", "头像", "攻击", "生命"}, {"", "sc", "sc", "sc", "c", "s", "s"}, {"comment", "int", "comment", "int", "string", "int", "int"}, {"", "id", "name", "type", "icon", "attack", "life"}, {"1001", "1001", "名字1", "1", "icon/head1001", "10", "100"}, {"1002", "1001", "名字2", "2", "icon/head1002", "11", "101"}, {"1003", "1003", "名字3", "3", "icon/head1003", "12", "102"}, {"1004", "1004", "名字4", "4", "icon/head1004", "13", "103"}},
			keyCount: 2,
			want: `test={1001={1={/*1001*/id=1001,/*名字1*/type=1,icon="icon/head1001",attack=10,life=100},2={/*1002*/id=1001,/*名字2*/type=2,icon="icon/head1002",attack=11,life=101}},1003={3={/*1003*/id=1003,/*名字3*/type=3,icon="icon/head1003",attack=12,life=102}},1004={4={/*1004*/id=1004,/*名字4*/type=4,icon="icon/head1004",attack=13,life=103}}}
`,
		},
		{ // 单key
			name:     "单key测试",
			data:     [][]string{{"备注", "id", "名称", "类型", "头像", "攻击", "生命"}, {"", "sc", "sc", "sc", "c", "s", "s"}, {"comment", "int", "string", "int", "string", "int", "int"}, {"", "id", "name", "type", "icon", "attack", "life"}, {"1001", "1001", "名字1", "1", "icon/head1001", "10", "100"}, {"1002", "1002", "名字2", "2", "icon/head1002", "11", "101"}, {"1003", "1003", "名字3", "3", "icon/head1003", "12", "102"}, {"1004", "1004", "名字4", "4", "icon/head1004", "13", "103"}},
			keyCount: 1,
			want: `test={1001={/*1001*/id=1001,name="名字1",type=1,icon="icon/head1001",attack=10,life=100},1002={/*1002*/id=1002,name="名字2",type=2,icon="icon/head1002",attack=11,life=101},1003={/*1003*/id=1003,name="名字3",type=3,icon="icon/head1003",attack=12,life=102},1004={/*1004*/id=1004,name="名字4",type=4,icon="icon/head1004",attack=13,life=103}}
`,
		},
		{ // 数组
			name:     "数组测试",
			data:     [][]string{{"备注", "id", "名称", "类型", "头像", "攻击", "生命"}, {"", "sc", "sc", "sc", "c", "s", "s"}, {"comment", "int", "string", "int", "string", "int", "int"}, {"", "id", "name", "type", "icon", "attack", "life"}, {"1001", "1001", "名字1", "1", "icon/head1001", "10", "100"}, {"1002", "1001", "名字2", "2", "icon/head1002", "11", "101"}, {"1003", "1003", "名字3", "3", "icon/head1003", "12", "102"}, {"1004", "1004", "名字4", "4", "icon/head1004", "13", "103"}},
			keyCount: 0,
			want: `test={{/*1001*/id=1001,name="名字1",type=1,icon="icon/head1001",attack=10,life=100},{/*1002*/id=1001,name="名字2",type=2,icon="icon/head1002",attack=11,life=101},{/*1003*/id=1003,name="名字3",type=3,icon="icon/head1003",attack=12,life=102},{/*1004*/id=1004,name="名字4",type=4,icon="icon/head1004",attack=13,life=103}}
`,
		},
	}
)

func TestOpenXLSXFile(t *testing.T) {
	assert := assert.New(t)
	xlsfile, err := excel.OpenFile("test.xlsx")
	assert.Nilf(err, "error opening excel:%v", err)

	sname := xlsfile.GetSheetName(1)
	data := xlsfile.GetRows(sname)
	if !assert.Truef(len(data) > 7, "非法格式的excel文件") {
		t.Log(data[3 : len(data)-1])
		return
	}
}

func TestXLSXReader(t *testing.T) {
	assert := assert.New(t)

	for _, testcase := range tests {
		r := New("test", testcase.data, "s", testcase.keyCount, 1, 3, 2, 4)
		result, err := r.ReadAll()
		assert.Nilf(err, "error reading content:%v", err)
		assert.Equalf(testcase.want, string(result), "%v 出错", testcase.name)
	}
}
