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

func newO2TInitCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "o2t-init <prepare|dump-data|generate-conf|load-data|all>",
		Short: "o2t-init",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Help()
			}
			cfg := config.InitOTOConfig(configPath)
			logger.InitLogger(logLevel, logPath, cfg.Log)
			log.Info("Welcome to o2t-init")
			log.Debug(fmt.Sprintf("Flags:%+v", cmd.Flags()))
			log.Debug(fmt.Sprintf("arguments:%s", strings.Join(args, ",")))
			switch args[0] {
			case "prepare":
				err := createO2TInitMeta(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				log.Info("Finished prepare without errors.")
				fmt.Printf("Create table success.\nPls init table data,refer sql:\n")
				fmt.Printf("INSERT INTO %s.o2t_config() VALUES ", cfg.TiDBConfig.Database)
			case "dump-data":
				err := runO2TDumpDataControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			case "generate-conf":
				err := runO2TGenerateConfControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			case "load-data":
				err := runO2TLoadDataControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			case "all":
				err := runO2TDumpDataControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = runO2TGenerateConfControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = runO2TLoadDataControl(cfg)
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

// type DumpTableInfo struct {
// 	Id               int
// 	TableSchema      string
// 	TableName        string
// 	TableSchemaOrale string
// }

type LightingTomlTemplate struct {
	BatchID     string
	DumpDataDir string
	TiDBDB      config.DBConfig
}

func createO2TInitMeta(cfg config.OTOConfig) error {
	log.Info("Start to create o2t_config table")
	tableSql := `
	CREATE TABLE IF NOT EXISTS o2t_config(
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
		log.Error("Create table o2t_config failed!")
		log.Error(fmt.Sprintf("Execute statement error:%v", err))
		return err
	}
	return nil
}

func runO2TDumpDataControl(cfg config.OTOConfig) error {
	var err error
	err = os.MkdirAll(cfg.O2TInit.DumpDataDir, 0755)
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
			runO2TDumpData(cfg, tmpi, tasks)
			//testFunc(tmpi, tasks)
		}()
	}
	var db *sql.DB

	db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		return err
	}
	countQuerySql := fmt.Sprintf(`SELECT count(*) FROM %s.o2t_config where dump_status='%s'`, cfg.TiDBConfig.Database, StatusWaiting)
	rows, err := db.Query(countQuerySql)
	if err != nil {
		log.Error(err)
		return err
	}
	for rows.Next() {
		rows.Scan(&tableCount)
	}

	querysql := fmt.Sprintf(`SELECT id,table_schema,table_name,ifnull(table_schema_oracle,table_schema) FROM %s.o2t_config where dump_status='%s'`, cfg.TiDBConfig.Database, StatusWaiting)
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

func runO2TDumpData(cfg config.OTOConfig, threadID int, tasks <-chan DumpTableInfo) {
	for task := range tasks {
		handleCount = handleCount + 1
		dumpStartTime := time.Now()
		log.Info(fmt.Sprintf("[Thread-%d]Start to dump %s.%s data", threadID, task.TableSchema, task.TableName))
		log.Info(fmt.Sprintf("Process dump-data %d/%d", handleCount, tableCount))
		stdLogPath := filepath.Join(cfg.Log.LogDir, fmt.Sprintf("sqlldr2_%s.%s.log", task.TableSchema, task.TableName))
		dataPath := filepath.Join(cfg.O2TInit.DumpDataDir, fmt.Sprintf("%s.%s.%%B.csv",
			task.TableSchema, task.TableName))
		cmd := fmt.Sprintf("%s user=%s/%s@%s query=%s.%s file=%s %s > %s 2>&1",
			cfg.O2TInit.Sqlldr2BinPath, cfg.OracleConfig.User, cfg.OracleConfig.Password,
			cfg.OracleConfig.ServiceName, strings.ToUpper(task.TableSchemaOrale), strings.ToUpper(task.TableName),
			dataPath, cfg.O2TInit.Sqlldr2ExtraArgs, stdLogPath)
		c := exec.Command("bash", "-c", cmd)
		output, err := c.CombinedOutput()
		dumpEndTime := time.Now()
		dumpDuration := int(dumpEndTime.Sub(dumpStartTime).Seconds())
		var updateStatusSql string
		if err != nil {
			log.Error(fmt.Sprintf("Run command:%s failed. Check log:%s", cmd, logPath))
			log.Error(fmt.Sprintf("Run command stderr:%s", output))
			updateStatusSql = fmt.Sprintf(`update %s.o2t_config set
				dump_status = '%s',dump_duration = %d,last_dump_time = '%s' where id = %d`,
				cfg.TiDBConfig.Database, StatusFailed, dumpDuration, dumpStartTime.Format("2006-01-02 15:04:05"), task.Id)
		} else {
			log.Info(fmt.Sprintf("Run command:%s success.", cmd))
			updateStatusSql = fmt.Sprintf(`update %s.o2t_config set
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

func runO2TGenerateConfControl(cfg config.OTOConfig) error {
	var err error
	err = os.MkdirAll(cfg.T2OInit.OracleCtlFileDir, 0755)
	if err != nil {
		log.Error(err)
		return err
	}
	return runO2TGenerateConf(cfg)

}

func runO2TGenerateConf(cfg config.OTOConfig) error {
	log.Info(fmt.Sprintf("Start to generate tidb-lightning.toml "))
	generateStartTime := time.Now()
	tpl, err := template.ParseFiles(cfg.O2TInit.LightningTomlTemplate)
	if err != nil {
		log.Error(fmt.Sprintf("template parsefiles failed,err:%v", err))
	}
	batchid = time.Now().Format("20060102150405")
	lightingToml := LightingTomlTemplate{
		BatchID:     batchid,
		DumpDataDir: cfg.O2TInit.DumpDataDir,
		TiDBDB:      cfg.TiDBConfig,
	}
	lightningTomlDir, _ := filepath.Split(cfg.O2TInit.LightningTomlTemplate)

	lightningTomlPath := filepath.Join(lightningTomlDir, "tidb-lightning.toml")
	f, err := os.Create(lightningTomlPath)
	defer f.Close()
	if err != nil {
		log.Error(err)
		return err
	}
	err = tpl.Execute(f, lightingToml)
	if err != nil {
		log.Error(err)
		return err
	}
	generateEndTime := time.Now()
	generateDuration := int(generateEndTime.Sub(generateStartTime).Seconds())
	var updateStatusSql string
	updateStatusSql = fmt.Sprintf(`update %s.o2t_config set
				generate_ctl_status = '%s',generate_ctl_duration = %d,last_generate_ctl_time = '%s' where dump_status='%s'`,
		cfg.TiDBConfig.Database, StatusSuccess, generateDuration, generateStartTime.Format("2006-01-02 15:04:05"), StatusSuccess)
	db, err := database.OpenMySQLDB(&cfg.TiDBConfig)
	defer db.Close()
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		log.Error("Update generate tidb-lightning.toml status failed")
	}
	_, err = db.Exec(updateStatusSql)
	if err != nil {
		log.Error(fmt.Sprintf("Execute statement error:%v", err))
		log.Error("Update generate tidb-lightning.toml status failed")
	}
	log.Info("Finished generate tidb-lighting.toml")
	return nil
}

func runO2TLoadDataControl(cfg config.OTOConfig) error {

	return runO2TLoadData(cfg)
}

func runO2TLoadData(cfg config.OTOConfig) error {

	log.Info(fmt.Sprintf("Start to load data by tidb-lightning "))
	loadStartTime := time.Now()

	lightningTomlDir, _ := filepath.Split(cfg.O2TInit.LightningTomlTemplate)

	lightningTomlPath := filepath.Join(lightningTomlDir, "tidb-lightning.toml")
	lightningStdPath := filepath.Join(lightningTomlDir, "tidb_lighting_stderr.log")
	cmd := fmt.Sprintf("%s -config %s %s > %s 2>&1", cfg.O2TInit.LightningBinPath, lightningTomlPath,
		cfg.O2TInit.LightningExtraArgs, lightningStdPath)
	c := exec.Command("bash", "-c", cmd)
	output, err := c.CombinedOutput()
	loadEndTime := time.Now()
	loadDuration := int(loadEndTime.Sub(loadStartTime).Seconds())
	var updateStatusSql string
	if err != nil {
		log.Error(fmt.Sprintf("Run command:%s failed. Check log:%s", cmd, logPath))
		log.Error(fmt.Sprintf("Run command stderr:%s", output))
		updateStatusSql = fmt.Sprintf(`update %s.o2t_config set
				load_status = '%s',load_duration = %d,last_load_time = '%s' where dump_status='%s'`,
			cfg.TiDBConfig.Database, StatusFailed, loadDuration, loadStartTime.Format("2006-01-02 15:04:05"), StatusSuccess)
	} else {
		log.Info(fmt.Sprintf("Run command:%s success.", cmd))
		updateStatusSql = fmt.Sprintf(`update %s.o2t_config set
				load_status = '%s',load_duration = %d,last_load_time = '%s' where dump_status='%s'`,
			cfg.TiDBConfig.Database, StatusSuccess, loadDuration, loadStartTime.Format("2006-01-02 15:04:05"), StatusSuccess)
	}

	db, err := database.OpenMySQLDB(&cfg.TiDBConfig)
	defer db.Close()
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		log.Error("Update tidb-lightning load status failed")
	}
	_, err = db.Exec(updateStatusSql)
	if err != nil {
		log.Error(fmt.Sprintf("Execute statement error:%v", err))
		log.Error("Update tidb-lightning status failed")
	}
	log.Info("Finished load data by tidb-lightning")
	return nil
}
