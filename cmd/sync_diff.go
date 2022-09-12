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
	"github.com/Win-Man/dbcompare/pkg/logger"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// [] sync-diff-o2t
// [] sync-diff-inspector

type SyncDiffTemplate struct {
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
}

func newSyncDiffCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "sync-diff <prepare|run>",
		Short: "sync-diff",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Help()
			}
			cfg := config.InitSyncDiffConfig(configPath)
			logger.InitLogger(logLevel, logPath, cfg.Log)
			log.Info("Welcome to sync-diff")
			log.Debug(fmt.Sprintf("Flags:%+v", cmd.Flags()))
			log.Debug(fmt.Sprintf("arguments:%s", strings.Join(args, ",")))
			switch args[0] {
			case "prepare":
				err := createConfigTables(cfg)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				log.Info("Finished without errors.")
				fmt.Printf("Create table success.\nPls init table data,refer sql:\n")
				fmt.Printf(fmt.Sprintf("insert into %s.syncdiff_config_result(table_schema,table_name,sync_status) select table_schema,table_name,'waiting' from information_schema.tables where table_schema='mydb' \n", cfg.TiDBConfig.Database))
			case "run":
				runSyncDiffControl(cfg)
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

func createConfigTables(cfg config.SyncDiffConfig) error {
	log.Info("Start to create syncdiff_config_result and syncdiff_config_result_his table")
	table_sql := `
	CREATE TABLE IF NOT EXISTS syncdiff_config_result (
		id int(11) NOT NULL AUTO_INCREMENT,
		updatetime datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		table_schema_oracle varchar(64) ,
		table_schema varchar(64) NOT NULL,
		table_name varchar(64) NOT NULL,
		batchid varchar(128) ,
		table_count int(11) ,
		sync_status varchar(32) ,
		sync_starttime datetime ,
		sync_endtime datetime ,
		sync_duration int(11) ,
		sync_messages varchar(1000) ,
		job_starttime datetime ,
		chunk_num int(11),
		check_success_num int(11),
		check_failed_num int(11),
		check_ignore_num int(11),
		state varchar(32),
		config_hash varchar(50),
		update_time datetime,
		ignore_columns varchar(128) ,
		filter_clause_tidb varchar(128) ,
		filter_clause_ora varchar(128) ,
		chunk_size int DEFAULT 100000,
		check_thread_count int DEFAULT 10,
		use_snapshot varchar(10) ,
		snapshot_source varchar(100) ,
		snapshot_target varchar(100) ,
		use_tso varchar(10) ,
		tso_info varchar(100) ,
		source_info varchar(256) ,
		target_info varchar(256) ,
		contain_datatypes varchar(256) ,
		table_label varchar(128) ,
		remark varchar(128) ,
		PRIMARY KEY (id) ,
		UNIQUE KEY uk_tab (table_schema,table_name)
	  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin 
	`
	table_his_sql := `
	CREATE TABLE  IF NOT EXISTS syncdiff_config_result_his (
		id int(11) NOT NULL ,
		updatetime datetime DEFAULT CURRENT_TIMESTAMP,
		table_schema_oracle varchar(64) ,
		table_schema varchar(64) NOT NULL,
		table_name varchar(64) NOT NULL,
		batchid varchar(128) ,
		table_count int(11) ,
		sync_status varchar(32) ,
		sync_starttime datetime ,
		sync_endtime datetime ,
		sync_duration int(11) ,
		sync_messages varchar(1000) ,
		job_starttime datetime ,
		chunk_num int(11),
		check_success_num int(11),
		check_failed_num int(11),
		check_ignore_num int(11),
		state varchar(32),
		config_hash varchar(50),
		update_time datetime,
		ignore_columns varchar(128) ,
		filter_clause_tidb varchar(128) ,
		filter_clause_ora varchar(128) ,
		chunk_size int DEFAULT 100000,
		check_thread_count int DEFAULT 10,
		use_snapshot varchar(10) ,
		snapshot_source varchar(100) ,
		snapshot_target varchar(100) ,
		use_tso varchar(10) ,
		tso_info varchar(100) ,
		source_info varchar(256) ,
		target_info varchar(256) ,
		contain_datatypes varchar(256) ,
		table_label varchar(128) ,
		remark varchar(128) ,
		archivetime datetime DEFAULT CURRENT_TIMESTAMP,
		KEY uk_tab (table_name,table_schema)
	  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin 
	`
	var db *sql.DB
	var err error
	db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
	defer db.Close()
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		return err
	}
	_, err = db.Exec(table_sql)
	if err != nil {
		log.Error("Create table syncdiff_config_result failed!")
		log.Error(fmt.Sprintf("Execute statement error:%v", err))
		return err
	}
	_, err = db.Exec(table_his_sql)
	if err != nil {
		log.Error("Create table syncdiff_config_result_his failed!")
		log.Error(fmt.Sprintf("Execute statement error:%v", err))
		return err
	}

	return nil
}

