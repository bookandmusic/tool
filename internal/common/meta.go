package common

import (
	"fmt"
)

// 定义 PluginType 类型
type PluginType int

const (
	Soft    PluginType = iota // 0
	Command                   // 1
)

// 为 PluginType 类型实现 String 方法
func (p PluginType) String() string {
	switch p {
	case Soft:
		return "soft"
	case Command:
		return "command"
	default:
		return "unknown"
	}
}

// 为 PluginType 类型实现 UnmarshalYAML 方法
func (p *PluginType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	// 根据字符串值设置 PluginType
	switch s {
	case "soft":
		*p = Soft
	case "command":
		*p = Command
	default:
		return fmt.Errorf("invalid plugin type: %s", s)
	}
	return nil
}

type CommandFlag struct {
	Name    string `yaml:"name"`
	Desc    string `yaml:"desc"`
	Default any    `yaml:"default"`
}

type CommandDef struct {
	Name  string         `yaml:"name"`
	Flags []*CommandFlag `yaml:"flags"`
	Desc  string         `yaml:"desc"`
}

type Meta struct {
	Name     string         `yaml:"name"`
	Desc     string         `yaml:"desc"`
	Type     PluginType     `yaml:"type"`
	Exec     string         `yaml:"exec"`
	ExecType string         `yaml:"exec_type"`
	Flags    []*CommandFlag `yaml:"flags"`    // 通用 flags
	Commands []CommandDef   `yaml:"commands"` // 子命令
	Dir      string         `yaml:"-"`        // 插件目录
	BuiltIn  bool           `yaml:"-"`        // 内置插件
	Service  Service        `yaml:"-"`        // 插件绑定的服务实例
}

func (m *Meta) GetCommand(name string) *CommandDef {
	var cmd *CommandDef
	for _, c := range m.Commands {
		if c.Name == name {
			cmd = &c
			break
		}
	}
	return cmd
}
