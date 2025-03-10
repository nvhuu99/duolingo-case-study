package test

import (
	"duolingo/lib/config-reader"
	"path/filepath"
)

var (
	conf = config.NewJsonReader(filepath.Join("..", "infra", "config"))
)