func runSyncDiffControl(cfg config.SyncDiffConfig) error {
	//generateSyncDiffConfig("dbdb", "tabletable")
	var err error
	err = os.MkdirAll(cfg.SyncCtlConfig.ConfDir, 0755)
	if err != nil {
		log.Error(err)
		return err
	}
	os.MkdirAll(cfg.Log.LogDir, 0755)
	if err != nil {
		log.Error(err)
		return err
	}
	threadCount := cfg.SyncCtlConfig.Concurrency
	tasks := make(chan int, threadCount)
	var wg sync.WaitGroup
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
	var db *sql.DB

	db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		return err
	}
	querysql := fmt.Sprintf(`SELECT id FROM %s.syncdiff_config_result where sync_status='waiting'`, cfg.TiDBConfig.Database)
	var rowid int
	rows, err := db.Query(querysql)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	for rows.Next() {
		rows.Scan(&rowid)
		tasks <- rowid
	}

	db.Close()
	close(tasks)
	// var db *sql.DB
	// var err error
	// db, err = database.OpenMySQLDB(&cfg.TiDBConfig)

	// db.Close()
	wg.Wait()
	return nil
}

func testFunc(threadID int, tasks <-chan int) {

	for jobid := range tasks {
		log.Info(fmt.Sprintf("[Thread-%d]Start to run sync diff taskid:%d", threadID, jobid))
		time.Sleep(1 * time.Second)
	}
}

