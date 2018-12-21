package xlsx

import (
	"strings"
)

type Reader interface {
	SetFilterFunc(filterFunc FilterFunc)
	ReadAll() (string, error)
}

//FilterFunc 过滤器
type FilterFunc func(filter string, dest string) bool

func DefaultFilterFunc(filter string, dest string) bool {
	if strings.Contains(filter, dest) {
		return true
	}
	return false
}

func New(typ string) Reader {
	switch typ {
	case "base":
		return nil
	}
	return nil
}
