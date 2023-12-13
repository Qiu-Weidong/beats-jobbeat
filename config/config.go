// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"time"
)

type Config struct {
	Period time.Duration `config:"period"`

	RegistrarPath string `config:"registrar_path"`

	Path []string `config:"path"`
}

var DefaultConfig = Config{

	Period: 1 * time.Second,

	RegistrarPath: "./data/registrar",

	Path: []string{}, // 将默认路径设置为空，此时采集 /tmp 目录
}
