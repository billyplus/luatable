package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkGenConfig(b *testing.B) {
	conf := Config{
		DataPath: "../xlsx",
		Out: []ExportConfig{
			{
				Path:     "./server",
				Filter:   "s",
				Format:   "lua",
				OutCellX: 1,
				OutCellY: 1,
			},
		},
	}
	for i := 0; i < b.N; i++ {
		gen := newGenerator()
		gen.config = &conf
		gen.iterateXlsx("../xlsx/test.xlsx")

	}
}

func TestGenConfig(t *testing.T) {
	conf := Config{
		DataPath: "../xlsx",
		Out: []ExportConfig{
			{
				Path:     "./server",
				Filter:   "s",
				Format:   "lua",
				OutCellX: 1,
				OutCellY: 1,
			},
			{
				Path:     "./client",
				Filter:   "c",
				Format:   "json",
				OutCellX: 1,
				OutCellY: 1,
			},
		},
	}
	gen := newGenerator()
	gen.config = &conf
	assert.NoError(t, gen.iterateXlsx("../xlsx/test.xlsx"), "err should be nil")

}
