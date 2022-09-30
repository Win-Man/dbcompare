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
	SyncDiffControl SyncDiffControlConfig `toml:"sync-diff-control" json:"sync-diff-control"`
	T2OInit         T2OInitConfig         `toml:"t2o-init" json:"t2o-init"`
	O2TInit         O2TInitConfig         `toml:"o2t-init" json:"o2t-init"`
}

type PerformanceConfig struct {
	Concurrency   int  `toml:"concurrency" json:"concurrency"`
	CheckRowCount bool `toml:"check-row-count" json:"check-row-count"`
}

type SyncDiffControlConfig struct {
	ConfDir      string `toml:"conf-dir" json:"conf-dir"`
	BinPath      string `toml:"bin-path" json:"bin-path"`
	SyncTemplate string `toml:"sync-template" json:"sync-template"`
}

type T2OInitConfig struct {
	DumplingBinPath    string `toml:"dumpling-bin-path" json:"dumpling-bin-path"`
	DumpDataDir        string `toml:"dump-data-dir" json:"dump-data=dir"`
	DumpExtraArgs      string `toml:"dump-extra-args" json:"dump-extra-args"`
	OracleCtlFileDir   string `toml:"oracle-ctl-file-dir" json:"oracle-ctl-file-dir"`
	CtlTemplate        string `toml:"ctl-template" json:"ctl-template"`
	SqlldrBinPath      string `toml:"sqlldr-bin-path" json:"sqlldr-bin-path"`
	SqlldrExtraArgs    string `toml:"sqlldr-extra-args" json:"sqlldr-extra-args"`
	TruncateBeforeLoad bool   `toml:"truncate-before-load" json:"truncate-before-load"`
}

type O2TInitConfig struct {
	Sqluldr2BinPath       string `toml:"sqluldr2-bin-path" json:"sqluldr2-bin-path"`
	Sqluldr2ExtraArgs     string `toml:"sqluldr2-extra-args" json:"sqluldr2-extra-args"`
	DumpDataDir           string `toml:"dump-data-dir" json:"dump-data-dir"`
	LightningBinPath      string `toml:"lightning-bin-path" json:"lightning-bin-path"`
	LightningTomlTemplate string `toml:"lightning-toml-template" json:"lightning-toml-template"`
	LightningTomlDir      string `toml:"lightning-toml-dir" json:"lightning-toml-dir"`
	LightningExtraArgs    string `toml:"lightning-extra-args" json:"lightning-extra-args"`
}

// InitConfig Func
func InitOTOConfig(configPath string) (cfg OTOConfig) {

	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
