package main

import (
	// "bytes"
	// "github.com/billyplus/luatable/encode"
	"bytes"

	"github.com/billyplus/luatable/xlsx"
)

// EncodeTSClass 根据 interface 生成ts定义文件
// func EncodeTSClass(v interface) ([]byte,error){
// 	switch val:=v.(type) {
// 	case []interface{}:
// 	case map[string]interface{}:
// 	default:
// 		return nil, errors.Errorf("EncodeTSClass不支持的类型:%s:%v\n", reflect.TypeOf(value).Kind(), value)
// 	}
// }

type prop struct {
	Name string
	Type string
	Comm string
}

func GenTSFile(sheet *WorkSheet, filter string) ([]byte, error) {
	nameRow := 0
	filterRow := 1
	typRow := 3
	commRow := 2

	builder := new(bytes.Buffer)
	// builder := bytes.NewBuffer(make([]byte, 2048))
	builder.WriteString("class ")
	builder.WriteString(sheet.Name)
	builder.WriteString(" {\n")
	builder.WriteString(`    public static ClassName = "`)
	builder.WriteString(sheet.Name)
	builder.WriteString("\"; \n")
	for i := 0; i < len(sheet.Data[0]); i++ {
		if xlsx.DefaultFilterFunc(sheet.Data[filterRow][i], filter) {
			name := sheet.Data[nameRow][i]
			typ := sheet.Data[typRow][i]
			comm := sheet.Data[commRow][i]
			builder.WriteString("    /** ")
			builder.WriteString(comm)
			builder.WriteString(" */\n")
			builder.WriteString("    public ")
			builder.WriteString(name)
			builder.WriteString(": ")
			builder.WriteString(getTypeFromString(typ))
			builder.WriteString(";\n")
		}
	}
	builder.WriteString("}\n")

	if builder.Len() > 50+2*len(sheet.Name) {
		return builder.Bytes(), nil
	}
	return nil, xlsx.ErrNoContent
}

func getTypeFromString(v string) string {
	switch v {
	case "string":
		return "string"
	case "int", "float":
		return "number"
	case "bool":
		return "boolean"
	default:
		return "any"
	}
}
