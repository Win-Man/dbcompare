/*
 * Created: 2022-09-10 20:37:15
 * Author : Win-Man
 * Email : gang.shen0423@gmail.com
 * -----
 * Last Modified:
 * Modified By:
 * -----
 * Description:
 */

package config

import "github.com/BurntSushi/toml"

type OTOConfig struct {
	Log             Log                   `toml:"log" json:"log"`
	Performance     PerformanceConfig     `toml:"performance" json:"performance"`
	TiDBConfig      DBConfig              `toml:"tidb-config" json:"tidb-config"`
	OracleConfig    OracleDBConfig        `toml:"oracle-config" json:"oracle-config"`
	SyncDiffControl SyncDiffControlConfig `toml:"sync-diff-control-config" json:"sync-diff-control-config"`
	T2OInit         T2OInitConfig         `toml:"t2o-init-config" json:"t2o-init-config"`
}

type PerformanceConfig struct {
	Concurrency int `toml:"concurrency" json:"concurrency"`
}

type SyncDiffControlConfig struct {
	ConfDir      string `toml:"conf-dir" json:"conf-dir"`
	Concurrency  int    `toml:"concurrency" json:"concurrency"`
	BinPath      string `toml:"bin-path" json:"bin-path"`
	SyncTemplate string `toml:"sync-template" json:"sync-template"`
}

type T2OInitConfig struct {
	DumplingBinPath    string `toml:"dumpling-bin-path" json:"dumpling-bin-path"`
	DumpDataDir        string `toml:"dump-data-dir" json:"dump-data=dir"`
	Concurrency        int    `toml:"concurrency" json:"concurrency"`
	DumpExtraArgs      string `toml:"dump-extra-args" json:"dump-extra-args"`
	OracleCtlFileDir   string `toml:"oracle-ctl-file-dir" json:"oracle-ctl-file-dir"`
	CtlTemplate        string `toml:"ctl-template" json:"ctl-template"`
	SqlldrBinPath      string `toml:"sqlldr-bin-path" json:"sqlldr-bin-path"`
	TruncateBeforeLoad bool   `toml:"truncate-before-load" json:"truncate-before-load"`
}

// InitConfig Func
func InitOTOConfig(configPath string) (cfg OTOConfig) {

	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
