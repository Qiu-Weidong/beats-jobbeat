package cmd

import (
	"github.com/Qiu-Weidong/jobbeat/beater"

	cmd "github.com/elastic/beats/v7/libbeat/cmd"
	"github.com/elastic/beats/v7/libbeat/cmd/instance"
)

// Name of this beat
var Name = "jobbeat"

// RootCmd to handle beats cli
var RootCmd = cmd.GenRootCmdWithSettings(beater.New, instance.Settings{Name: Name})