func runSyncDiff(cfg config.SyncDiffConfig, threadID int, tasks <-chan int) error {

	for taskid := range tasks {
		log.Info(fmt.Sprintf("[Thread-%d]Start to run sync diff for id:%d", threadID, taskid))
		var db *sql.DB
		var err error
		db, err = database.OpenMySQLDB(&cfg.TiDBConfig)
		defer db.Close()
		if err != nil {
			log.Error(fmt.Sprintf("[Thread-%d]Connect source database error:%v", threadID, err))
			return err
		}
		stmtQuery := fmt.Sprintf(`
            select concat(table_schema,'.',table_name),ifnull(ignore_columns,'') 
                  ,concat(ifnull(table_schema_oracle,table_schema),'.',table_name)
                  ,ifnull(chunk_size,1000),ifnull(check_thread_count,10)
                  ,ifnull(use_snapshot,'NO'),snapshot_source,snapshot_target
            from %s.syncdiff_config_result t 
            where sync_status='waiting' and id = %d
            `, cfg.TiDBConfig.Database, taskid)

		rows, err := db.Query(stmtQuery)
		if err == sql.ErrNoRows {
			log.Info(fmt.Sprintf("[Thread-%d]id:%d sync_status != waiting", threadID, taskid))
			continue
		} else if err != nil {
			log.Error(err)
			continue
		}
		var syncTableName string
		var ignoreColumns string
		var syncTableNameTarget string
		var chunk_size int
		var check_thread_count int
		var use_snapshot string
		var snapshot_source string
		var snapshot_target string
		var tableSchema string
		var tableName string
		var tableSchemaTarget string
		rows.Next()
		rows.Scan(&syncTableName, &ignoreColumns, &syncTableNameTarget, &chunk_size, &check_thread_count, &use_snapshot, &snapshot_source, &snapshot_target)
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
		//TODO set remark and batchid
		stmt_updt0 := fmt.Sprintf(`
			update %s.syncdiff_config_result
			set batchid = '%s',
				job_starttime = now(),
				sync_status = 'running',
				sync_starttime = null,
				sync_endtime = null ,
				remark = '%s',
				chunk_num = null,
				check_success_num = null,
				check_failed_num = null,
				check_ignore_num = null
			where id=%d`, cfg.TiDBConfig.Database, "batchid", "pidremark", taskid)
		db.Exec(stmt_updt0)
		log.Info(fmt.Sprintf("[Thread-%d]Finish update config to running id:%d %s", threadID, taskid, syncTableName))
		stmtQuery = fmt.Sprintf("select count(1) from %s t", syncTableName)
		rows, err = db.Query(stmtQuery)
		if err != nil {
			log.Error(err)
			continue
		}
		var rowCount int
		rows.Next()
		rows.Scan(&rowCount)
		log.Info(fmt.Sprintf("[Thread-%d]id:%d table %s count in tidb:%d", threadID, taskid, syncTableName, rowCount))
		syncStartTime := time.Now()
		//Generate sync condig
		generateSyncDiffConfig(tableSchema, tableName, tableSchemaTarget, ignoreColumns,
			cfg.SyncCtlConfig.ConfDir, chunk_size, check_thread_count, snapshot_source, snapshot_target, cfg)
		//Do sync-diff
		confPath := filepath.Join(cfg.SyncCtlConfig.ConfDir, fmt.Sprintf("sync_diff_%s.%s.toml", tableSchema, tableName))
		logPath := filepath.Join(cfg.Log.LogDir, fmt.Sprintf("sync_diff_%s.%s.log", tableSchema, tableName))
		rtCode := runSyncDiffTask(cfg.SyncCtlConfig.BinPath, confPath, logPath)

		syncEndTime := time.Now()
		durationTime := int(syncEndTime.Sub(syncStartTime).Seconds())
		stmtUpdate1 := fmt.Sprintf(`update %s.syncdiff_config_result set 
			table_count = %d,
			sync_status = '%s',
			sync_starttime = '%s',
			sync_endtime = '%s',
			sync_duration = %d where id = %d
		`, cfg.TiDBConfig.Database, rowCount, rtCode, syncStartTime.Format("2006-01-02 15:05:04"), syncEndTime.Format("2006-01-02 15:05:04"), durationTime, taskid)
		_, err = db.Exec(stmtUpdate1)
		if err != nil {
			log.Error(err)
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
			where id = %d`, cfg.TiDBConfig.Database, taskid)
		_, err = db.Exec(stmtUpdt2)
		if err != nil {
			log.Error(fmt.Sprintf("[Thread-%d]Update sync summary log failed. error:%v", threadID, err))
			continue
		} else {
			log.Info(fmt.Sprintf("[Thread-%d]Update sync summary success.", threadID))
		}

		log.Info(fmt.Sprintf("[Thread-%d]Finished run sync diff for id:%d %s.%s", threadID, taskid, tableSchema, tableName))

	}

	return nil
}

func generateSyncDiffConfig(tableSchema string, tableName string, tableSchemaTarget string, ignoreCols string,
	confDir string, chunkSize int, checkThreadCount int, snapSource string, snapTarget string,
	cfg config.SyncDiffConfig) error {
	log.Info(fmt.Sprintf("Start to generate o2t-sync-diff config for %s.%s", tableSchema, tableName))
	tpl, err := template.ParseFiles(cfg.SyncCtlConfig.SyncTemplate)
	if err != nil {
		log.Error(fmt.Sprintf("template parsefiles failed,err:%v", err))
	}
	syncdifftmp := SyncDiffTemplate{
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
	}
	if ignoreCols != "" {
		syncdifftmp.IgnoreCols = strings.Replace(ignoreCols, ",", "\",\"", -1)
	}
	f, err := os.Create(filepath.Join(cfg.SyncCtlConfig.ConfDir, fmt.Sprintf("sync_diff_%s.%s.toml", tableSchema, tableName)))
	defer f.Close()
	if err != nil {
		log.Error(err)
		return err
	}

	tpl.Execute(f, syncdifftmp)

	return nil
}

func runSyncDiffTask(binPath string, confPath string, logPath string) string {
	var rtCode string
	log.Info(fmt.Sprintf("Start to run o2t-sync-diff program:%s -config %s > %s",
		binPath, confPath, logPath))
	cmd := fmt.Sprintf("%s -config %s > %s", binPath, confPath, logPath)
	c := exec.Command("bash", "-c", cmd)
	// cmdTest := fmt.Sprintf("%s %s", binPath, confPath)
	// c := exec.Command("bash", "-c", cmdTest)
	output, err := c.CombinedOutput()
	if err != nil {
		log.Error(fmt.Sprintf("Run command:%s failed. Check log:%s", cmd, logPath))
		log.Error(fmt.Sprintf("Run command stderr:%s", output))
		rtCode = "compare_fail"
	} else {
		log.Info(fmt.Sprintf("Run command:%s success.", cmd))
		rtCode = "compare_succ"
	}
	return rtCode
}
