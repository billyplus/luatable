package token

// filter 自定义过滤器，用于过滤不需要的token
type filter interface {
	IsValid(cond interface{}) bool
}
