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

type SyncDiffConfig struct {
	Log           Log               `toml:"log" json:"log"`
	TiDBConfig    DBConfig          `toml:"tidb-config" json:"tidb-config"`
	OracleConfig  OracleDBConfig    `toml:"oracle-config" json:"oracle-config"`
	SyncCtlConfig SyncControlConfig `toml:"sync-ctl-config" json:"sync-ctl-config"`
	SyncFixConfig SyncFix           `toml:"sync-fix-config" json:"sync-fix-config"`
}

type SyncControlConfig struct {
	ConfDir      string `toml:"conf-dir" json:"conf-dir"`
	Concurrency  int    `toml:"concurrency" json:"concurrency"`
	BinPath      string `toml:"bin-path" json:"bin-path"`
	SyncTemplate string `toml:"sync-template" json:"sync-template"`
}

type SyncFix struct {
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
func InitSyncDiffConfig(configPath string) (cfg SyncDiffConfig) {

	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
