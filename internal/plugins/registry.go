package plugins

import "github.com/bookandmusic/tool/internal/common"

// 注册表，用来存储每个 PluginType 对应的 Meta 切片
var Registry = map[common.PluginType][]*common.Meta{}

// RegisterMeta 注册新的插件元数据到 Registry
func RegisterMeta(meta *common.Meta) {
	pluginType := meta.Type
	Registry[pluginType] = append(Registry[pluginType], meta)
}

// ListAll 返回所有插件，不按类型分类
func ListAll() []*common.Meta {
	var allPlugins []*common.Meta
	for _, plugins := range Registry {
		allPlugins = append(allPlugins, plugins...)
	}
	return allPlugins
}

// ListManyByType 根据 PluginType 返回多个插件
func ListManyByType(pluginType common.PluginType) []*common.Meta {
	return Registry[pluginType]
}

// GetMetaByName 根据插件名称获取 Meta
// 如果插件不存在，返回 nil
func GetMetaByName(name string) *common.Meta {
	// 遍历 Registry 中的所有插件类型
	for _, plugins := range Registry {
		// 遍历当前类型的所有插件
		for _, plugin := range plugins {
			if plugin.Name == name {
				// 找到匹配的插件，返回对应的 Meta
				return plugin
			}
		}
	}
	// 如果没有找到插件，返回 nil
	return nil
}
