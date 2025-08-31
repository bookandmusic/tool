package plugins

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bookandmusic/tool/internal/common"
	"github.com/bookandmusic/tool/internal/utils"
)

func LoadAll(root *cobra.Command) error {
	console := common.GlobalCfg.Logger
	var softCmd *cobra.Command
	for _, meta := range ListManyByType(common.Command) {
		cmd := BuildPluginCmd(meta)
		if meta.Name == "soft" {
			softCmd = cmd
		}
		root.AddCommand(cmd)
		console.Debug(fmt.Sprintf("[PLUGIN] %s loaded (type: %s)\n", meta.Name, meta.Type))
	}
	if softCmd == nil {
		softCmd = root
	}
	for _, meta := range ListManyByType(common.Soft) {
		cmd := BuildPluginCmd(meta)
		softCmd.AddCommand(cmd)
		console.Debug(fmt.Sprintf("[PLUGIN] %s loaded (type: %s)\n", meta.Name, meta.Type))
	}
	return nil
}

func addFlagsToCmd(cmd *cobra.Command, flags []*common.CommandFlag) error {
	for _, flag := range flags {
		flagValue := flag.Default
		flagName := flag.Name
		flagUsage := flag.Desc
		if flagUsage == "" {
			flagUsage = fmt.Sprintf("Flag for %s", flagName)
		}
		// 假设 flags 是 "key" -> "value" 的结构，可以根据需要添加不同类型的标志
		switch v := flagValue.(type) {
		case string:
			// 添加字符串类型的标志
			cmd.Flags().String(flagName, v, flagUsage)
		case bool:
			// 添加布尔类型的标志
			cmd.Flags().Bool(flagName, v, flagUsage)
		case int:
			// 添加整型类型的标志
			cmd.Flags().Int(flagName, v, flagUsage)
		default:
			// 如果是其他类型，可以根据需要继续扩展
			cmd.Flags().String(flagName, fmt.Sprintf("%v", flagValue), flagUsage)
		}
		console := common.GlobalCfg.Logger
		console.Debug(fmt.Sprintf("[PLUGIN] Added flag '%s' (default: %v) to command '%s'\n", flagName, flagValue, cmd.Use))
	}
	return nil
}

func BuildPluginCmd(meta *common.Meta) *cobra.Command {
	console := common.GlobalCfg.Logger
	short := meta.Desc
	if short == "" {
		short = fmt.Sprintf("%s %s plugin", meta.Name, meta.Type)
	}

	// 创建顶层命令
	pluginCmd := &cobra.Command{
		Use:   meta.Name,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return meta.Service.Handler(cmd, &common.CmdParams{
				Flags: utils.CmdFlagsToMap(meta.Flags),
			}, args, nil)
		},
	}
	if meta.Flags != nil {
		if err := addFlagsToCmd(pluginCmd, meta.Flags); err != nil {
			console.Warning("[PLUGIN] Error adding flags to '%s': %v", meta.Name, err)
		}
	}

	// 为每个子命令创建子命令并添加到顶层命令
	for _, c := range meta.Commands {
		sub := c
		subShort := sub.Desc
		if subShort == "" {
			subShort = fmt.Sprintf("%s %s", meta.Name, sub.Name)
		}
		subCmd := &cobra.Command{
			Use:   sub.Name,
			Short: subShort,
			RunE: func(cmd *cobra.Command, args []string) error {

				return meta.Service.Handler(cmd, &common.CmdParams{
					Name:  c.Name,
					Flags: utils.CmdFlagsToMap(c.Flags),
				}, args, nil)
			},
		}
		if sub.Flags != nil {
			if err := addFlagsToCmd(subCmd, c.Flags); err != nil {
				console.Warning("[PLUGIN] Error adding flags to subcommand '%s': %v", c.Name, err)
			}
		}
		pluginCmd.AddCommand(subCmd)
		console.Debug(fmt.Sprintf("[PLUGIN] Subcommand '%s' added to '%s'\n", sub.Name, pluginCmd.Use))
	}
	return pluginCmd
}
