package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bookandmusic/tool/internal/common"
	"github.com/bookandmusic/tool/internal/logger"
	"github.com/bookandmusic/tool/internal/plugins"
	_ "github.com/bookandmusic/tool/internal/plugins/builtin_plugins"
	extraplugins "github.com/bookandmusic/tool/internal/plugins/extra_plugins"
)

var (
	configPath string
	debug      bool
	cfg        *common.Config
	console    logger.Logger
)

var rootCmd = &cobra.Command{
	Use:           "tool",
	Short:         "tool with plugin system",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error { return rootCmd.Execute() }

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")

	_ = rootCmd.PersistentFlags().Parse(os.Args)
	console = logger.NewConsoleLogger(os.Stdout, debug)
	// 提前加载配置和插件命令
	if configPath == "" {
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".config", "tool.yml")
	}

	if _, err := os.Stat(configPath); err == nil {
		cfg, err = common.LoadConfig(configPath)
		if err != nil {
			console.Debug(fmt.Sprintf("[CONFIG] Failed to parse config file: %s, error: %v\n", configPath, err))
		} else {
			console.Debug(fmt.Sprintf("[CONFIG] Loaded configuration from %s\n", configPath))
		}
	} else {
		console.Debug(fmt.Sprintf("[CONFIG] Config file does not exist: %s\n", configPath))
	}

	if cfg == nil {
		cfg = common.GenerateDefault()
		console.Warning("[CONFIG] Config file not found, using default configuration. Run `tool init config [cfg-path]` to generate default config.")
	}
	common.GlobalCfg = &common.GlobalConfig{
		Cfg:     cfg,
		CfgPath: configPath,
		Logger:  console,
	}
	_ = extraplugins.LoadAllExtraPluginMeta(cfg)
	if err := plugins.LoadAll(rootCmd); err != nil {
		console.Error("[PLUGIN] Failed to load plugins: %v", err)
	} else {
		console.Debug("[PLUGIN] All plugins loaded successfully\n")
	}
}
