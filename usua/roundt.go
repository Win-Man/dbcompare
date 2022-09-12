package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"

	"github.com/Win-Man/dbcompare/compare"
	"github.com/Win-Man/dbcompare/config"
	"github.com/Win-Man/dbcompare/database"
	"github.com/Win-Man/dbcompare/pkg/logger"
)

var (
	configPath  string
	tableName   string
	colName     string
	threadCount int
	rangeMax    int
	h           bool
)

var TiDBSQLTemplate = "select round(%s,%d) as aa from %s where id=%d"
var OracleSQLTemplate = "select trim(to_char(round(%s,%d),'9999999999999.%s')) as aa from %s where id=%d"
var silentMode = true

func init() {
	flag.StringVar(&configPath, "config", "", "config file path")
	flag.StringVar(&tableName, "tbname", "t1", "table name")
	flag.StringVar(&colName, "colname", "col1", "column name")
	flag.IntVar(&threadCount, "thNum", 100, "connection thread count")
	flag.IntVar(&rangeMax, "count", 1000000, "max id")
	flag.BoolVar(&h, "h", false, "this help")
	flag.Parse()
}

func main() {

	flag.Parse()
	if h {
		flag.Usage()
		os.Exit(0)
	}
	cfg := config.InitConfig(configPath)
	// initialize log
	logger.InitLogger(cfg.Log.Level, cfg.Log.LogPath, cfg.Log)
	log.Info("Welcome to roundt")

	var tidbdb *sql.DB
	var oracledb *sql.DB
	var err error
	tidbdb, err = database.OpenMySQLDB(&cfg.TiDBConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Connect dest database error:%v", err))
	}
	oracledb, err = database.OpenOracleDB(&cfg.OracleConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Connect dest database error:%v", err))
	}
	wg := sync.WaitGroup{}
	wg.Add(threadCount)
	batchSize := rangeMax / threadCount
	remainderCount := rangeMax % threadCount
	t0 := time.Now()
	for i := 0; i < threadCount; i++ {

		go func(threadid int) {
			t1 := time.Now()
			log.Info(fmt.Sprintf("[Thread-%d] Started.", threadid))
			defer wg.Done()
			log.Debug(fmt.Sprintf("[Thread-%d] test from %d to %d ", threadid, threadid*batchSize+1, (threadid+1)*batchSize))
			for id := threadid*batchSize + 1; id <= (threadid+1)*batchSize; id++ {
				tidbSQLs := GenerateTiDBSQL(id, tableName, colName)
				oracleSQLs := GenerateOracleSQL(id, tableName, colName)
				err = compare.CompareSelect(tidbdb, oracledb, tidbSQLs, oracleSQLs, &cfg, silentMode)
				if err != nil {
					log.Error(fmt.Sprintf("Compare id %d error.error:%v", id, err))
				}
			}
			t2 := time.Now()
			td := t2.Sub(t1)
			log.Info(fmt.Sprintf("[Thread-%d] finished. id range:[%d,%d] Cost time: %f seconds.", i, threadid*batchSize+1, (threadid+1)*batchSize, td.Seconds()))

		}(i)
	}
	wg.Wait()
	if remainderCount > 0 {
		tt0 := time.Now()
		log.Debug(fmt.Sprintf("[Thread-remainder-%d] test from %d to %d ", remainderCount, batchSize*threadCount+1, rangeMax))
		for id := batchSize*threadCount + 1; id <= rangeMax; id++ {
			tidbSQLs := GenerateTiDBSQL(id, tableName, colName)
			oracleSQLs := GenerateOracleSQL(id, tableName, colName)
			err = compare.CompareSelect(tidbdb, oracledb, tidbSQLs, oracleSQLs, &cfg, silentMode)
			if err != nil {
				log.Error(fmt.Sprintf("Compare id %d error.error:%v", id, err))
			}
		}
		tt1 := time.Now()
		td := tt1.Sub(tt0)
		log.Info(fmt.Sprintf("[Thread-remainder-%d] finished. id range:[%d,%d] Cost time: %f seconds.", remainderCount, batchSize*threadCount+1, rangeMax, td.Seconds()))
	}
	t3 := time.Now()
	td := t3.Sub(t0)
	log.Info(fmt.Sprintf("Finished All. Cost time: %f seconds.", td.Seconds()))
}

// select round(col1,1) as aa from t1 where id=887766;
func GenerateTiDBSQL(id int, tbname string, colname string) []string {
	var res []string
	for i := 1; i <= 5; i++ {
		sql := fmt.Sprintf(TiDBSQLTemplate, colname, i, tbname, id)
		res = append(res, sql)
	}
	return res
}

//select trim(to_char(round(col1,1),'99999999999.9')) as aa from t1 where id=887766;
func GenerateOracleSQL(id int, tbname string, colname string) []string {
	var res []string
	for i := 1; i <= 5; i++ {
		sql := fmt.Sprintf(OracleSQLTemplate, colname, i, strings.Repeat("9", i), tbname, id)
		res = append(res, sql)
	}
	return res
}
