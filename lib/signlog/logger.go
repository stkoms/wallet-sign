package signlog

import (
	"os"

	logging "github.com/ipfs/go-log/v2"
)

// SetupLogLevels 初始化日志等级
func SetupLogLevels() {
	if _, set := os.LookupEnv("GOLOG_LOG_LEVEL"); !set {
		_ = logging.SetLogLevel("*", "INFO")

	}
}
