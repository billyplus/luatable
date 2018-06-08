package luatable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLexer(t *testing.T) {
	//t:=test.T
	assert := assert.New(t)

	for _, tc := range lexerTests {
		l := newLexer(tc.name, tc.data)
		i := 0
		// tokCount := len(tc.want)
		// got := make([]Type, 0, tokCount)
		for tok := l.NextToken(); tok.Type() != tokEOF; tok = l.NextToken() {
			if tok.Type() == tokError {
				t.Logf("tok is %v", tok.Type().ToString())
			}

			//t.Logf("%v---%v---%v---%v\n", i, tok.typ.String(), expected[i].String(), tok.val)
			// if tok.typ != expected[i] {
			// 	t.Errorf("got %v expected %v at %v\n", tok.typ.String(), expected[i].String(), tok.val)
			// }
			// fmt.Printf("%q\n", tok.typ)
			typ := tok.Type().ToString()
			// assert.Truef(i <= tokCount, "tok数量超出")
			assert.Equalf(tc.want[i].ToString(), typ, "%v: tok{%v}类型不对", tc.name, tok.Value())
			// got = append(got, typ)
			i++
		}
		// for _, v := range got {
		// 	// t.Logf("%s", v.ToString())

		// }
	}

}

func BenchmarkLexer(b *testing.B) {
	var l Lexer
	for i := 0; i < b.N; i++ {
		l = newLexer("test", tests[0].data)
		for tok := l.NextToken(); tok.Type() != tokEOF; tok = l.NextToken() {
		}
	}
}

var (
	lexerTests = []struct {
		name string
		data string
		want []tokType
	}{
		{ // 双key
			name: "双key测试",
			data: `test={1001={name1={/*1001*/id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},name2={/*1002*/id=1001,name="name2",type=2,icon="icon/head1002",attack=11,life=101}},1003={name3={/*1003*/id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102}},1004={name4={/*1004*/id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}}
`,
			want: []tokType{tokIdent, tokAssign, tokLBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokIdent, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokRBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokRBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokRBrace, tokRBrace, tokEOF},
		},
		{ // 双key
			name: "双key测试中间插注释",
			data: `test={1001={1={/*1001*/id=1001,/*name1*/type=1,icon="icon/head1001",attack=10,life=100}}}
`,
			want: []tokType{tokIdent, tokAssign, tokLBrace, tokInt, tokAssign, tokLBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokRBrace, tokRBrace, tokEOF},
		},
		{ // 单key
			name: "单key测试",
			data: `test={1001={/*1001*/id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},1002={/*1002*/id=1002,name="name2",type=2,icon="icon/head1002",attack=11,life=101},1003={/*1003*/id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102},1004={/*1004*/id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}
`,
			want: []tokType{tokIdent, tokAssign, tokLBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokInt, tokAssign, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokRBrace, tokEOF},
		},
		{ // 数组
			name: "数组测试",
			data: `test={{/*1001*/id=1001,name="name1",type=1,icon="icon/head1001",attack=10,life=100},{/*1002*/id=1001,name="name2",type=2,icon="icon/head1002",attack=11,life=101},{/*1003*/id=1003,name="name3",type=3,icon="icon/head1003",attack=12,life=102},{/*1004*/id=1004,name="name4",type=4,icon="icon/head1004",attack=13,life=103}}
`,
			want: []tokType{tokIdent, tokAssign, tokLBrace, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokLBrace, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokString, tokIdent, tokAssign, tokInt, tokIdent, tokAssign, tokInt, tokRBrace, tokRBrace, tokEOF},
		},
	}
)
