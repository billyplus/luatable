package xlsx

import (
	"github.com/pkg/errors"
	"strings"
)

type Reader interface {
	SetFilterFunc(filterFunc FilterFunc)
	ReadAll() (string, error)
}

//FilterFunc 过滤器
type FilterFunc func(filter string, dest string) bool

var ErrNoContent = errors.New("内容为空")

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
