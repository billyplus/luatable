package token

// Position 表示token所在的位置
type Position interface {
	String() string // 返回position的字符串表达式
	// IsValid() bool  // 判断postion是否有效
}
