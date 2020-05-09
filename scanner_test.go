package luatable

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScan(t *testing.T) {
	for _, test := range []struct {
		tok           tokType
		src, val, err string
	}{
		{tokInt, "-1000", "-1000", ""},
		{tokInt, "1000", "1000", ""},
		{tokInt, "9933300", "9933300", ""},
		{tokInt, "0003333", "0003333", ""},
		{tokFloat, "0.3333", "0.3333", ""},
		{tokFloat, "7545460.3333", "7545460.3333", ""},
		{tokFloat, "00.3333", "00.3333", ""},
		{tokString, "\"aa888809\"", "\"aa888809\"", ""},
		{tokString, "\"'aa888809\"", "\"'aa888809\"", ""},
		{tokString, "'\"aa888809\"'", "\"\\\"aa888809\\\"\"", ""},
		{tokString, "''", "\"\"", ""},
		{tokString, "\"\"", "\"\"", ""},
		{tokString, "'aa888809\"'", "\"aa888809\\\"\"", ""},
		{tokIdent, "__test", "__test", ""},
		{tokIdent, "aatext", "aatext", ""},
		{tokIdent, "aat_ex_t", "aat_ex_t", ""},
		{tokIdent, "aat.ex_t", "aat.ex_t", ""},
		{tokComment, "--[[aat.ex_t]]", "--[[aat.ex_t]]", ""},
		{tokComment, "--aaccc  \n     ", "--aaccc  ", ""},
		{tokComment, " --满血\n     ", "--满血", ""},
		{tokError, ".ex_t", "", ""},
	} {
		s := scanner{}
		s.Init([]byte(test.src))
		_, tok, val := s.Scan()
		assert.Equalf(t, test.tok, tok, "%s: token should be equal", test.src)
		assert.Equalf(t, test.val, val, "%s: val should be equal", test.src)
	}
}

func TestScanner(t *testing.T) {
	//t:=test.T
	assert := assert.New(t)
	s := scanner{}

	for _, tc := range scannerTests {
		s.Init([]byte(tc.data))
		i := 0
		for pos, tok, val := s.Scan(); tok != tokEOF; pos, tok, val = s.Scan() {
			fmt.Println(pos, tok, val)
			assert.Equalf(tc.want[i], tok, "%v: i=%d tok={%v}类型不对, expect: %s, got %s", tc.name, i, val, tc.want[i].ToString(), tok.ToString())
			i++
		}
	}
}

var (
	scannerTests = []struct {
		name string
		data string
		want []tokType
	}{
		{ // 双key
			name: "双key测试",
			data: `test={1001={name1={--[[1001]]id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},name2={id=1001,name="name2",type=2,icon="icon/head1002",attack=11,life=101}},1003={name3={id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102}},1004={name4={id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}}
`,
			want: []tokType{tokIdent, tokAssign, tokLBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokLBrace, tokComment, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokIdent, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokRBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokRBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokRBrace, tokRBrace, tokEOF},
		},
		{ // 双key
			name: "双key测试中间插注释",
			data: `test={1001={1={id=1001,type=1,icon="icon/head1001",attack=10,life=100}}}
`,
			want: []tokType{tokIdent, tokAssign, tokLBrace, tokInt, tokAssign, tokLBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokRBrace, tokRBrace, tokEOF},
		},
		{ // 单key
			name: "单key测试",
			data: `test={1001={id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},1002={id=1002,name="name2",type=2,icon="icon/head1002",attack=11,life=101},1003={id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102},1004={id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}
`,
			want: []tokType{tokIdent, tokAssign, tokLBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokRBrace, tokEOF},
		},
		{ // 数组
			name: "数组测试",
			data: `test={{id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},{id=1001,name="name2",type=2,icon="icon/head1002",attack=11,life=101},{id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102},{id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}
`,
			want: []tokType{tokIdent, tokAssign, tokLBrace, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign,
				tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign,
				tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokLBrace, tokIdent, tokAssign, tokInt,
				tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString,
				tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokLBrace, tokIdent,
				tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent,
				tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace,
				tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign,
				tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign,
				tokInt, tokRBrace, tokRBrace, tokEOF},
		},
		{
			name: "注释测试",
			data: `{
	{
		id=1001,
		name="name1",
		life=100,
		param={
	{
		delay = 10000,
		count = 1,
		actions={{type=6}}, --测试
	}
}
	},
	{
		id=1002,
		name="name2",
		life=101,
		param={}
	},
	{
		id=1004,
		name="name4",
		life=103,
		param={
	{
		delay = 10000,
		count = 1,
		actions={{type=6}}, --满血
	}
}
	}
}
`,
			want: []tokType{tokLBrace, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent,
				tokAssign, tokInt, tokIdent, tokAssign, tokLBrace, tokLBrace, tokIdent, tokAssign, tokInt,
				tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokLBrace, tokLBrace, tokIdent, tokAssign,
				tokInt, tokRBrace, tokRBrace, tokComment, tokRBrace, tokRBrace, tokRBrace, tokLBrace, tokIdent,
				tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent,
				tokAssign, tokLBrace, tokRBrace, tokRBrace, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent,
				tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokLBrace, tokLBrace,
				tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokLBrace,
				tokLBrace, tokIdent, tokAssign, tokInt, tokRBrace, tokRBrace, tokComment, tokRBrace, tokRBrace,
				tokRBrace, tokRBrace, tokEOF},
		},
	}
)
