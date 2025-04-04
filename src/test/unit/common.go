package unit

import (
	config "duolingo/lib/config_reader"
	"path/filepath"
)

var (
	conf = config.NewJsonReader(filepath.Join("..", "..", "infra", "config"))
)
