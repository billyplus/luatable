package encode

import (
	// "fmt"
	"github.com/beevik/etree"
	"github.com/pkg/errors"
	"reflect"
	"sort"
	"strconv"
)

// EncodeXML 导出xml的配置文件
func EncodeXML(v interface{}) ([]byte, error) {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="utf-8"`)
	switch reflect.TypeOf(v).Kind() {
	case reflect.Map:
		value := reflect.ValueOf(v)
		keys := value.MapKeys()
		if len(keys) == 0 {
			return nil, nil
		}
		key := keys[0]
		iElem := value.MapIndex(key).Interface()
		switch elem := iElem.(type) {
		case []interface{}:
			if len(elem) == 0 {
				return nil, nil
			}
			for _, arr := range elem {
				child, err := encodeXMLFromArray(arr)
				if err != nil {
					return nil, err
				}
				doc.AddChild(child)
			}
		case map[string]interface{}:
			child, err := encodeMap(elem)
			if err != nil {
				return nil, err
			}
			doc.AddChild(child)
		default:
			return nil, errors.Errorf("不支持的类型:%s:%v\n", reflect.TypeOf(elem).Kind(), elem)
		}
	}
	doc.Indent(4)
	return doc.WriteToBytes()
}

func encodeXMLFromArray(value interface{}) (*etree.Element, error) {

	switch realV := value.(type) {
	case map[string]interface{}:
		return encodeMap(realV)
	default:
		return nil, errors.Errorf("不支持的类型:%s:%v\n", reflect.TypeOf(realV).Kind(), realV)
	}
}

func encodeMap(v map[string]interface{}) (*etree.Element, error) {
	elem := etree.NewElement("i")
	// keys := reflect.ValueOf(v).MapKeys()
	var err error
	sortedMapString(v, func(k string, value interface{}) {
		str, err2 := interfaceToString(value)
		if err2 != nil {
			err = err2
			return
		}
		elem.CreateAttr(k, str)
	})
	if err != nil {
		return nil, err
	}
	// for i := 0; i < len(keys); i++ {

	// 	key := keys[i].Interface().(string)
	// 	value := v[key]

	// }
	return elem, nil
}

func interfaceToString(v interface{}) (string, error) {
	switch value := v.(type) {
	case string:
		return value, nil
	case int64:
		return strconv.FormatInt(value, 10), nil
	case float64:
		return strconv.FormatFloat(value, 'f', 4, 64), nil
	default:
		return "", errors.Errorf("不支持的类型:%s:%v\n", reflect.TypeOf(value).Kind(), value)
	}
}

func sortedMapString(m map[string]interface{}, f func(string, interface{})) {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f(k, m[k])
	}
}
