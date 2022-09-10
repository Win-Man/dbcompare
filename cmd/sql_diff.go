/*
 * Created: 2022-09-10 11:48:20
 * Author : Win-Man
 * Email : gang.shen0423@gmail.com
 * -----
 * Last Modified:
 * Modified By:
 * -----
 * Description:
 */

package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/Win-Man/dbcompare/compare"
	"github.com/Win-Man/dbcompare/config"
	"github.com/Win-Man/dbcompare/database"
	"github.com/Win-Man/dbcompare/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newSqlDiffCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "sql-diff",
		Short: "sql-diff",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.InitConfig(configPath)
			logger.InitLogger(logLevel, logPath, cfg)
			log.Info("Welcome to sql-diff")
			log.Debug(fmt.Sprintf("Flags:%+v", cmd.Flags()))
			log.Debug(fmt.Sprintf("arguments:%s", strings.Join(args, ",")))

			executeSqlDiff(cfg)

			return nil
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "C", "", "config file path")
	cmd.Flags().StringVarP(&logLevel, "log-level", "L", "info", "log level: info, debug, warn, error, fatal")
	cmd.Flags().StringVar(&logPath, "log-path", "", "The path of log file")
	cmd.Flags().StringVar(&sqlString, "sql", "", "single sql statement")
	cmd.Flags().StringVar(&output, "output", "", "print|file")
	return cmd
}

func executeSqlDiff(cfg config.Config) {
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
	var sqlListSource []string
	var sqlListDest []string
	if sqlString != "" {
		sqlListSource = append(sqlListSource, sqlString)
		sqlListDest = append(sqlListDest, sqlString)
	} else if cfg.CompareConfig.SQLSource == "diff" && cfg.CompareConfig.SQLFileSource != "" && cfg.CompareConfig.SQLFileDest != "" {
		sqlListSource = ReadSQLFile(cfg.CompareConfig.SQLFileSource, cfg.CompareConfig.Delimiter)
		sqlListDest = ReadSQLFile(cfg.CompareConfig.SQLFileDest, cfg.CompareConfig.Delimiter)
	} else if cfg.CompareConfig.SQLSource == "default" && cfg.CompareConfig.SQLFileDefault != "" {
		sqlListSource = ReadSQLFile(cfg.CompareConfig.SQLFileDefault, cfg.CompareConfig.Delimiter)
		sqlListDest = ReadSQLFile(cfg.CompareConfig.SQLFileDefault, cfg.CompareConfig.Delimiter)
	} else {
		log.Error("Please input test sql statments!!!")
		os.Exit(1)
	}
	err = compare.CompareSelect(sourcedb, destdb, sqlListSource, sqlListDest, &cfg, false)
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
