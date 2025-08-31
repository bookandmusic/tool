package common

import (
	"github.com/bookandmusic/tool/internal/logger"
)

var GlobalCfg *GlobalConfig

type GlobalConfig struct {
	Cfg     *Config
	Logger  logger.Logger
	CfgPath string
}
