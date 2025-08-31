package builtinplugins

import (
	"github.com/bookandmusic/tool/internal/common"
	"github.com/bookandmusic/tool/internal/plugins"
	"github.com/bookandmusic/tool/internal/service"
)

var toolInitMeta = &common.Meta{
	Name:    "init",
	Desc:    "tool init command",
	Type:    common.Command,
	BuiltIn: true,
	Commands: []common.CommandDef{
		{
			Name: "config [cfg-path]",
			Desc: "Generate default config file, default: ~/.config/tool.yml",
		},
	},
	Service: &service.ToolInitService{},
}

func init() {
	plugins.RegisterMeta(toolInitMeta)
}
