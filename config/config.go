/*
 * Created: 2021-03-25 14:34:55
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

type Config struct {
	Log           Log            `toml:"log" json:"log"`
	MySQLConfig   DBConfig       `toml:"mysql-config" json:"mysql-config"`
	TiDBConfig    DBConfig       `toml:"tidb-config" json:"tidb-config"`
	CompareConfig CompareConfig  `toml:"compare-config" json:"compare-config"`
	OracleConfig  OracleDBConfig `toml:"oracle-config" json:"oracle-config"`
}

type Log struct {
	Level   string `toml:"log-level" json:"log-level"`
	LogPath string `toml:"log-path" json:"log-path"`
	LogDir  string `toml:"log-dir" json:"log-dir"`
}

type DBConfig struct {
	Host       string `toml:"host" json:"host"`
	Port       int    `toml:"port" json:"port"`
	User       string `toml:"user" json:"user"`
	Password   string `toml:"password" json:"password"`
	Database   string `toml:"database" json:"database"`
	StatusPort int    `toml:"status-port" json:"status-port"`
	PDAddr     string `toml:"pd-addr" json:"pd-addr"`
}

type OracleDBConfig struct {
	User        string `toml:"user" json:"user"`
	Password    string `toml:"password" json:"password"`
	Host        string `toml:"host" json:"host"`
	Port        int    `toml:"port" json:"port"`
	ServiceName string `toml:"service-name" json:"service-name"`
	SchemaName  string `toml:"schema-name" json:"schema-name"`
}

// type TiDBConfig struct{
// 	Host string `toml:"host" json:"host"`
// 	Port int `toml:"port" json:"port"`
// 	User string `toml:"user" json:"user"`
// 	Password string `toml:"password" json:"password"`
// 	Database string `toml:"database" json:"database"`
// }

type CompareConfig struct {
	SourceType     string `toml:"source-type" json:"source-type"`
	DestType       string `toml:"dest-type" json:"dest-type"`
	SQLSource      string `toml:"sql-source" json:"sql-source"`
	SQLFileDefault string `toml:"sqlfile-default" json:"sqlfile-default"`
	SQLFileSource  string `toml:"sqlfile-source" json:"sqlfile-source"`
	SQLFileDest    string `toml:"sqlfile-dest" json:"sqlfile-dest"`
	Delimiter      string `toml:"sqlfile-delimiter" json:"sqlfile-delimiter"`
	Output         string `toml:"output" json:"output"`
	OutputPrefix   string `toml:"outputprefix" json:"outputprefix"`
}

// InitConfig Func
func InitConfig(configPath string) (cfg Config) {

	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
