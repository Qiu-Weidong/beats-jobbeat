package main

import (
	"os"

	"github.com/Qiu-Weidong/jobbeat/cmd"

	_ "github.com/Qiu-Weidong/jobbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
