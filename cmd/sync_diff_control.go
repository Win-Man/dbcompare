/*
 * Created: 2022-09-10 11:49:33
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

// [] sync-diff-o2t
// [] sync-diff-inspector

type SyncDiffTemplate struct {
	LogLevel          string
	ChunkSize         int
	CheckThreadCount  int
	SyncTableName     string
	TableSchema       string
	TableOracleSchema string
	TableName         string
	OracleDB          config.OracleDBConfig
	TiDBDB            config.DBConfig
	SnapSource        string
	SnapTarget        string
	IgnoreCols        string
	LogDir            string
	FilterClauseTidb  string
	FilterClauseOra   string
	IndexFields       string
	OracleHint        string
	TidbHint          string
}

var batchid string
var tableCount int64
var handleCount int64

func newSyncDiffCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "sync-diff <prepare|run>",
		Short: "sync-diff",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Help()
			}
			cfg := config.InitOTOConfig(configPath)
			logger.InitLogger(logLevel, logPath, cfg.Log)
			log.Info("Welcome to sync-diff-control")
			log.Debug(fmt.Sprintf("Flags:%+v", cmd.Flags()))
			log.Debug(fmt.Sprintf("arguments:%s", strings.Join(args, ",")))
			switch args[0] {
			case "prepare":
				err := database.InitDB(cfg.TiDBConfig)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = createConfigTables(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				log.Info("Finished prepare without errors.")
				fmt.Printf("Create table success.\nPls init table data,refer sql:\n")
				fmt.Printf("insert into %s.syncdiff_config_result(table_schema,table_name,sync_status) select table_schema,table_name,'%s' from information_schema.tables where table_schema='mydb' \n",
					cfg.TiDBConfig.Database, SyncWaiting)
			case "run":
				err := database.InitDB(cfg.TiDBConfig)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				err = runSyncDiffControl(cfg)
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

func createConfigTables(cfg config.OTOConfig) error {
	log.Info("Start to create syncdiff_config_result table")

	if !database.DB.Migrator().HasTable(&models.SyncdiffConfigModel{}) {
		err := database.DB.Migrator().CreateTable(&models.SyncdiffConfigModel{})
		if err != nil {
			log.Error(fmt.Sprintf("Create table %s.syncdiff_config_result failed!Error:%v", cfg.TiDBConfig.Database, err))
			return err
		}
	} else {
		err := database.DB.AutoMigrate(&models.SyncdiffConfigModel{})
		if err != nil {
			log.Error(fmt.Sprintf("Migrate table %s.syncdiff_config_result failed!Error:%v", cfg.TiDBConfig.Database, err))
			return err
		}
	}

	return nil
}

func runSyncDiffControl(cfg config.OTOConfig) error {
	//generateSyncDiffConfig("dbdb", "tabletable")
	log.Debug("Start to run SyncdiffControl")
	batchid = time.Now().Format("20060102150405")
	err := os.MkdirAll(cfg.SyncDiffControl.ConfDir, 0755)
	if err != nil {
		log.Error(err)
		return err
	}
	os.MkdirAll(cfg.Log.LogDir, 0755)
	if err != nil {
		log.Error(err)
		return err
	}
	res := database.DB.Model(&models.SyncdiffConfigModel{}).Where("sync_status in (?,?)", SyncWaiting, SyncRunning).Count(&tableCount)
	if res.Error != nil {
		log.Errorf("Execute SQL get error:%v", res.Error)
	}
	fmt.Printf("Fetch %d rows from syncdiff_config_result where sync_status in (%s,%s)\n", tableCount, SyncWaiting, SyncRunning)
	log.Info(fmt.Sprintf("Fetch %d rows from syncdiff_config_result where sync_status in (%s,%s)", tableCount, SyncWaiting, SyncRunning))
	if tableCount == 0 {
		return nil
	}
	ProcessBar = pkg.New(tableCount, pkg.WithFiller("="))
	threadCount := cfg.Performance.Concurrency
	tasks := make(chan models.SyncdiffConfigModel, threadCount)
	var wg sync.WaitGroup
	handleCount = 0
	for i := 1; i <= threadCount; i++ {
		wg.Add(1)
		tmpi := i
		go func() {
			defer wg.Done()
			//log.Debug(fmt.Sprintf("go func i:%d", tmpi))
			runSyncDiff(cfg, tmpi, tasks)
			//testFunc(tmpi, tasks)
		}()
	}

	var records []models.SyncdiffConfigModel
	res = database.DB.Model(&models.SyncdiffConfigModel{}).Where("sync_status in (?,?)", SyncWaiting, SyncRunning).Order("priority desc").Scan(&records)
	if res.Error != nil {
		log.Errorf("Execute SQL get error:%v", res.Error)
	}
	for _, record := range records {
		tasks <- record
	}

	close(tasks)

	wg.Wait()
	ProcessBar.Finish()
	return nil
}

func runSyncDiff(cfg config.OTOConfig, threadID int, tasks <-chan models.SyncdiffConfigModel) {

	for task := range tasks {
		handleCount = handleCount + 1
		ProcessBar.Done(1)
		log.Info(fmt.Sprintf("[Thread-%d]Start to run sync diff for %s.%s", threadID, task.TableSchema, task.TableNameTidb))
		log.Info(fmt.Sprintf("Process sync-diff %d/%d", handleCount, tableCount))

		var syncTableName = fmt.Sprintf("%s.%s", task.TableSchema, task.TableNameTidb)
		var ignoreColumns = task.IgnoreColumns

		var syncTableNameTarget string
		if task.TableSchemaOracle == "" {
			syncTableNameTarget = syncTableName
		} else {
			syncTableNameTarget = fmt.Sprintf("%s.%s", task.TableSchemaOracle, task.TableNameTidb)
		}
		var chunk_size = task.ChunkSize
		var check_thread_count = task.CheckThreadCount
		var use_snapshot string
		if task.UseSnapshot == "" {
			use_snapshot = "NO"
		}
		var snapshot_source = task.SnapshotSource
		var snapshot_target = task.SnapshotTarget
		var tableSchema string
		var tableName string
		var tableSchemaTarget string
		tableSchema = strings.Split(syncTableName, ".")[0]
		tableName = strings.Split(syncTableName, ".")[1]
		tableSchemaTarget = strings.Split(syncTableNameTarget, ".")[0]
		if strings.ToLower(use_snapshot) == "y" || strings.ToLower(use_snapshot) == "yes" {

			if snapshot_source != "" {
				snapshot_source = fmt.Sprintf("snapshot = \"%s\"", snapshot_source)
			}
			if snapshot_target != "" {
				snapshot_target = fmt.Sprintf("oracle-scn = \"%s\"", snapshot_target)
			}
		} else {
			snapshot_source = ""
			snapshot_target = ""
		}

		// update sync_status=running
		task.Batchid = batchid
		task.JobStarttime = time.Now()
		task.SyncStatus = SyncRunning
		task.Remark = fmt.Sprintf("%d", threadID)
		res := database.DB.Save(&task)
		if res.Error != nil {
			log.Error(res.Error)
		}
		log.Info(fmt.Sprintf("[Thread-%d]Finish update config to running id:%d %s", threadID, task.Id, syncTableName))

		var db *sql.DB
		var err error
		db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
		defer db.Close()
		if err != nil {
			log.Error(fmt.Sprintf("Connect source database error:%v", err))
			continue
		}
		stmtQuery := fmt.Sprintf("select count(1) from %s ", syncTableName)
		rows, err := db.Query(stmtQuery)
		if err != nil {
			log.Error(err)
			continue
		}
		var rowCount int
		rows.Next()
		rows.Scan(&rowCount)
		log.Info(fmt.Sprintf("[Thread-%d]id:%d table %s count in tidb:%d", threadID, task.Id, syncTableName, rowCount))
		syncStartTime := time.Now()
		//Generate sync condig
		err = generateSyncDiffConfig(tableSchema, tableName, tableSchemaTarget, ignoreColumns,
			cfg.SyncDiffControl.ConfDir, chunk_size, check_thread_count, snapshot_source,
			snapshot_target, task, cfg)
		if err != nil {
			log.Error(fmt.Sprintf("[Thread-%d]GenerateSyncDiffConfig error:%v", threadID, err))
			continue
		}
		//Do sync-diff

		rtCode := runSyncDiffTask(cfg, tableSchema, tableName)

		// update sync-diff status
		syncEndTime := time.Now()
		durationTime := int(syncEndTime.Sub(syncStartTime).Seconds())

		task.TableCount = rowCount
		task.SyncStatus = rtCode
		task.SyncStarttime = syncStartTime
		task.SyncEndtime = syncEndTime
		task.SyncDuration = durationTime
		res = database.DB.Save(&task)
		if res.Error != nil {
			log.Error(res.Error)
			continue
		}

		if rtCode == CompareFailed {
			continue
		}

		// update from summary
		stmtUpdt2 := fmt.Sprintf(`update %s.syncdiff_config_result t1
			inner join sync_diff_inspector.summary t2 
				on t1.table_schema = t2.schema
				and t1.table_name = t2.table 
			set t1.chunk_num = t2.chunk_num,
				t1.check_success_num = t2.check_success_num,
				t1.check_failed_num = t2.check_failed_num,
				t1.check_ignore_num = t2.check_ignore_num,
				t1.state = t2.state,
				t1.config_hash = t2.config_hash,
				t1.update_time = t2.update_time
			where id = %d`, cfg.TiDBConfig.Database, task.Id)
		_, err = db.Exec(stmtUpdt2)
		if err != nil {
			log.Error(fmt.Sprintf("[Thread-%d]Update sync summary log failed. error:%v", threadID, err))
			continue
		} else {
			log.Info(fmt.Sprintf("[Thread-%d]Update sync summary success.", threadID))
		}

		log.Info(fmt.Sprintf("[Thread-%d]Finished run sync diff for id:%d %s.%s. [table count %d/%d]", threadID, task.Id, tableSchema, tableName, handleCount, tableCount))

	}
}

func generateSyncDiffConfig(tableSchema string, tableName string, tableSchemaTarget string, ignoreCols string,
	confDir string, chunkSize int, checkThreadCount int, snapSource string, snapTarget string,
	task models.SyncdiffConfigModel, cfg config.OTOConfig) error {
	log.Info(fmt.Sprintf("Start to generate o2t-sync-diff config for %s.%s", tableSchema, tableName))
	tpl, err := template.ParseFiles(cfg.SyncDiffControl.SyncTemplate)
	if err != nil {
		log.Error(fmt.Sprintf("template parsefiles failed,err:%v", err))
		return err
	}
	syncdifftmp := SyncDiffTemplate{
		LogLevel:          task.LogLevel,
		ChunkSize:         chunkSize,
		CheckThreadCount:  checkThreadCount,
		SyncTableName:     fmt.Sprintf("%s.%s", tableSchema, tableName),
		TableSchema:       tableSchema,
		TableName:         tableName,
		TableOracleSchema: tableSchemaTarget,
		OracleDB:          cfg.OracleConfig,
		TiDBDB:            cfg.TiDBConfig,
		SnapSource:        snapSource,
		SnapTarget:        snapTarget,
		LogDir:            cfg.Log.LogDir,
		FilterClauseTidb:  task.FilterClauseTidb,
		FilterClauseOra:   task.FilterClauseOra,
		IndexFields:       task.IndexFields,
		OracleHint:        task.OracleHint,
		TidbHint:          task.TidbHint,
	}
	if ignoreCols != "" {
		syncdifftmp.IgnoreCols = strings.Replace(ignoreCols, ",", "\",\"", -1)
	}
	f, err := os.Create(filepath.Join(cfg.SyncDiffControl.ConfDir, fmt.Sprintf("sync_diff_%s.%s.toml", tableSchema, tableName)))
	defer f.Close()
	if err != nil {
		log.Error(err)
		return err
	}

	err = tpl.Execute(f, syncdifftmp)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func runSyncDiffTask(cfg config.OTOConfig, tableSchema string, tableName string) string {
	var rtCode string
	log.Info(fmt.Sprintf("Start to run o2t-sync-diff for %s.%s", tableSchema, tableName))
	confPath := filepath.Join(cfg.SyncDiffControl.ConfDir, fmt.Sprintf("sync_diff_%s.%s.toml", tableSchema, tableName))
	stdLogPath := filepath.Join(cfg.Log.LogDir, fmt.Sprintf("sync_diff_%s.%s.log", tableSchema, tableName))
	cmd := fmt.Sprintf("%s -config %s > %s 2>&1", cfg.SyncDiffControl.BinPath, confPath, stdLogPath)
	c := exec.Command("bash", "-c", cmd)
	// cmdTest := fmt.Sprintf("%s %s", binPath, confPath)
	// c := exec.Command("bash", "-c", cmdTest)
	output, err := c.CombinedOutput()
	if err != nil {
		log.Error(fmt.Sprintf("Run command:%s failed. Check log:%s", cmd, stdLogPath))
		log.Error(fmt.Sprintf("Run command stderr:%s", output))
		rtCode = CompareFailed
	} else {
		log.Info(fmt.Sprintf("Run command:%s success.", cmd))
		rtCode = CompareSuccess
	}
	return rtCode
}
