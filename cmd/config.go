package main

type Config struct {
	DataPath string `mapstructure:"dataPath"`
	Out      []ExportConfig
}

type ExportConfig struct {
	Filter string
	Format string
	Path   string
}
