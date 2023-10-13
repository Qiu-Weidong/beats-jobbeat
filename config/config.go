// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"time"
)

type Config struct {
	Period time.Duration `config:"period"`

	Path string `config: "path"`

	RegistrarPath string `config: "registrar_path"`
}

var DefaultConfig = Config{

	Period: 1 * time.Second,

	Path: ".",

	RegistrarPath: "./data/registrar",
}
