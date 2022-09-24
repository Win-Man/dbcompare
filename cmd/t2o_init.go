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
	"time"

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

func newT2OInitCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "t2o-init <prepare|dump-data|generate-conf|load-data|all>",
		Short: "t2o-init",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Help()
			}
			cfg := config.InitOTOConfig(configPath)
			logger.InitLogger(logLevel, logPath, cfg.Log)
			log.Info("Welcome to t2o-init")
			log.Debug(fmt.Sprintf("Flags:%+v", cmd.Flags()))
			log.Debug(fmt.Sprintf("arguments:%s", strings.Join(args, ",")))
			switch args[0] {
			case "prepare":
				err := createT2OInitMeta(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				log.Info("Finished prepare without errors.")
				fmt.Printf("Create table success.\nPls init table data,refer sql:\n")
				fmt.Printf(`INSERT INTO %s.t2o_config(
					table_schema_tidb,table_name_tidb,table_schema_oracle,
					dump_status,generate_ctl_status,load_status) 
					VALUES ('mydb','mytab','ordb','%s','%s','%s')`,
					cfg.TiDBConfig.Database, StatusWaiting, StatusWaiting, StatusWaiting)
			case "dump-data":
				err := runT2ODumpDataControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			case "generate-conf":
				err := runT2OGeneratorControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			case "load-data":
				err := runT2OLoadControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			case "all":
				err := runT2ODumpDataControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = runT2OGeneratorControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = runT2OLoadControl(cfg)
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

func createT2OInitMeta(cfg config.OTOConfig) error {
	log.Info("Start to create t2o_config table")
	tableSql := `
	CREATE TABLE IF NOT EXISTS t2o_config(
		id int(11) NOT NULL AUTO_INCREMENT,
		table_schema_tidb varchar(64) NOT NULL,
		table_name_tidb varchar(64) NOT NULL,
		table_schema_oracle varchar(64) ,
		dump_status varchar(20) NOT NULL,
		dump_duration int(11),
		last_dump_time datetime,
		generate_ctl_status varchar(20) NOT NULL,
		generate_ctl_duration int(11),
		last_generate_ctl_time datetime,
		load_status varchar(20) NOT NULL,
		load_duration int(11),
		last_load_duration datetime,
		PRIMARY KEY(id),
		UNIQUE KEY uk_tab(table_schema_tidb,table_name_tidb)
	)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`
	var db *sql.DB
	var err error
	db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
	defer db.Close()
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		return err
	}
	_, err = db.Exec(tableSql)
	if err != nil {
		log.Error("Create table t2o_config failed!")
		log.Error(fmt.Sprintf("Execute statement error:%v", err))
		return err
	}
	return nil
}

func runT2ODumpDataControl(cfg config.OTOConfig) error {
	var err error
	err = os.MkdirAll(cfg.T2OInit.DumpDataDir, 0755)
	if err != nil {
		return err
	}

	threadCount := cfg.Performance.Concurrency
	tasks := make(chan DumpTableInfo, threadCount)
	var wg sync.WaitGroup
	handleCount = 0
	for i := 1; i <= threadCount; i++ {
		wg.Add(1)
		tmpi := i
		go func() {
			defer wg.Done()
			//log.Debug(fmt.Sprintf("go func i:%d", tmpi))
			runT2ODumpData(cfg, tmpi, tasks)
			//testFunc(tmpi, tasks)
		}()
	}
	var db *sql.DB

	db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		return err
	}
	countQuerySql := fmt.Sprintf(`SELECT count(*) FROM %s.t2o_config where dump_status='%s'`, cfg.TiDBConfig.Database, StatusWaiting)
	rows, err := db.Query(countQuerySql)
	if err != nil {
		log.Error(err)
		return err
	}
	for rows.Next() {
		rows.Scan(&tableCount)
	}

	querysql := fmt.Sprintf(`SELECT id,table_schema_tidb,table_name_tidb,ifnull(table_schema_oracle,table_schema_tidb) FROM %s.t2o_config where dump_status='%s'`,
		cfg.TiDBConfig.Database, StatusWaiting)
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
	wg.Wait()
	return nil
}

