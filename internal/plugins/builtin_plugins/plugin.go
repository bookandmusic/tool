package builtinplugins

import (
	"github.com/bookandmusic/tool/internal/common"
	"github.com/bookandmusic/tool/internal/plugins"
	"github.com/bookandmusic/tool/internal/service"
)

var pluginMeta = &common.Meta{
	Name:    "soft",
	Desc:    "Manage soft (enable/disable/install/uninstall)",
	Type:    common.Command,
	BuiltIn: true,
	Commands: []common.CommandDef{
		{
			Name: "enable [name]",
			Desc: "Enable a soft plugin (writes install/uninstall flags into config)",
		},
		{
			Name: "disable [name]",
			Desc: "Disable a soft plugin (removes it from enabled list and config)",
		},
		{
			Name: "ls",
			Desc: "List soft plugins",
		},
		{
			Name: "install",
			Desc: "install all enabled soft plugins",
		},
		{
			Name: "destroy",
			Desc: "uninstall all enabled soft plugins",
		},
	},
	Service: &service.PluginService{},
}

func init() {
	plugins.RegisterMeta(pluginMeta)
}
