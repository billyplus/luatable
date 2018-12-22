package main

import (
	"testing"
)

func BenchmarkGenConfig(b *testing.B) {
	conf := Config{
		DataPath: "../xlsx",
		Out: []ExportConfig{
			ExportConfig{
				Path:   "./server",
				Filter: "s",
				Format: "xml",
			},
			ExportConfig{
				Path:   "./client",
				Filter: "c",
				Format: "json",
			},
		},
	}
	for i := 0; i < b.N; i++ {
		gen := newGenerator(20)
		gen.GenConfig(conf)

	}
}
