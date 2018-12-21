package xlsx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	tinytests = []struct {
		name  string
		data  [][]string
		wants string
		wantc string
	}{
		{
			name: "测试tiny表",
			data: [][]string{{"子id", "sc", "childId", "int", "113"},
				{"开服活动", "sc", "type", "string", "这是一行"},
				{"持续时间", "s", "keep_t", "raw", "{0,604800}"},
				{"等级要求", "s", "lmLevel", "int", "25"},
				{"等级要求", "c", "itemLevel", "int", "250"},
				{"道具名称", "c", "itemName", "string", "道具名称"}},
			wants: `testConf={childId=113,type="这是一行",keep_t={0,604800},lmLevel=25,}`,
			wantc: `testConf={childId=113,type="这是一行",itemLevel=250,itemName="道具名称",}`,
		},
	}
)

func TinyReaderTest(t *testing.T, name string, data [][]string, filter, want string) {
	r := NewTinyReader("testConf", data, filter, 1, 2, 3, 4)
	result, _ := r.ReadAll()
	assert.Equal(t, want, result, name)
}

func TestTinyReaderForS(t *testing.T) {
	for _, c := range tinytests {
		TinyReaderTest(t, c.name, c.data, "s", c.wants)
	}
}

func TestTinyReaderForC(t *testing.T) {
	for _, c := range tinytests {
		TinyReaderTest(t, c.name, c.data, "c", c.wantc)
	}
}