func runT2ODumpData(cfg config.OTOConfig, threadID int, tasks <-chan DumpTableInfo) {
	for task := range tasks {
		handleCount = handleCount + 1
		dumpStartTime := time.Now()
		log.Info(fmt.Sprintf("[Thread-%d]Start to dump %s.%s data", threadID, task.TableSchema, task.TableName))
		log.Info(fmt.Sprintf("Process dump-data %d/%d", handleCount, tableCount))
		stdLogPath := filepath.Join(cfg.Log.LogDir, fmt.Sprintf("dumpling_%s.%s.log", task.TableSchema, task.TableName))
		cmd := fmt.Sprintf("%s -u %s -P %d -h %s -p \"%s\" --filter \"%s.%s\" -o %s %s> %s 2>&1", cfg.T2OInit.DumplingBinPath, cfg.TiDBConfig.User,
			cfg.TiDBConfig.Port, cfg.TiDBConfig.Host, cfg.TiDBConfig.Password,
			task.TableSchema, task.TableName, cfg.T2OInit.DumpDataDir,
			cfg.T2OInit.DumpExtraArgs, stdLogPath)
		c := exec.Command("bash", "-c", cmd)
		output, err := c.CombinedOutput()
		dumpEndTime := time.Now()
		dumpDuration := int(dumpEndTime.Sub(dumpStartTime).Seconds())
		var updateStatusSql string
		if err != nil {
			log.Error(fmt.Sprintf("Run command:%s failed. Check log:%s", cmd, logPath))
			log.Error(fmt.Sprintf("Run command stderr:%s", output))
			updateStatusSql = fmt.Sprintf(`update %s.t2o_config set
				dump_status = '%s',dump_duration = %d,last_dump_time = '%s' where id = %d`,
				cfg.TiDBConfig.Database, StatusFailed, dumpDuration, dumpStartTime.Format("2006-01-02 15:04:05"), task.Id)
		} else {
			log.Info(fmt.Sprintf("Run command:%s success.", cmd))
			updateStatusSql = fmt.Sprintf(`update %s.t2o_config set
				dump_status = '%s',dump_duration = %d,last_dump_time = '%s' where id = %d`,
				cfg.TiDBConfig.Database, StatusSuccess, dumpDuration, dumpStartTime.Format("2006-01-02 15:04:05"), task.Id)
		}
		var db *sql.DB
		db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
		defer db.Close()
		if err != nil {
			log.Error(fmt.Sprintf("Connect source database error:%v", err))
			log.Error("[Thread-%d]Update dump %s.%s status failed", threadID, task.TableSchema, task.TableName)
		}
		_, err = db.Exec(updateStatusSql)
		if err != nil {
			log.Error(fmt.Sprintf("Execute statement error:%v", err))
			log.Error("[Thread-%d]Update dump %s.%s status failed", threadID, task.TableSchema, task.TableName)
		}
		log.Info(fmt.Sprintf("[Thread-%d]Finished dump %s.%s data", threadID, task.TableSchema, task.TableName))
	}

}

