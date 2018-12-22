package encode

type EncodeFunc func(v interface{}) ([]byte, error)
