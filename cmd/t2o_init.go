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
	"github.com/Win-Man/dbcompare/models"
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
				err := database.InitDB(cfg.TiDBConfig)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = createT2OInitMeta(cfg)
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
				err := database.InitDB(cfg.TiDBConfig)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = runT2ODumpDataControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			case "generate-conf":
				err := database.InitDB(cfg.TiDBConfig)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = runT2OGeneratorControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			case "load-data":
				err := database.InitDB(cfg.TiDBConfig)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = runT2OLoadControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			case "all":
				err := database.InitDB(cfg.TiDBConfig)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = runT2ODumpDataControl(cfg)
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

func createT2OInitMeta(cfg config.OTOConfig) error {
	log.Info("Start to create t2o_config table")

	if !database.DB.Migrator().HasTable(&models.T2OConfigModel{}) {
		err := database.DB.Migrator().CreateTable(&models.T2OConfigModel{})
		if err != nil {
			log.Error(fmt.Sprintf("Create table %s.t2o_config failed!Error:%v", cfg.TiDBConfig.Database, err))
			return err
		}
	} else {
		err := database.DB.AutoMigrate(&models.T2OConfigModel{})
		if err != nil {
			log.Error(fmt.Sprintf("Migrate table %s.t2o_config failed!Error:%v", cfg.TiDBConfig.Database, err))
			return err
		}
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
	tasks := make(chan models.T2OConfigModel, threadCount)
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

	res := database.DB.Model(&models.T2OConfigModel{}).Where("dump_status = ?", StatusWaiting).Count(&tableCount)
	if res.Error != nil {
		log.Error(res.Error)
	}
	var records []models.T2OConfigModel
	res = database.DB.Model(&models.T2OConfigModel{}).Where("dump_status = ?", StatusWaiting).Scan(&records)
	if res.Error != nil {
		log.Error(res.Error)
	}
	for _, record := range records {
		tasks <- record
	}

	close(tasks)
	wg.Wait()
	return nil
}

func runT2ODumpData(cfg config.OTOConfig, threadID int, tasks <-chan models.T2OConfigModel) {
	for task := range tasks {
		handleCount = handleCount + 1
		dumpStartTime := time.Now()
		log.Info(fmt.Sprintf("[Thread-%d]Start to dump %s.%s data", threadID, task.TableSchemaTidb, task.TableNameTidb))
		log.Info(fmt.Sprintf("Process dump-data %d/%d", handleCount, tableCount))
		stdLogPath := filepath.Join(cfg.Log.LogDir, fmt.Sprintf("dumpling_%s.%s.log", task.TableSchemaTidb, task.TableNameTidb))
		cmd := fmt.Sprintf("%s -u %s -P %d -h %s -p \"%s\" --filter \"%s.%s\" -o %s %s> %s 2>&1", cfg.T2OInit.DumplingBinPath, cfg.TiDBConfig.User,
			cfg.TiDBConfig.Port, cfg.TiDBConfig.Host, cfg.TiDBConfig.Password,
			task.TableSchemaTidb, task.TableNameTidb, cfg.T2OInit.DumpDataDir,
			cfg.T2OInit.DumpExtraArgs, stdLogPath)
		c := exec.Command("bash", "-c", cmd)
		output, err := c.CombinedOutput()
		dumpEndTime := time.Now()
		dumpDuration := int(dumpEndTime.Sub(dumpStartTime).Seconds())
		if err != nil {
			log.Error(fmt.Sprintf("Run command:%s failed. Check log:%s", cmd, logPath))
			log.Error(fmt.Sprintf("Run command stderr:%s", output))

			task.DumpStatus = StatusFailed
			task.DumpDuration = dumpDuration
			task.LastDumpTime = dumpStartTime
		} else {
			log.Info(fmt.Sprintf("Run command:%s success.", cmd))
			task.DumpStatus = StatusSuccess
			task.DumpDuration = dumpDuration
			task.LastDumpTime = dumpStartTime
		}
		res := database.DB.Save(&task)
		if res.Error != nil {
			log.Error(res.Error)
		}
		log.Info(fmt.Sprintf("[Thread-%d]Finished dump %s.%s data", threadID, task.TableSchemaTidb, task.TableNameTidb))
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
	tasks := make(chan models.T2OConfigModel, threadCount)
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
	res := database.DB.Model(&models.T2OConfigModel{}).Where("generate_ctl_status = ?", StatusWaiting).Count(&tableCount)
	if res.Error != nil {
		log.Error(res.Error)
	}
	var records []models.T2OConfigModel
	res = database.DB.Model(&models.T2OConfigModel{}).Where("generate_ctl_status = ?", StatusWaiting).Scan(&records)
	if res.Error != nil {
		log.Error(res.Error)
	}
	for _, record := range records {
		tasks <- record
	}
	close(tasks)

	wg.Wait()
	return nil
}

func runT2OGenerator(cfg config.OTOConfig, threadID int, tasks <-chan models.T2OConfigModel) {
	for task := range tasks {
		handleCount = handleCount + 1
		generateStartTime := time.Now()
		log.Info(fmt.Sprintf("[Thread-%d]Start to generate oracle sqlldr ctl file for %s.%s", threadID, task.TableSchemaOracle, task.TableNameTidb))
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
			strings.ToUpper(task.TableSchemaOracle), strings.ToUpper(task.TableNameTidb))
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
			FilePath:          filepath.Join(cfg.T2OInit.DumpDataDir, fmt.Sprintf("%s.%s.000000000.csv", task.TableSchemaOracle, task.TableNameTidb)),
			BadFilePath:       filepath.Join(cfg.T2OInit.OracleCtlFileDir, fmt.Sprintf("%s.%s.bad", task.TableSchemaOracle, task.TableNameTidb)),
			DiscardFilePath:   filepath.Join(cfg.T2OInit.OracleCtlFileDir, fmt.Sprintf("%s.%s.disc", task.TableSchemaOracle, task.TableNameTidb)),
			TableOracleSchema: task.TableSchemaOracle,
			TableName:         task.TableNameTidb,
			Columns:           strings.Join(colNames, ","),
		}
		ctlFilePath := filepath.Join(cfg.T2OInit.OracleCtlFileDir, fmt.Sprintf("%s.%s.ctl", task.TableSchemaOracle, task.TableNameTidb))
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
		task.GenerateCtlStatus = StatusSuccess
		task.GenerateCtlDuration = generateDuration
		task.LastGenerateCtlTime = generateStartTime
		res := database.DB.Save(&task)
		if res.Error != nil {
			log.Error(res.Error)
		}

		log.Info(fmt.Sprintf("[Thread-%d]Finished generate ctl for %s.%s", threadID, task.TableSchemaOracle, task.TableNameTidb))
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
	tasks := make(chan models.T2OConfigModel, threadCount)
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
	res := database.DB.Model(&models.T2OConfigModel{}).Where("load_status = ?", StatusWaiting).Count(&tableCount)
	if res.Error != nil {
		log.Error(res.Error)
	}
	var records []models.T2OConfigModel
	res = database.DB.Model(&models.T2OConfigModel{}).Where("load_status = ?", StatusWaiting).Scan(&records)
	if res.Error != nil {
		log.Error(res.Error)
	}
	for _, record := range records {
		tasks <- record
	}
	close(tasks)

	wg.Wait()
	return nil
}

func runT2OLoad(cfg config.OTOConfig, threadID int, tasks <-chan models.T2OConfigModel) error {
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
				strings.ToUpper(task.TableSchemaOracle), strings.ToUpper(task.TableNameTidb))
			log.Debug(truncateSql)
			_, err = db.Exec(truncateSql)
			if err != nil {
				log.Error(err)
				continue
			}
			log.Info(fmt.Sprintf("[Thread-%d]Finished truncate table %s.%s", threadID, task.TableSchemaOracle, task.TableNameTidb))
		}

		log.Info(fmt.Sprintf("[Thread-%d]Start to sqlldr load data %s.%s", threadID, task.TableSchemaOracle, task.TableNameTidb))
		log.Info(fmt.Sprintf("Process load-data %d/%d", handleCount, tableCount))
		ctlFilePath := filepath.Join(cfg.T2OInit.OracleCtlFileDir, fmt.Sprintf("%s.%s.ctl", task.TableSchemaOracle, task.TableNameTidb))
		sqlldrLogPath := filepath.Join(cfg.T2OInit.OracleCtlFileDir, fmt.Sprintf("%s.%s.log", task.TableSchemaOracle, task.TableNameTidb))
		stdLogPath := filepath.Join(cfg.Log.LogDir, fmt.Sprintf("sqlldr_load_%s.%s.log", task.TableSchemaOracle, task.TableNameTidb))
		cmd := fmt.Sprintf("%s %s/%s control=%s log=%s %s> %s 2>&1", cfg.T2OInit.SqlldrBinPath, cfg.OracleConfig.User, cfg.OracleConfig.Password, ctlFilePath, sqlldrLogPath, cfg.T2OInit.SqlldrExtraArgs, stdLogPath)
		c := exec.Command("bash", "-c", cmd)
		output, err := c.CombinedOutput()
		loadEndTime := time.Now()
		loadDuration := int(loadEndTime.Sub(loadStartTime).Seconds())
		if err != nil {
			log.Error(fmt.Sprintf("Run command:%s failed. Check log:%s", cmd, logPath))
			log.Error(fmt.Sprintf("Run command stderr:%s", output))
			task.LoadStatus = StatusFailed
			task.LoadDuration = loadDuration
			task.LastLoadTime = loadStartTime

		} else {
			log.Info(fmt.Sprintf("Run command:%s success.", cmd))
			task.LoadStatus = StatusSuccess
			task.LoadDuration = loadDuration
			task.LastLoadTime = loadStartTime
		}
		res := database.DB.Save(&task)
		if res.Error != nil {
			log.Error(res.Error)
		}
		log.Info(fmt.Sprintf("[Thread-%d]Finished load %s.%s data", threadID, task.TableSchemaOracle, task.TableNameTidb))
	}

	return nil
}