func runT2OGeneratorControl(cfg config.OTOConfig) error {
	var err error
	err = os.MkdirAll(cfg.T2OInit.OracleCtlFileDir, 0755)
	if err != nil {
		log.Error(err)
		return err
	}

	threadCount := cfg.Performance.Concurrency
	tasks := make(chan DumpTableInfo, threadCount)
	var wg sync.WaitGroup
	handleCount = 0
	for i := 1; i <= threadCount; i++ {
		wg.Add(1)
		tmpi := i
		go func() {
			defer wg.Done()
			//log.Debug(fmt.Sprintf("go func i:%d", tmpi))
			runT2OGenerator(cfg, tmpi, tasks)
			//testFunc(tmpi, tasks)
		}()
	}
	var db *sql.DB

	db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		return err
	}
	countQuerySql := fmt.Sprintf(`SELECT count(*) FROM %s.t2o_config where generate_ctl_status='%s'`, cfg.TiDBConfig.Database, StatusWaiting)
	rows, err := db.Query(countQuerySql)
	if err != nil {
		log.Error(err)
		return err
	}
	for rows.Next() {
		rows.Scan(&tableCount)
	}

	querysql := fmt.Sprintf(`SELECT id,table_schema_tidb,table_name_tidb,ifnull(table_schema_oracle,table_schema_tidb) FROM %s.t2o_config where generate_ctl_status='%s'`,
		cfg.TiDBConfig.Database, StatusWaiting)
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

	wg.Wait()
	return nil
}

func runT2OGenerator(cfg config.OTOConfig, threadID int, tasks <-chan DumpTableInfo) {
	for task := range tasks {
		handleCount = handleCount + 1
		generateStartTime := time.Now()
		log.Info(fmt.Sprintf("[Thread-%d]Start to generate oracle sqlldr ctl file for %s.%s", threadID, task.TableSchema, task.TableName))
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
		tpl, err := template.ParseFiles(cfg.T2OInit.CtlTemplate)
		if err != nil {
			log.Error(fmt.Sprintf("template parsefiles failed,err:%v", err))
			continue
		}

		ctltmpl := CtlTemplate{
			Character:         "UTF8",
			FilePath:          filepath.Join(cfg.T2OInit.DumpDataDir, fmt.Sprintf("%s.%s.000000000.csv", task.TableSchema, task.TableName)),
			BadFilePath:       filepath.Join(cfg.T2OInit.OracleCtlFileDir, fmt.Sprintf("%s.%s.bad", task.TableSchemaOrale, task.TableName)),
			DiscardFilePath:   filepath.Join(cfg.T2OInit.OracleCtlFileDir, fmt.Sprintf("%s.%s.disc", task.TableSchemaOrale, task.TableName)),
			TableOracleSchema: task.TableSchemaOrale,
			TableName:         task.TableName,
			Columns:           strings.Join(colNames, ","),
		}
		ctlFilePath := filepath.Join(cfg.T2OInit.OracleCtlFileDir, fmt.Sprintf("%s.%s.ctl", task.TableSchemaOrale, task.TableName))
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
		generateEndTime := time.Now()
		generateDuration := int(generateEndTime.Sub(generateStartTime).Seconds())
		var updateStatusSql string
		updateStatusSql = fmt.Sprintf(`update %s.t2o_config set
				generate_ctl_status = '%s',generate_ctl_duration = %d,last_generate_ctl_time = '%s' where id = %d`,
			cfg.TiDBConfig.Database, StatusSuccess, generateDuration, generateStartTime.Format("2006-01-02 15:04:05"), task.Id)
		db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
		defer db.Close()
		if err != nil {
			log.Error(fmt.Sprintf("Connect source database error:%v", err))
			log.Error("[Thread-%d]Update generate ctl %s.%s status failed", threadID, task.TableSchema, task.TableName)
		}
		_, err = db.Exec(updateStatusSql)
		if err != nil {
			log.Error(fmt.Sprintf("Execute statement error:%v", err))
			log.Error("[Thread-%d]Update generate ctl %s.%s status failed", threadID, task.TableSchema, task.TableName)
		}
		log.Info(fmt.Sprintf("[Thread-%d]Finished generate ctl for %s.%s", threadID, task.TableSchema, task.TableName))
	}

}

