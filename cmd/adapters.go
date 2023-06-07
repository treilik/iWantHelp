package cmd

import (
	"github.com/treilik/walder"
	"github.com/treilik/walder/lib"
)

var adapters = []walder.Dimensioner{
	lib.DotDim{},
	lib.FSDim{},
	lib.GoastDim{},
	lib.GitDim{},
	lib.CfgDim{},
	lib.String{},
	lib.Sway{},
}
