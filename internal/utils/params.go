package utils

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/bookandmusic/tool/internal/common"
)

func CmdFlagsToMap(cmdFlags []*common.CommandFlag) map[string]any {
	flags := map[string]any{}
	for _, flag := range cmdFlags {
		flags[flag.Name] = flag.Default
	}
	return flags
}

func mergeCover(base, global map[string]any) {
	for k, v := range global {
		if _, ok := base[k]; ok {
			base[k] = v
		}
	}
}

func mergeFill(base, global map[string]any) {
	for k, v := range global {
		if _, ok := base[k]; !ok {
			base[k] = v
		}
	}
}

func mergeFull(base, global map[string]any) {
	for k, v := range global {
		base[k] = v
	}
}

// MergeFlags merges global into base with specific strategy.
// modes:
//   - "cover": overwrite base's existing keys only (no new keys)
//   - "fill" : add missing keys only (no overwrite)
//   - "full" : add + overwrite (global wins on conflict)
func MergeFlags(base, global map[string]any, mode string) map[string]any {
	if base == nil {
		base = make(map[string]any)
	}
	if len(global) == 0 {
		return base
	}

	switch mode {
	case "cover":
		mergeCover(base, global)
	case "fill":
		mergeFill(base, global)
	case "full":
		mergeFull(base, global)
	default:
		mergeFill(base, global) // 默认 fill
	}
	return base
}

func MergeFlagsAndArgs(flags, kwargs map[string]interface{}, cmd *cobra.Command) (map[string]any, error) {
	// 创建一个 map 来存储最终的参数
	cmdFlags := cmd.Flags()
	flags = MergeFlags(flags, kwargs, "cover")

	// 遍历所有的标志
	cmdFlags.VisitAll(func(f *pflag.Flag) {
		if _, ok := flags[f.Name]; ok {
			var err error
			switch f.Value.Type() {
			case "string":
				// 如果是 string 类型，直接使用 f.Value.String()
				flags[f.Name] = f.Value.String()
			case "int":
				// 如果是 int 类型，使用 strconv.Atoi 将 string 转换为 int
				flags[f.Name], err = strconv.Atoi(f.Value.String())
			case "int64":
				// 如果是 int64 类型，使用 strconv.ParseInt 将 string 转换为 int64
				flags[f.Name], err = strconv.ParseInt(f.Value.String(), 10, 64)
			case "bool":
				// 如果是 bool 类型，使用 strconv.ParseBool 将 string 转换为 bool
				flags[f.Name], err = strconv.ParseBool(f.Value.String())
			case "float64":
				// 如果是 float64 类型，使用 strconv.ParseFloat 将 string 转换为 float64
				flags[f.Name], err = strconv.ParseFloat(f.Value.String(), 64)
			default:
				// 如果没有匹配的类型，使用字符串表示
				flags[f.Name] = f.Value.String()
			}
			// 如果转换出错，返回错误
			if err != nil {
				flags[f.Name] = f.Value.String()
			}
		}
	})

	return flags, nil
}
