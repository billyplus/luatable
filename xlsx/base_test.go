package xlsx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	excel "github.com/tealeg/xlsx"
)

var (
	tests = []struct {
		name     string
		data     [][]string
		keyCount int
		wants    string
		wantc    string
	}{
		{ // 双key
			name:     "双key测试",
			data:     [][]string{{"备注", "id", "名称", "类型", "头像", "攻击", "生命"}, {"", "sc", "sc", "sc", "c", "s", "s"}, {"comment", "int", "string", "int", "string", "int", "int"}, {"", "id", "name", "type", "icon", "attack", "life"}, {"1001", "1001", "名字1", "1", "icon/head1001", "10", "100"}, {"1002", "1001", "名字2", "2", "icon/head1002", "11", "101"}, {"1003", "1003", "名字3", "3", "icon/head1003", "12", "102"}, {"1004", "1004", "名字4", "4", "icon/head1004", "13", "103"}},
			keyCount: 2,
			wants: `{1001={名字1={id=1001,name="名字1",type=1,attack=10,life=100},名字2={id=1001,name="名字2",type=2,attack=11,life=101}},1003={名字3={id=1003,name="名字3",type=3,attack=12,life=102}},1004={名字4={id=1004,name="名字4",type=4,attack=13,life=103}}}
`,
			wantc: `{1001={名字1={id=1001,name="名字1",type=1,icon="icon/head1001",},名字2={id=1001,name="名字2",type=2,icon="icon/head1002",}},1003={名字3={id=1003,name="名字3",type=3,icon="icon/head1003",}},1004={名字4={id=1004,name="名字4",type=4,icon="icon/head1004",}}}
`,
		},
		{ // 双key
			name:     "双key测试中间插注释",
			data:     [][]string{{"备注", "id", "名称", "类型", "头像", "攻击", "生命"}, {"", "sc", "sc", "sc", "c", "s", "s"}, {"comment", "int", "comment", "int", "string", "int", "int"}, {"", "id", "name", "type", "icon", "attack", "life"}, {"1001", "1001", "名字1", "1", "icon/head1001", "10", "100"}, {"1002", "1001", "名字2", "2", "icon/head1002", "11", "101"}, {"1003", "1003", "名字3", "3", "icon/head1003", "12", "102"}, {"1004", "1004", "名字4", "4", "icon/head1004", "13", "103"}},
			keyCount: 2,
			wants: `{1001={1={id=1001,type=1,attack=10,life=100},2={id=1001,type=2,attack=11,life=101}},1003={3={id=1003,type=3,attack=12,life=102}},1004={4={id=1004,type=4,attack=13,life=103}}}
`,
			wantc: `{1001={1={id=1001,type=1,icon="icon/head1001",},2={id=1001,type=2,icon="icon/head1002",}},1003={3={id=1003,type=3,icon="icon/head1003",}},1004={4={id=1004,type=4,icon="icon/head1004",}}}
`,
		},
		{ // 单key
			name:     "单key测试",
			data:     [][]string{{"备注", "id", "名称", "类型", "头像", "攻击", "生命"}, {"", "sc", "sc", "sc", "c", "s", "s"}, {"comment", "int", "string", "int", "string", "int", "int"}, {"", "id", "name", "type", "icon", "attack", "life"}, {"1001", "1001", "名字1", "1", "icon/head1001", "10", "100"}, {"1002", "1002", "名字2", "2", "icon/head1002", "11", "101"}, {"1003", "1003", "名字3", "3", "icon/head1003", "12", "102"}, {"1004", "1004", "名字4", "4", "icon/head1004", "13", "103"}},
			keyCount: 1,
			wants: `{1001={id=1001,name="名字1",type=1,attack=10,life=100},1002={id=1002,name="名字2",type=2,attack=11,life=101},1003={id=1003,name="名字3",type=3,attack=12,life=102},1004={id=1004,name="名字4",type=4,attack=13,life=103}}
`,
			wantc: `{1001={id=1001,name="名字1",type=1,icon="icon/head1001",},1002={id=1002,name="名字2",type=2,icon="icon/head1002",},1003={id=1003,name="名字3",type=3,icon="icon/head1003",},1004={id=1004,name="名字4",type=4,icon="icon/head1004",}}
`,
		},
		{ // 数组
			name:     "数组测试",
			data:     [][]string{{"备注", "id", "名称", "类型", "头像", "攻击", "生命"}, {"", "sc", "sc", "sc", "c", "s", "s"}, {"comment", "int", "string", "int", "string", "int", "int"}, {"", "id", "name", "type", "icon", "attack", "life"}, {"1001", "1001", "名字1", "1", "icon/head1001", "10", "100"}, {"1002", "1001", "名字2", "2", "icon/head1002", "11", "101"}, {"1003", "1003", "名字3", "3", "icon/head1003", "12", "102"}, {"1004", "1004", "名字4", "4", "icon/head1004", "13", "103"}},
			keyCount: 0,
			wants: `{{id=1001,name="名字1",type=1,attack=10,life=100},{id=1001,name="名字2",type=2,attack=11,life=101},{id=1003,name="名字3",type=3,attack=12,life=102},{id=1004,name="名字4",type=4,attack=13,life=103}}
`,
			wantc: `{{id=1001,name="名字1",type=1,icon="icon/head1001",},{id=1001,name="名字2",type=2,icon="icon/head1002",},{id=1003,name="名字3",type=3,icon="icon/head1003",},{id=1004,name="名字4",type=4,icon="icon/head1004",}}
`,
		},
	}
)

var datawant = `{{id=1001,name="name1",type=1,attack=10,life=100},{id=1001,name="name2",type=2,attack=11,life=101},{id=1003,name="name3",type=3,attack=12,life=102},{id=1004,name="name4",type=4,attack=13,life=103}}
`

func TestOpenXLSXFile(t *testing.T) {
	assert := assert.New(t)
	xlsfile, err := excel.OpenFile("test.xlsx")
	assert.Nilf(err, "error opening excel:%v", err)

	sh := xlsfile.Sheet["data"]
	data, err := SheetToSlice(sh)
	assert.Nilf(err, "sheet to slice:%v", err)
	r := NewBaseReader("test", data[3:], "s", 0, 1, 0, 3, 4)
	result, err := r.ReadAll()
	assert.Nilf(err, "error reading content:%v", err)
	assert.Equalf(datawant, result, "%v 出错", "测试xlsx的base格式")
}

func TestBaseReaderForS(t *testing.T) {
	assert := assert.New(t)

	for _, testcase := range tests {
		r := NewBaseReader("test", testcase.data, "s", testcase.keyCount, 1, 3, 2, 4)
		result, err := r.ReadAll()
		assert.Nilf(err, "error reading content:%v", err)
		assert.Equalf(testcase.wants, result, "%v 出错", testcase.name)
	}
}

func TestBaseReaderForC(t *testing.T) {
	assert := assert.New(t)

	for _, testcase := range tests {
		r := NewBaseReader("test", testcase.data, "c", testcase.keyCount, 1, 3, 2, 4)
		result, err := r.ReadAll()
		assert.Nilf(err, "error reading content:%v", err)
		assert.Equalf(testcase.wantc, string(result), "%v 出错", testcase.name)
	}
}
