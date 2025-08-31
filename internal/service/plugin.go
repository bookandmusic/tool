package service

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/bookandmusic/tool/internal/common"
	"github.com/bookandmusic/tool/internal/plugins"
	"github.com/bookandmusic/tool/internal/utils"
)

type PluginService struct{}

func (p *PluginService) enabled(args []string) error {
	console := common.GlobalCfg.Logger
	cfg := common.GlobalCfg.Cfg
	cfgPath := common.GlobalCfg.CfgPath

	if len(args) == 0 {
		console.Error("[PLUGIN] No plugin name specified")
		return nil
	}

	if cfg.Plugins == nil {
		cfg.Plugins = make(map[string]common.PluginConfig)
	}

	for _, name := range args {
		meta, err := p.validateSoftPlugin(name)
		if err != nil {
			continue
		}

		if p.isPluginEnabled(cfg, name) {
			console.Debug(fmt.Sprintf("[PLUGIN] Plugin '%s' already enabled, skipping\n", name))
			continue
		}

		p.enablePlugin(cfg, name, meta)
		console.Info("[PLUGIN] Plugin '%s' enabled", name)
	}

	if err := common.SaveConfig(cfgPath, cfg); err != nil {
		return err
	}
	common.GlobalCfg.Cfg = cfg

	console.Success("[PLUGIN] Plugins enabled successfully. You can modify parameters manually in the config file.")
	return nil
}

func (p *PluginService) disable(args []string) error {
	console := common.GlobalCfg.Logger
	cfg := common.GlobalCfg.Cfg
	cfgPath := common.GlobalCfg.CfgPath

	if len(args) == 0 {
		console.Error("[PLUGIN] No plugin name specified to disable.")
		return nil
	}

	enabledSet := p.buildEnabledSet(cfg)
	removed := []string{}

	for _, name := range args {
		if _, ok := enabledSet[name]; !ok {
			console.Warning("[PLUGIN] Plugin '%s' not enabled, skipping", name)
			continue
		}
		p.removePlugin(cfg, name)
		removed = append(removed, name)
	}

	if len(removed) > 0 {
		if err := common.SaveConfig(cfgPath, cfg); err != nil {
			return err
		}
		for _, name := range removed {
			console.Success("[PLUGIN] Plugin '%s' disabled", name)
		}
	} else {
		console.Info("[PLUGIN] No plugins were disabled")
	}

	return nil
}

// ------------------- PluginService 辅助方法 -------------------

func (p *PluginService) validateSoftPlugin(name string) (*common.Meta, error) {
	console := common.GlobalCfg.Logger
	meta := plugins.GetMetaByName(name)
	if meta == nil {
		console.Error("[PLUGIN] Plugin '%s' not found\n", name)
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}
	if meta.Type != common.Soft {
		console.Error("[PLUGIN] Plugin '%s' is not a soft plugin and cannot be enabled\n", name)
		return nil, fmt.Errorf("plugin '%s' is not a soft plugin", name)
	}
	if meta.GetCommand("install") == nil || meta.GetCommand("uninstall") == nil {
		console.Error("[PLUGIN] Soft plugin '%s' must define 'install' and 'uninstall' commands in meta.yml\n", name)
		return nil, fmt.Errorf("soft plugin '%s' missing install/uninstall commands", name)
	}
	console.Debug(fmt.Sprintf("[PLUGIN] Plugin '%s' validated successfully\n", name))
	return meta, nil
}

func (p *PluginService) isPluginEnabled(cfg *common.Config, name string) bool {
	for _, n := range cfg.EnabledPlugins {
		if n == name {
			return true
		}
	}
	return false
}

func (p *PluginService) enablePlugin(cfg *common.Config, name string, meta *common.Meta) {
	installDef := meta.GetCommand("install")
	uninstallDef := meta.GetCommand("uninstall")
	globalFlags := utils.CmdFlagsToMap(meta.Flags)
	installFags := utils.CmdFlagsToMap(installDef.Flags)
	uninstallFlags := utils.CmdFlagsToMap(uninstallDef.Flags)

	cfg.Plugins[name] = common.PluginConfig{
		Install:   utils.MergeFlags(installFags, globalFlags, "fill"),
		Uninstall: utils.MergeFlags(uninstallFlags, globalFlags, "fill"),
	}
	cfg.EnabledPlugins = append(cfg.EnabledPlugins, name)
}

