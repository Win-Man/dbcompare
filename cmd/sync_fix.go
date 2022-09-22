/*
 * Created: 2022-09-10 11:49:54
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
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/Win-Man/dbcompare/config"
	"github.com/Win-Man/dbcompare/database"
	"github.com/Win-Man/dbcompare/pkg/logger"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type CtlTemplate struct {
	Character         string
	FilePath          string
	BadFilePath       string
	DiscardFilePath   string
	TableOracleSchema string
	TableName         string
	Terminator        string
	Columns           string
}

func newSyncFixCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "sync-fix <dump-data|generate-ctl|load-data|all>",
		Short: "sync-fix",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Help()
			}
			cfg := config.InitSyncDiffConfig(configPath)
			logger.InitLogger(logLevel, logPath, cfg.Log)
			log.Info("Welcome to sync-fix")
			log.Debug(fmt.Sprintf("Flags:%+v", cmd.Flags()))
			log.Debug(fmt.Sprintf("arguments:%s", strings.Join(args, ",")))
			switch args[0] {
			case "dump-data":
				err := runDumpDataControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			case "generate-ctl":
				err := runGeneratorControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			case "load-data":
				err := runLoadControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			case "all":
				err := runDumpDataControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = runGeneratorControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = runLoadControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			default:
				return cmd.Help()
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "C", "", "config file path")
	cmd.Flags().StringVarP(&logLevel, "log-level", "L", "", "log level: info, debug, warn, error, fatal")
	return cmd
}

type DumpTableInfo struct {
	Id               int
	TableSchema      string
	TableName        string
	TableSchemaOrale string
}

func runDumpDataControl(cfg config.SyncDiffConfig) error {
	var err error
	err = os.MkdirAll(cfg.SyncFixConfig.DumpDataDir, 0755)
	if err != nil {
		return err
	}

	threadCount := cfg.SyncFixConfig.Concurrency
	tasks := make(chan DumpTableInfo, threadCount)
	var wg sync.WaitGroup
	handleCount = 0
	for i := 1; i <= threadCount; i++ {
		wg.Add(1)
		tmpi := i
		go func() {
			defer wg.Done()
			//log.Debug(fmt.Sprintf("go func i:%d", tmpi))
			runDumpData(cfg, tmpi, tasks)
			//testFunc(tmpi, tasks)
		}()
	}
	var db *sql.DB

	db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		return err
	}
	countQuerySql := fmt.Sprintf(`SELECT count(*) FROM %s.syncdiff_config_result where sync_status='%s'`, cfg.TiDBConfig.Database, CompareFailed)
	rows, err := db.Query(countQuerySql)
	if err != nil {
		log.Error(err)
		return err
	}
	for rows.Next() {
		rows.Scan(&tableCount)
	}

	querysql := fmt.Sprintf(`SELECT id,table_schema,table_name,ifnull(table_schema_oracle,table_schema) FROM %s.syncdiff_config_result where sync_status='%s'`, cfg.TiDBConfig.Database, CompareFailed)
	var dumpRow DumpTableInfo
	rows, err = db.Query(querysql)
	if err != nil {
		return err
	}
	for rows.Next() {
		rows.Scan(&dumpRow.Id, &dumpRow.TableSchema, &dumpRow.TableName, &dumpRow.TableSchemaOrale)
		tasks <- dumpRow
	}

	db.Close()
	close(tasks)
	// var db *sql.DB
	// var err error	// db, err = database.OpenMySQLDB(&cfg.TiDBConfig)

	// db.Close()
	wg.Wait()
	return nil
}

func runDumpData(cfg config.SyncDiffConfig, threadID int, tasks <-chan DumpTableInfo) {
	for task := range tasks {
		handleCount = handleCount + 1
		log.Info(fmt.Sprintf("[Thread-%d]Start to dump %s.%s data", threadID, task.TableSchema, task.TableName))
		log.Info(fmt.Sprintf("Process dump-data %d/%d", handleCount, tableCount))
		logPath := filepath.Join(cfg.Log.LogDir, fmt.Sprintf("dumpling_%s.%s.log", task.TableSchema, task.TableName))
		cmd := fmt.Sprintf("%s -u %s -P %d -h %s -p \"%s\" --filter \"%s.%s\" -o %s %s> %s", cfg.SyncFixConfig.DumplingBinPath, cfg.TiDBConfig.User,
			cfg.TiDBConfig.Port, cfg.TiDBConfig.Host, cfg.TiDBConfig.Password,
			task.TableSchema, task.TableName, cfg.SyncFixConfig.DumpDataDir,
			cfg.SyncFixConfig.DumpExtraArgs, logPath)
		c := exec.Command("bash", "-c", cmd)
		output, err := c.CombinedOutput()
		if err != nil {
			log.Error(fmt.Sprintf("Run command:%s failed. Check log:%s", cmd, logPath))
			log.Error(fmt.Sprintf("Run command stderr:%s", output))
			continue
		} else {
			log.Info(fmt.Sprintf("Run command:%s success.", cmd))
		}
		log.Info(fmt.Sprintf("[Thread-%d]Finished dump %s.%s data", threadID, task.TableSchema, task.TableName))
	}

}

func runGeneratorControl(cfg config.SyncDiffConfig) error {
	var err error
	err = os.MkdirAll(cfg.SyncFixConfig.OracleCtlFileDir, 0755)
	if err != nil {
		log.Error(err)
		return err
	}

	threadCount := cfg.SyncFixConfig.Concurrency
	tasks := make(chan DumpTableInfo, threadCount)
	var wg sync.WaitGroup
	handleCount = 0
	for i := 1; i <= threadCount; i++ {
		wg.Add(1)
		tmpi := i
		go func() {
			defer wg.Done()
			//log.Debug(fmt.Sprintf("go func i:%d", tmpi))
			runGenerator(cfg, tmpi, tasks)
			//testFunc(tmpi, tasks)
		}()
	}
	var db *sql.DB

	db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		return err
	}
	countQuerySql := fmt.Sprintf(`SELECT count(*) FROM %s.syncdiff_config_result where sync_status='%s'`, cfg.TiDBConfig.Database, CompareFailed)
	rows, err := db.Query(countQuerySql)
	if err != nil {
		log.Error(err)
		return err
	}
	for rows.Next() {
		rows.Scan(&tableCount)
	}

	querysql := fmt.Sprintf(`SELECT id,table_schema,table_name,ifnull(table_schema_oracle,table_schema) FROM %s.syncdiff_config_result where sync_status='%s'`, cfg.TiDBConfig.Database, CompareFailed)
	var dumpRow DumpTableInfo
	rows, err = db.Query(querysql)
	if err != nil {
		return err
	}
	for rows.Next() {
		rows.Scan(&dumpRow.Id, &dumpRow.TableSchema, &dumpRow.TableName, &dumpRow.TableSchemaOrale)
		tasks <- dumpRow
	}

	db.Close()
	close(tasks)
	// var db *sql.DB
	// var err error	// db, err = database.OpenMySQLDB(&cfg.TiDBConfig)

	// db.Close()
	wg.Wait()
	return nil
}

func runGenerator(cfg config.SyncDiffConfig, threadID int, tasks <-chan DumpTableInfo) {
	for task := range tasks {
		handleCount = handleCount + 1
		log.Info(fmt.Sprintf("Start to generate oracle sqlldr ctl file for %s.%s", task.TableSchema, task.TableName))
		log.Info(fmt.Sprintf("Process generate-ctl %d/%d", handleCount, tableCount))
		var db *sql.DB
		var err error
		db, err = database.OpenOracleDB(&cfg.OracleConfig)
		defer db.Close()
		if err != nil {
			log.Error(fmt.Sprintf("Connect source database error:%v", err))
			continue
		}
		querySql := fmt.Sprintf("select column_name from dba_tab_columns where owner='%s' and table_name='%s' order by column_id",
			strings.ToUpper(task.TableSchemaOrale), strings.ToUpper(task.TableName))
		log.Debug(querySql)
		rows, err := db.Query(querySql)
		if err != nil {
			log.Error(err)
			continue
		}
		colNames := make([]string, 0)
		for rows.Next() {
			var tmp string
			rows.Scan(&tmp)
			colNames = append(colNames, tmp)
		}
		tpl, err := template.ParseFiles(cfg.SyncFixConfig.CtlTemplate)
		if err != nil {
			log.Error(fmt.Sprintf("template parsefiles failed,err:%v", err))
			continue
		}

		ctltmpl := CtlTemplate{
			Character:         "UTF8",
			FilePath:          filepath.Join(cfg.SyncFixConfig.DumpDataDir, fmt.Sprintf("%s.%s.000000000.csv", task.TableSchema, task.TableName)),
			BadFilePath:       filepath.Join(cfg.SyncFixConfig.OracleCtlFileDir, fmt.Sprintf("%s.%s.bad", task.TableSchemaOrale, task.TableName)),
			DiscardFilePath:   filepath.Join(cfg.SyncFixConfig.OracleCtlFileDir, fmt.Sprintf("%s.%s.disc", task.TableSchemaOrale, task.TableName)),
			TableOracleSchema: task.TableSchemaOrale,
			TableName:         task.TableName,
			Columns:           strings.Join(colNames, ","),
		}
		ctlFilePath := filepath.Join(cfg.SyncFixConfig.OracleCtlFileDir, fmt.Sprintf("%s.%s.ctl", task.TableSchemaOrale, task.TableName))
		f, err := os.Create(ctlFilePath)
		defer f.Close()
		if err != nil {
			log.Error(err)
			continue
		}
		err = tpl.Execute(f, ctltmpl)
		if err != nil {
			log.Error(err)
			continue
		}

	}

}

func runLoadControl(cfg config.SyncDiffConfig) error {
	var err error
	err = os.MkdirAll(cfg.SyncFixConfig.DumpDataDir, 0755)
	if err != nil {
		log.Error(err)
		return err
	}

	threadCount := cfg.SyncFixConfig.Concurrency
	tasks := make(chan DumpTableInfo, threadCount)
	var wg sync.WaitGroup
	handleCount = 0
	for i := 1; i <= threadCount; i++ {
		wg.Add(1)
		tmpi := i
		go func() {
			defer wg.Done()
			runLoad(cfg, tmpi, tasks)

		}()
	}
	var db *sql.DB

	db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		return err
	}
	countQuerySql := fmt.Sprintf(`SELECT count(*) FROM %s.syncdiff_config_result where sync_status='%s'`, cfg.TiDBConfig.Database, CompareFailed)
	rows, err := db.Query(countQuerySql)
	if err != nil {
		log.Error(err)
		return err
	}
	for rows.Next() {
		rows.Scan(&tableCount)
	}
	querysql := fmt.Sprintf(`SELECT id,table_schema,table_name,ifnull(table_schema_oracle,table_schema) FROM %s.syncdiff_config_result where sync_status='%s'`, cfg.TiDBConfig.Database, CompareFailed)
	var dumpRow DumpTableInfo
	rows, err = db.Query(querysql)
	if err != nil {
		log.Error(err)
		return err
	}
	for rows.Next() {
		rows.Scan(&dumpRow.Id, &dumpRow.TableSchema, &dumpRow.TableName, &dumpRow.TableSchemaOrale)
		tasks <- dumpRow
	}

	db.Close()
	close(tasks)

	wg.Wait()
	return nil
}

func runLoad(cfg config.SyncDiffConfig, threadID int, tasks <-chan DumpTableInfo) error {
	for task := range tasks {
		handleCount = handleCount + 1
		if cfg.SyncFixConfig.TruncateBeforeLoad == true {
			var db *sql.DB
			var err error
			db, err = database.OpenOracleDB(&cfg.OracleConfig)
			defer db.Close()
			if err != nil {
				log.Error(fmt.Sprintf("Connect source database error:%v", err))
				continue
			}
			truncateSql := fmt.Sprintf("truncate table  %s.%s ",
				strings.ToUpper(task.TableSchemaOrale), strings.ToUpper(task.TableName))
			log.Debug(truncateSql)
			_, err = db.Exec(truncateSql)
			if err != nil {
				log.Error(err)
				continue
			}
			log.Info(fmt.Sprintf("[Thread-%d]Finished truncate table %s.%s", threadID, task.TableSchemaOrale, task.TableName))
		}

		log.Info(fmt.Sprintf("[Thread-%d]Start to sqlldr load data %s.%s", threadID, task.TableSchemaOrale, task.TableName))
		log.Info(fmt.Sprintf("Process load-data %d/%d", handleCount, tableCount))
		ctlFilePath := filepath.Join(cfg.SyncFixConfig.OracleCtlFileDir, fmt.Sprintf("%s.%s.ctl", task.TableSchemaOrale, task.TableName))
		logPath := filepath.Join(cfg.Log.LogDir, fmt.Sprintf("sqlldr_load_%s.%s.log", task.TableSchemaOrale, task.TableName))
		cmd := fmt.Sprintf("%s %s/%s control=%s > %s", cfg.SyncFixConfig.SqlldrBinPath, cfg.OracleConfig.User, cfg.OracleConfig.Password, ctlFilePath, logPath)
		c := exec.Command("bash", "-c", cmd)
		// cmdTest := fmt.Sprintf("%s %s", binPath, confPath)
		// c := exec.Command("bash", "-c", cmdTest)
		output, err := c.CombinedOutput()
		if err != nil {
			log.Error(fmt.Sprintf("Run command:%s failed. Check log:%s", cmd, logPath))
			log.Error(fmt.Sprintf("Run command stderr:%s", output))
			continue

		} else {
			log.Info(fmt.Sprintf("Run command:%s success.", cmd))

		}
	}

	return nil
}