func runT2OLoadControl(cfg config.OTOConfig) error {
	var err error
	err = os.MkdirAll(cfg.T2OInit.DumpDataDir, 0755)
	if err != nil {
		log.Error(err)
		return err
	}

	threadCount := cfg.Performance.Concurrency
	tasks := make(chan DumpTableInfo, threadCount)
	var wg sync.WaitGroup
	handleCount = 0
	for i := 1; i <= threadCount; i++ {
		wg.Add(1)
		tmpi := i
		go func() {
			defer wg.Done()
			runT2OLoad(cfg, tmpi, tasks)

		}()
	}
	var db *sql.DB

	db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		return err
	}
	countQuerySql := fmt.Sprintf(`SELECT count(*) FROM %s.t2o_config where load_status='%s'`, cfg.TiDBConfig.Database, StatusWaiting)
	rows, err := db.Query(countQuerySql)
	if err != nil {
		log.Error(err)
		return err
	}
	for rows.Next() {
		rows.Scan(&tableCount)
	}
	querysql := fmt.Sprintf(`SELECT id,table_schema,table_name,ifnull(table_schema_oracle,table_schema) FROM %s.t2o_config where load_status='%s'`, cfg.TiDBConfig.Database, StatusWaiting)
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

func runT2OLoad(cfg config.OTOConfig, threadID int, tasks <-chan DumpTableInfo) error {
	for task := range tasks {
		handleCount = handleCount + 1
		loadStartTime := time.Now()
		if cfg.T2OInit.TruncateBeforeLoad == true {
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
		ctlFilePath := filepath.Join(cfg.T2OInit.OracleCtlFileDir, fmt.Sprintf("%s.%s.ctl", task.TableSchemaOrale, task.TableName))
		sqlldrLogPath := filepath.Join(cfg.T2OInit.OracleCtlFileDir, fmt.Sprintf("%s.%s.log", task.TableSchemaOrale, task.TableName))
		stdLogPath := filepath.Join(cfg.Log.LogDir, fmt.Sprintf("sqlldr_load_%s.%s.log", task.TableSchemaOrale, task.TableName))
		cmd := fmt.Sprintf("%s %s/%s control=%s log=%s %s> %s 2>&1", cfg.T2OInit.SqlldrBinPath, cfg.OracleConfig.User, cfg.OracleConfig.Password, ctlFilePath, sqlldrLogPath, cfg.T2OInit.SqlldrExtraArgs, stdLogPath)
		c := exec.Command("bash", "-c", cmd)
		output, err := c.CombinedOutput()
		loadEndTime := time.Now()
		loadDuration := int(loadEndTime.Sub(loadStartTime).Seconds())
		var updateStatusSql string
		if err != nil {
			log.Error(fmt.Sprintf("Run command:%s failed. Check log:%s", cmd, logPath))
			log.Error(fmt.Sprintf("Run command stderr:%s", output))
			updateStatusSql = fmt.Sprintf(`update %s.t2o_config set
				load_status = '%s',load_duration = %d,last_load_time = '%s' where id = %d`,
				cfg.TiDBConfig.Database, StatusFailed, loadDuration, loadStartTime.Format("2006-01-02 15:04:05"), task.Id)

		} else {
			log.Info(fmt.Sprintf("Run command:%s success.", cmd))
			updateStatusSql = fmt.Sprintf(`update %s.t2o_config set
				load_status = '%s',load_duration = %d,last_load_time = '%s' where id = %d`,
				cfg.TiDBConfig.Database, StatusSuccess, loadDuration, loadStartTime.Format("2006-01-02 15:04:05"), task.Id)
		}
		var db *sql.DB
		db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
		defer db.Close()
		if err != nil {
			log.Error(fmt.Sprintf("Connect source database error:%v", err))
			log.Error("[Thread-%d]Update load %s.%s status failed", threadID, task.TableSchema, task.TableName)
		}
		_, err = db.Exec(updateStatusSql)
		if err != nil {
			log.Error(fmt.Sprintf("Execute statement error:%v", err))
			log.Error("[Thread-%d]Update load %s.%s status failed", threadID, task.TableSchema, task.TableName)
		}
		log.Info(fmt.Sprintf("[Thread-%d]Finished load %s.%s data", threadID, task.TableSchema, task.TableName))
	}

	return nil
}
