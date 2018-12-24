package encode

import (
	"bytes"
	"encoding/json"
)

func EncodeJSON(v interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	err := enc.Encode(v)
	return buf.Bytes(), err
}
