package encode

import (
	"encoding/json"
)

func EncodeJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "    ")
}
