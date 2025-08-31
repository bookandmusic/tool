package service

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bookandmusic/tool/internal/common"
)

type ToolInitService struct {
}

func (h *ToolInitService) cfg(args []string) error {
	console := common.GlobalCfg.Logger
	var configPath string
	if len(args) == 0 {
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".config", "tool.yml")
	} else {
		configPath = args[0]
	}

	defaultCfg := common.GenerateDefault()
	if err := common.SaveConfig(configPath, defaultCfg); err != nil {
		return err
	}
	console.Success("Generated config at %s", configPath)
	return nil
}

func (h *ToolInitService) Handler(cmd *cobra.Command, cmdParams *common.CmdParams, args []string, kwargs map[string]any) error {
	switch cmdParams.Name {
	case "":
		return cmd.Help()
	case "config [cfg-path]":
		return h.cfg(args)
	}
	return nil
}
