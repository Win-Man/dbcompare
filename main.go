package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Win-Man/dbcompare/compare"
	"github.com/Win-Man/dbcompare/config"
	"github.com/Win-Man/dbcompare/database"
	"github.com/Win-Man/dbcompare/pkg/logger"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	Dbs []*sqlx.DB
)

var (
	configPath string
	logLevel   string
	logPath    string
	sqlString  string
	output     string
	h          bool
)

func init() {
	flag.StringVar(&configPath, "config", "", "config file path")
	flag.StringVar(&logLevel, "L", "info", "log level: info, debug, warn, error, fatal")
	flag.StringVar(&logPath, "log-path", "", "The path of log file")
	flag.StringVar(&sqlString, "sql", "", "single sql statement")
	flag.StringVar(&output, "output", "", "print|file")
	flag.BoolVar(&h, "h", false, "this help")

	flag.Parse()
}

func main() {
	flag.Parse()
	if h {
		flag.Usage()
		os.Exit(0)
	}

	// initialize config
	cfg := config.InitConfig(configPath)
	// initialize log
	logger.InitLogger(logLevel, logPath, cfg)
	log.Info("Welcome to dbcompare")
	log.Info(fmt.Sprintf("arguments:%v", flag.Args()))
	var sourcedb *sql.DB
	var destdb *sql.DB
	var err error
	switch strings.ToLower(cfg.CompareConfig.SourceType) {
	case "mysql":
		sourcedb, err = database.OpenMySQLDB(&cfg.MySQLConfig)
	case "tidb":
		sourcedb, err = database.OpenMySQLDB(&cfg.TiDBConfig)
	case "oracle":
		sourcedb, err = database.OpenOracleDB(&cfg.OracleConfig)
	default:
		log.Error(fmt.Sprintf("Unknow source type:%s", cfg.CompareConfig.SourceType))
	}
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
	}
	switch strings.ToLower(cfg.CompareConfig.DestType) {
	case "mysql":
		destdb, err = database.OpenMySQLDB(&cfg.MySQLConfig)
	case "tidb":
		destdb, err = database.OpenMySQLDB(&cfg.TiDBConfig)
	case "oracle":
		destdb, err = database.OpenOracleDB(&cfg.OracleConfig)
	default:
		log.Error(fmt.Sprintf("Unknow dest type:%s", cfg.CompareConfig.DestType))
	}
	if err != nil {
		log.Error(fmt.Sprintf("Connect dest database error:%v", err))
	}

	// output config
	if output != "" {
		cfg.CompareConfig.Output = output
	}

	// 从命令读取测试 SQL ，最高优先级，读取不到则读取配置文件中指定的 sql 文本地址
	var sqlList []string
	if sqlString != "" {
		sqlList = append(sqlList, sqlString)
	} else if cfg.CompareConfig.SQLFile != "" {
		sqlList = ReadSQLFile(cfg.CompareConfig.SQLFile, cfg.CompareConfig.Delimiter)
	} else {
		log.Error("Please input test sql statments!!!")
		os.Exit(1)
	}
	log.Debug("xxx")
	err = compare.CompareSelect(sourcedb, destdb, sqlList, &cfg)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func ReadSQLFile(filePath string, delimiter string) []string {
	var res []string
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		log.Error(err)
	}
	buf := bufio.NewScanner(f)
	var line string
	for {
		if !buf.Scan() {
			break
		}
		tmpstr := buf.Text()
		pos := strings.Index(tmpstr, delimiter)
		if pos != -1 {
			line = line + tmpstr[:pos]
			res = append(res, line)
			line = ""
		} else {
			line = line + tmpstr
		}
	}
	return res
}
