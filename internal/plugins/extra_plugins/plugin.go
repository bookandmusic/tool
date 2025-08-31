package extraplugins

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"

	"github.com/bookandmusic/tool/internal/common"
	"github.com/bookandmusic/tool/internal/plugins"
	"github.com/bookandmusic/tool/internal/service"
)

// LoadMeta 解析插件 meta.yml
func LoadMeta(dir string) (*common.Meta, error) {
	metaPath := filepath.Join(dir, "meta.yml")
	metaPath = filepath.Clean(metaPath)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}
	var m common.Meta
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	m.Dir = dir
	m.BuiltIn = false
	m.Service = &service.ExtraService{ExecDir: m.Dir, Exec: m.Exec, PluginName: m.Name, ExecType: m.ExecType}
	return &m, nil
}

func LoadAllExtraPluginMeta(cfg *common.Config) error {
	console := common.GlobalCfg.Logger
	for _, baseDir := range cfg.PluginDirs {
		info, err := os.Stat(baseDir)
		if err != nil {
			if os.IsNotExist(err) {
				console.Debug(fmt.Sprintf("[PLUGIN] Plugin directory does not exist: %s\n", baseDir))
				continue // 插件目录不存在，跳过
			}
			console.Warning("[PLUGIN] Cannot stat %s: %v", baseDir, err)
			continue
		}
		if !info.IsDir() {
			console.Debug(fmt.Sprintf("[PLUGIN] Not a directory, skipping: %s\n", baseDir))
			continue // 不是目录，跳过
		}

		entries, err := os.ReadDir(baseDir)
		if err != nil {
			console.Warning("[PLUGIN] Cannot read directory %s: %v", baseDir, err)
			continue
		}

		for _, e := range entries {
			if !e.IsDir() {
				console.Debug(fmt.Sprintf("[PLUGIN] Skipping non-directory entry: %s\n", e.Name()))
				continue
			}
			dir := filepath.Join(baseDir, e.Name())

			meta, err := LoadMeta(dir)
			if err != nil {
				console.Warning("[PLUGIN] Skipping %s due to load error: %v", dir, err)
				continue
			}

			plugins.RegisterMeta(meta)
			console.Debug(fmt.Sprintf("[PLUGIN] Loaded plugin: %s\n", meta.Name))
		}
	}
	return nil
}
