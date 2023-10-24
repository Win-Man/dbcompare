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
	"github.com/Win-Man/dbcompare/pkg"
	"github.com/Win-Man/dbcompare/pkg/logger"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ProcessBar *pkg.Bar

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
				err := database.InitDB(cfg.TiDBConfig)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = createO2TInitMeta(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				log.Info("Finished prepare without errors.")
				fmt.Printf("Create table success.\nPls init table data,refer sql:\n")
				fmt.Printf(`INSERT INTO %s.o2t_config(table_schema_tidb,table_name_tidb,table_schema_oracle,dump_status,generate_conf_status,load_status) VALUES ('mydb','mytab','ordb','%s','%s','%s')`,
					cfg.TiDBConfig.Database, StatusWaiting, StatusInitialize, StatusInitialize)
			case "dump-data":
				err := database.InitDB(cfg.TiDBConfig)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = runO2TDumpDataControl(cfg)
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
				err = runO2TGenerateConfControl(cfg)
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
				err = runO2TLoadDataControl(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				if cfg.Performance.CheckRowCount {
					err = runGetRowsControl(cfg)
					if err != nil {
						log.Error(err)
						os.Exit(1)
					}
				}
			case "all":
				err := database.InitDB(cfg.TiDBConfig)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = runO2TDumpDataControl(cfg)
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
				if cfg.Performance.CheckRowCount {
					err = runGetRowsControl(cfg)
					if err != nil {
						log.Error(err)
						os.Exit(1)
					}
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

type LightingTomlTemplate struct {
	BatchID     string
	DumpDataDir string
	TiDBDB      config.DBConfig
}

func createO2TInitMeta(cfg config.OTOConfig) error {
	log.Info("Start to create o2t_config table")
	fmt.Println("Start to create o2t_config table")
	if !database.DB.Migrator().HasTable(&models.O2TConfigModel{}) {
		err := database.DB.Migrator().CreateTable(&models.O2TConfigModel{})
		if err != nil {
			log.Error(fmt.Sprintf("Create table %s.o2t_config failed!Error:%v", cfg.TiDBConfig.Database, err))
			return err
		}
	} else {
		err := database.DB.AutoMigrate(&models.O2TConfigModel{})
		if err != nil {
			log.Error(fmt.Sprintf("Migrate table %s.o2t_config failed!Error:%v", cfg.TiDBConfig.Database, err))
			return err
		}
	}
	return nil
}

func runO2TDumpDataControl(cfg config.OTOConfig) error {
	var err error
	err = os.MkdirAll(cfg.O2TInit.DumpDataDir, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(cfg.Log.LogDir, 0755)
	if err != nil {
		return err
	}

	res := database.DB.Model(&models.O2TConfigModel{}).Where("dump_status in (?,?)", StatusWaiting, StatusRunning).Count(&tableCount)
	if res.Error != nil {
		log.Error(res.Error)
	}
	fmt.Printf("Fetch %d rows from o2t_config where dump_status in (%s,%s)\n", tableCount, StatusWaiting, StatusRunning)
	log.Info(fmt.Sprintf("Fetch %d rows from o2t_config where dump_status in (%s,%s)", tableCount, StatusWaiting, StatusRunning))
	if tableCount == 0 {
		return nil
	}
	ProcessBar = pkg.New(tableCount, pkg.WithFiller("="))
	threadCount := cfg.Performance.Concurrency
	tasks := make(chan models.O2TConfigModel, threadCount)
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
	var records []models.O2TConfigModel
	res = database.DB.Model(&models.O2TConfigModel{}).Where("dump_status in (?,?)", StatusWaiting, StatusRunning).Scan(&records)
	if res.Error != nil {
		log.Error(res.Error)
	}
	for _, record := range records {
		tasks <- record
	}
	close(tasks)

	wg.Wait()
	ProcessBar.Finish()
	return nil
}

func runO2TDumpData(cfg config.OTOConfig, threadID int, tasks <-chan models.O2TConfigModel) {

	for task := range tasks {
		handleCount = handleCount + 1
		ProcessBar.Done(1)

		log.Info(fmt.Sprintf("[Thread-%d]Start to dump %s.%s data", threadID, task.TableSchemaOracle, task.TableNameTidb))
		log.Info(fmt.Sprintf("Process dump-data %d/%d", handleCount, tableCount))

		// get oracle table row count
		if cfg.Performance.CheckRowCount {
			i, err := getOracleRowsCount(cfg.OracleConfig, task.TableSchemaOracle, task.TableNameTidb)
			if err != nil {
				log.Error(err)
				continue
			}
			task.OracleRowsCount = i
		}

		// update dump_status=running
		dumpStartTime := time.Now()
		task.DumpStatus = StatusRunning
		task.LastDumpTime = dumpStartTime

		res := database.DB.Save(&task)
		if res.Error != nil {
			log.Error(res.Error)
		}

		stdLogPath := filepath.Join(cfg.Log.LogDir, fmt.Sprintf("sqluldr2_%s.%s.log", task.TableSchemaOracle, task.TableNameTidb))
		dataPath := filepath.Join(cfg.O2TInit.DumpDataDir, fmt.Sprintf("%s.%s.%%B.csv",
			task.TableSchemaTidb, task.TableNameTidb))
		ctlPath := filepath.Join(cfg.Log.LogDir, fmt.Sprintf("%s.%s_sqlldr.ctl", strings.ToUpper(task.TableSchemaOracle), strings.ToUpper(task.TableNameTidb)))
		// cmd := fmt.Sprintf("%s user=%s/%s@%s query=%s.%s file=%s control=%s %s > %s 2>&1",
		// 	cfg.O2TInit.Sqluldr2BinPath, cfg.OracleConfig.User, cfg.OracleConfig.Password,
		// 	cfg.OracleConfig.ServiceName, strings.ToUpper(task.TableSchemaOracle), strings.ToUpper(task.TableNameTidb),
		// 	dataPath, ctlPath, cfg.O2TInit.Sqluldr2ExtraArgs, stdLogPath)
		extraColumnsStr := ""
		if task.DumpExtraCols != "" {
			extraColumnsStr = fmt.Sprintf(",%s", task.DumpExtraCols)
		}
		dumpFliter := ""
		if task.DumpFilterClauseOra != "" {
			dumpFliter = fmt.Sprintf("where %s", task.DumpFilterClauseOra)
		}
		cmd := fmt.Sprintf("%s user=%s/%s@%s query=\"SELECT *%s FROM %s.%s %s\" file=%s control=%s %s > %s 2>&1",
			cfg.O2TInit.Sqluldr2BinPath, cfg.OracleConfig.User, cfg.OracleConfig.Password,
			cfg.OracleConfig.ServiceName, extraColumnsStr, strings.ToUpper(task.TableSchemaOracle), strings.ToUpper(task.TableNameTidb),
			dumpFliter, dataPath, ctlPath, cfg.O2TInit.Sqluldr2ExtraArgs, stdLogPath)
		c := exec.Command("bash", "-c", cmd)
		output, err := c.CombinedOutput()
		dumpEndTime := time.Now()
		dumpDuration := int(dumpEndTime.Sub(dumpStartTime).Seconds())
		if err != nil {
			log.Error(fmt.Sprintf("Run command:%s failed. Check log:%s", cmd, stdLogPath))
			log.Error(fmt.Sprintf("Run command stderr:%s", output))
			task.DumpStatus = StatusFailed
			task.DumpDuration = dumpDuration

		} else {
			log.Info(fmt.Sprintf("Run command:%s success.", cmd))
			task.DumpStatus = StatusSuccess
			task.GenerateConfStatus = StatusWaiting
			task.LoadStatus = StatusWaiting
			task.DumpDuration = dumpDuration
		}
		res = database.DB.Save(&task)
		if res.Error != nil {
			log.Error(res.Error)
		}
		log.Info(fmt.Sprintf("[Thread-%d]Finished dump %s.%s data. [table count %d/%d]", threadID, task.TableSchemaOracle, task.TableNameTidb, handleCount, tableCount))
	}
}

func runO2TGenerateConfControl(cfg config.OTOConfig) error {
	err := os.MkdirAll(cfg.O2TInit.LightningTomlDir, 0755)
	if err != nil {
		log.Error(err)
		return err
	}
	return runO2TGenerateConf(cfg)

}

func runO2TGenerateConf(cfg config.OTOConfig) error {
	log.Info("Start to generate tidb-lightning.toml ")
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

	lightningTomlPath := filepath.Join(cfg.O2TInit.LightningTomlDir, "tidb-lightning.toml")
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
	res := database.DB.Model(models.O2TConfigModel{}).
		Where("generate_conf_status = ?", StatusWaiting).
		Updates(models.O2TConfigModel{GenerateConfStatus: StatusSuccess,
			GenerateDuration:     generateDuration,
			LastGenerateConfTime: generateStartTime,
			LoadStatus:           StatusWaiting})
	if res.Error != nil {
		log.Error(res.Error)
	}

	log.Info("Finished generate tidb-lighting.toml")
	return nil
}

func runO2TLoadDataControl(cfg config.OTOConfig) error {

	return runO2TLoadData(cfg)
}

func runO2TLoadData(cfg config.OTOConfig) error {

	log.Info("Start to load data by tidb-lightning ")
	loadStartTime := time.Now()

	lightningTomlPath := filepath.Join(cfg.O2TInit.LightningTomlDir, "tidb-lightning.toml")
	lightningStdPath := filepath.Join(cfg.O2TInit.LightningTomlDir, "tidb_lighting_stderr.log")
	cmd := fmt.Sprintf("%s -config %s %s > %s 2>&1", cfg.O2TInit.LightningBinPath, lightningTomlPath,
		cfg.O2TInit.LightningExtraArgs, lightningStdPath)
	c := exec.Command("bash", "-c", cmd)
	output, err := c.CombinedOutput()
	loadEndTime := time.Now()
	loadDuration := int(loadEndTime.Sub(loadStartTime).Seconds())
	if err != nil {
		log.Error(fmt.Sprintf("Run command:%s failed. Check log:%s", cmd, lightningStdPath))
		log.Error(fmt.Sprintf("Run command stderr:%s", output))
		res := database.DB.Model(models.O2TConfigModel{}).
			Where("load_status = ?", StatusWaiting).
			Updates(models.O2TConfigModel{LoadStatus: StatusFailed,
				LoadDuration: loadDuration,
				LastLoadTime: loadStartTime})
		if res.Error != nil {
			log.Error(res.Error)
		}

	} else {
		log.Info(fmt.Sprintf("Run command:%s success.", cmd))
		if cfg.Performance.CheckRowCount {
			res := database.DB.Model(models.O2TConfigModel{}).
				Where("load_status = ?", StatusWaiting).
				Updates(models.O2TConfigModel{LoadStatus: StatusWaiting,
					LoadDuration: loadDuration,
					LastLoadTime: loadStartTime})
			if res.Error != nil {
				log.Error(res.Error)
			}
		} else {
			res := database.DB.Model(models.O2TConfigModel{}).
				Where("load_status = ?", StatusWaiting).
				Updates(models.O2TConfigModel{LoadStatus: StatusSuccess,
					LoadDuration: loadDuration,
					LastLoadTime: loadStartTime})
			if res.Error != nil {
				log.Error(res.Error)
			}
		}

	}
	log.Info("Finished load data by tidb-lightning")
	return nil
}

func runGetRowsControl(cfg config.OTOConfig) error {

	threadCount := cfg.Performance.Concurrency
	tasks := make(chan models.O2TConfigModel, threadCount)
	var wg sync.WaitGroup
	for i := 1; i <= threadCount; i++ {
		wg.Add(1)
		go func() {
			runGetRows(cfg, tasks)
		}()
	}
	res := database.DB.Model(&models.O2TConfigModel{}).Where("load_status = ?", StatusWaiting).Count(&tableCount)
	if res.Error != nil {
		log.Error(res.Error)
	}
	if tableCount == 0 {
		return nil
	}
	fmt.Printf("Get rows count process\n")
	ProcessBar = pkg.New(tableCount, pkg.WithFiller("="))

	var records []models.O2TConfigModel
	res = database.DB.Model(&models.O2TConfigModel{}).Where("load_status = ?", StatusWaiting).Scan(&records)
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

func runGetRows(cfg config.OTOConfig, tasks <-chan models.O2TConfigModel) {
	for task := range tasks {
		// get tidb table row count
		if cfg.Performance.CheckRowCount {
			i, err := getTidbRowsCount(cfg.TiDBConfig, task.TableSchemaTidb, task.TableNameTidb)
			if err != nil {
				log.Error(err)
				continue
			}
			task.TidbRowsCount = i
			task.LoadStatus = StatusSuccess
		}
		res := database.DB.Save(&task)
		if res.Error != nil {
			log.Error(res.Error)
		}
		ProcessBar.Done(1)
	}
}

func getOracleRowsCount(dbcfg config.OracleDBConfig, tableSchemaOracle string, tableName string) (int, error) {
	log.Debug(fmt.Sprintf("Get Oracle %s.%s rows count", tableSchemaOracle, tableName))
	var db *sql.DB
	var err error
	db, err = database.OpenOracleDB(&dbcfg)
	defer db.Close()
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		return -1, err
	}
	querySql := fmt.Sprintf("select count(1) from %s.%s ", tableSchemaOracle, tableName)
	log.Debug(fmt.Sprintf("Sql: %s", querySql))
	rows, err := db.Query(querySql)
	if err != nil {
		log.Error(err)
		return -1, err
	}
	var oracleRowsCount int
	for rows.Next() {
		rows.Scan(&oracleRowsCount)
	}
	return oracleRowsCount, nil
}

func getTidbRowsCount(dbcfg config.DBConfig, tableSchema string, tableName string) (int, error) {
	log.Debug(fmt.Sprintf("Get TiDB %s.%s rows count", tableSchema, tableName))
	var db *sql.DB
	var err error
	db, err = database.OpenMySQLDB(&dbcfg)
	defer db.Close()
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		return -1, err
	}
	querySql := fmt.Sprintf("select count(1) from %s.%s ", tableSchema, tableName)
	log.Debug(fmt.Sprintf("Sql: %s", querySql))
	rows, err := db.Query(querySql)
	if err != nil {
		log.Error(err)
		return -1, err
	}
	var tidbRowsCount int
	for rows.Next() {
		rows.Scan(&tidbRowsCount)
	}
	return tidbRowsCount, nil
}
