package engine

import (
	"os"
)

// 所有和配置有关的信息都要通过config来进行操作
type Config struct {
	baseDirName         string //所有日志相关文件所在的目录
	ignoredHostFileName string //他是Searcher中要用到的忽略掉的主机源 文件名称
}

var config = &Config{baseDirName: "./", ignoredHostFileName: ".ignored_host_file"}

// 设置基目录，并且如果不存在会进行创建，但是如果因为你制定的目录需要超级权限，那么很可能会创建失败
// 所以有个error返回值
func (cfg *Config) SetBaseDirName(baseDirName string) error {
	cfg.baseDirName = baseDirName
	if len(cfg.baseDirName) > 0 {
		if _, err := os.Stat(cfg.baseDirName); err != nil {
			return os.MkdirAll(cfg.baseDirName, 0777)
		}
	}
	return nil
}

func (cfg *Config) BaseDirName() string {
	return cfg.baseDirName
}

func (cfg *Config) SetIgnoredHostFileName(ignoredHostFileName string) {
	cfg.ignoredHostFileName = ignoredHostFileName
}

func (cfg *Config) IgnoredHostFileName() string {
	return cfg.ignoredHostFileName
}
