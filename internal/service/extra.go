package service

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bookandmusic/tool/internal/common"
	"github.com/bookandmusic/tool/internal/utils"
)

type ExtraService struct {
	PluginName string
	ExecDir    string
	Exec       string
	ExecType   string
}

// 合并 flags/kwargs
func (e *ExtraService) getMergedArgs(cmd *cobra.Command, cmdParams *common.CmdParams, kwargs map[string]any) (map[string]any, error) {
	starParams := map[string]any{}
	if cmdParams != nil {
		starParams = cmdParams.Flags
	}
	return utils.MergeFlagsAndArgs(starParams, kwargs, cmd)
}

// 构建执行路径
func (e *ExtraService) buildExecPath() string {
	execPath := filepath.Clean(e.Exec)
	if !strings.HasPrefix(execPath, "./") {
		execPath = strings.TrimPrefix(execPath, "/")
		execPath = fmt.Sprintf("./%s", execPath)
	}
	return execPath
}

// 构建最终命令参数
func (e *ExtraService) buildFinalArgs(execPath string, cmdParams *common.CmdParams, mergedArgs map[string]any, args []string) []string {
	executor := common.GlobalCfg.Cfg.Executor
	var finalArgs []string

	// 选择执行器
	switch e.ExecType {
	case "shell":
		finalArgs = append(finalArgs, executor.Shell, execPath)
	case "python":
		finalArgs = append(finalArgs, executor.Python, execPath)
	default:
		finalArgs = append(finalArgs, execPath)
	}

	// 子命令名
	if cmdParams != nil && cmdParams.Name != "" {
		finalArgs = append(finalArgs, cmdParams.Name)
	}

	// 参数转换
	for key, value := range mergedArgs {
		switch v := value.(type) {
		case bool:
			if v {
				finalArgs = append(finalArgs, fmt.Sprintf("--%s", key))
			}
		default:
			finalArgs = append(finalArgs, fmt.Sprintf("--%s=%v", key, v))
		}
	}

	// 追加裸参数
	finalArgs = append(finalArgs, args...)
	return finalArgs
}

func (e *ExtraService) Handler(cmd *cobra.Command, cmdParams *common.CmdParams, args []string, kwargs map[string]any) error {
	console := common.GlobalCfg.Logger

	// 1. 获取参数
	mergedArgs, err := e.getMergedArgs(cmd, cmdParams, kwargs)
	if err != nil {
		return err
	}

	// 2. 生成执行路径
	execPath := e.buildExecPath()

	// 3. 构建最终命令参数
	finalArgs := e.buildFinalArgs(execPath, cmdParams, mergedArgs, args)

	// 4. 执行命令
	err = utils.RunCommand(console, false, nil, e.ExecDir, finalArgs...)
	if err != nil {
		console.Error("[PLUGIN] Plugin '%s' execution failed: %v", e.PluginName, err)
	}
	return err
}
