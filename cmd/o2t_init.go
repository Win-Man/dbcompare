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
				fmt.Printf("INSERT INTO %s.o2t_config() VALUES ", cfg.TiDBConfig.Database)
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
	database.DB.Model(&models.O2TConfigModel{}).Where("dump_status = ?", StatusWaiting).Count(&tableCount)
	var records []models.O2TConfigModel
	database.DB.Model(&models.O2TConfigModel{}).Where("dump_status = ?", StatusWaiting).Scan(&records)
	for _, record := range records {
		tasks <- record
	}
	close(tasks)

	wg.Wait()
	return nil
}

func runO2TDumpData(cfg config.OTOConfig, threadID int, tasks <-chan models.O2TConfigModel) {
	for task := range tasks {
		handleCount = handleCount + 1
		dumpStartTime := time.Now()
		log.Info(fmt.Sprintf("[Thread-%d]Start to dump %s.%s data", threadID, task.TableSchemaOracle, task.TableNameTidb))
		log.Info(fmt.Sprintf("Process dump-data %d/%d", handleCount, tableCount))
		stdLogPath := filepath.Join(cfg.Log.LogDir, fmt.Sprintf("sqlldr2_%s.%s.log", task.TableSchemaOracle, task.TableNameTidb))
		dataPath := filepath.Join(cfg.O2TInit.DumpDataDir, fmt.Sprintf("%s.%s.%%B.csv",
			task.TableSchemaOracle, task.TableNameTidb))
		cmd := fmt.Sprintf("%s user=%s/%s@%s query=%s.%s file=%s %s > %s 2>&1",
			cfg.O2TInit.Sqlldr2BinPath, cfg.OracleConfig.User, cfg.OracleConfig.Password,
			cfg.OracleConfig.ServiceName, strings.ToUpper(task.TableSchemaOracle), strings.ToUpper(task.TableNameTidb),
			dataPath, cfg.O2TInit.Sqlldr2ExtraArgs, stdLogPath)
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
		database.DB.Save(&task)
		log.Info(fmt.Sprintf("[Thread-%d]Finished dump %s.%s data", threadID, task.TableSchemaOracle, task.TableNameTidb))
	}
}

func runO2TGenerateConfControl(cfg config.OTOConfig) error {
	var err error
	err = os.MkdirAll(cfg.O2TInit.LightningTomlDir, 0755)
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
	database.DB.Model(models.O2TConfigModel{}).
		Where("dump_status = ?", StatusSuccess).
		Updates(models.O2TConfigModel{GenerateConfStatus: StatusSuccess,
			GenerateDuration:     generateDuration,
			LastGenerateConfTime: generateStartTime})

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
	if err != nil {
		log.Error(fmt.Sprintf("Run command:%s failed. Check log:%s", cmd, logPath))
		log.Error(fmt.Sprintf("Run command stderr:%s", output))
		database.DB.Model(models.O2TConfigModel{}).
			Where("dump_status = ?", StatusSuccess).
			Updates(models.O2TConfigModel{LoadStatus: StatusFailed,
				LoadDuration: loadDuration,
				LastLoadTime: loadStartTime})

	} else {
		log.Info(fmt.Sprintf("Run command:%s success.", cmd))
		database.DB.Model(models.O2TConfigModel{}).
			Where("dump_status = ?", StatusSuccess).
			Updates(models.O2TConfigModel{LoadStatus: StatusSuccess,
				LoadDuration: loadDuration,
				LastLoadTime: loadStartTime})
	}
	log.Info("Finished load data by tidb-lightning")
	return nil
}
