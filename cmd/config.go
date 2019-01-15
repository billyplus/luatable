package main

// Config 应用全局配置
type Config struct {
	DataPath   string         `mapstructure:"dataPath"`
	Prof       string         `mapstructure:"prof"`
	SkipSubDir bool           `mapstructure:"skipSubDir"`
	Use360     bool           `mapstructure:"use360"`
	Out        []ExportConfig // 每个导出配置会生成一个配置文件
}

// ExportConfig 导出配置
type ExportConfig struct {
	Filter string
	Format string
	GenLua bool `mapstructure:"genlua"`
	Path   string
}
