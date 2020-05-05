package encode

import (
	"github.com/billyplus/luatable"
	"github.com/stretchr/testify/assert"

	// "reflect"
	"testing"
)

func TestEncodeXML(t *testing.T) {
	for _, testcase := range xmltests {
		var v interface{}
		err := luatable.Unmarshal([]byte(testcase.data), &v)
		assert.Nilf(t, err, "error reading content:%v", err)
		data, err := EncodeXML(v)
		assert.Nilf(t, err, "error reading content:%v", err)
		assert.Equal(t, testcase.want, string(data))
		// assert.Equalf(t, testcase.wants, string(result), "%v 出错", testcase.name)
	}
}

var xmltests = []struct {
	data string
	want string
}{
	{
		data: `{{id=1001,name="名字1",type=1,attack=10,life=100},{id=1001,name="名字2",type=2,attack=11,life=101},{id=1003,name="名字3",type=3,attack=12,life=102},{id=1004,name="名字4",type=4,attack=13,life=103}}
		`,
		want: `<?xml version="1.0" encoding="utf-8"?>
<root>
    <i attack="10" id="1001" life="100" name="名字1" type="1"/>
    <i attack="11" id="1001" life="101" name="名字2" type="2"/>
    <i attack="12" id="1003" life="102" name="名字3" type="3"/>
    <i attack="13" id="1004" life="103" name="名字4" type="4"/>
</root>
`,
	},
}