func (p *PluginService) buildEnabledSet(cfg *common.Config) map[string]struct{} {
	enabledSet := make(map[string]struct{}, len(cfg.EnabledPlugins))
	for _, n := range cfg.EnabledPlugins {
		enabledSet[n] = struct{}{}
	}
	return enabledSet
}

func (p *PluginService) removePlugin(cfg *common.Config, name string) {
	delete(cfg.Plugins, name)

	newList := []string{}
	for _, n := range cfg.EnabledPlugins {
		if n != name {
			newList = append(newList, n)
		}
	}
	cfg.EnabledPlugins = newList
}

func (p *PluginService) list() error {
	cfg := common.GlobalCfg.Cfg
	console := common.GlobalCfg.Logger

	// 创建 table
	t := table.NewWriter()
	t.SetOutputMirror(console.Writer())
	t.AppendHeader(table.Row{"Plugin Name", "Type", "Status"})

	// 表头加粗
	t.Style().Format.Header = text.FormatDefault
	t.Style().Options.DrawBorder = false      // 去掉边框
	t.Style().Options.SeparateColumns = false // 去掉列分隔
	t.Style().Options.SeparateRows = false    // 去掉行分隔

	// 准备数据行
	for _, meta := range plugins.ListManyByType(common.Soft) {
		status := "Disabled"
		if contains(cfg.EnabledPlugins, meta.Name) {
			status = "Enabled"
		}

		// 状态列彩色显示
		var statusCell string
		if status == "Enabled" {
			statusCell = text.FgGreen.Sprint(status)
		} else {
			statusCell = text.FgHiBlack.Sprint(status)
		}

		t.AppendRow(table.Row{meta.Name, meta.Type.String(), statusCell})
		console.Debug(fmt.Sprintf("[PLUGIN] Plugin '%s' status: %s\n", meta.Name, status))
	}

	t.Render() // 渲染表格

	return nil
}

// 判断 slice 是否包含某个元素
func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func (e *PluginService) initOrDestroy(cmd *cobra.Command, action string, args []string) error {
	console := common.GlobalCfg.Logger
	cfg := common.GlobalCfg.Cfg
	enabledPlugins := cfg.EnabledPlugins
	if len(enabledPlugins) == 0 {
		console.Warning("[PLUGIN] No active plugins, use: tool plugin enable [plugin-name]")
		return nil
	}
	for _, name := range enabledPlugins {
		meta := plugins.GetMetaByName(name)
		if meta == nil || meta.Type != common.Soft {
			continue
		}

		// 找到对应的 install/uninstall 子命令

		subCmd := meta.GetCommand(action)
		if subCmd == nil {
			console.Error("[PLUGIN] Soft plugin '%s' missing %s command", name, action)
			continue
		}
		var kwargs map[string]any
		if action == "install" {
			kwargs = cfg.Plugins[name].Install
		} else {
			kwargs = cfg.Plugins[name].Uninstall
		}
		flags := utils.CmdFlagsToMap(subCmd.Flags)

		// 执行插件的处理函数
		console.Debug(fmt.Sprintf("[PLUGIN] Executing plugin '%s' %s command\n", name, action))
		if err := meta.Service.Handler(cmd, &common.CmdParams{
			Name:  subCmd.Name,
			Flags: flags,
		}, args, kwargs); err != nil {
			console.Error("[PLUGIN] Plugin '%s' %s failed: %v", name, action, err)
			return err
		}
		console.Debug(fmt.Sprintf("[PLUGIN] Plugin '%s' %s executed successfully\n", name, action))
	}
	return nil
}

func (p *PluginService) Handler(cmd *cobra.Command, cmdParams *common.CmdParams, args []string, kwargs map[string]any) error {
	switch cmdParams.Name {
	case "":
		return cmd.Help()
	case "enable [name]":
		return p.enabled(args)
	case "disable [name]":
		return p.disable(args)
	case "ls":
		return p.list()
	case "install":
		return p.initOrDestroy(cmd, "install", args)
	case "destroy":
		return p.initOrDestroy(cmd, "uninstall", args)
	}

	return nil
}
