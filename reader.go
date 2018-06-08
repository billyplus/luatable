package luatable

// Reader 接口读取全部内容，返回string
type Reader interface {
	ReadAll() (string, error)
}
