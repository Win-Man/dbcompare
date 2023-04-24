package cmd

import (
	"os"
	"testing"

	"github.com/Win-Man/dbcompare/config"
	"github.com/Win-Man/dbcompare/database"
	"github.com/Win-Man/dbcompare/models"
	log "github.com/sirupsen/logrus"
)

func TestGenerateSyncDiffConfig(t *testing.T) {
	tableSchema := "oracleSchema"
	tableName := "oracleTablename"
	tableSchemaTarget := "tableSchemaTarget"
	ignoreCols := "col1,col2"
	confDir := "./config/sync-diff-config.tmpl"
	chunkSize := 100
	checkThreadCount := 100
	snapSource := "snapsource"
	snapTarget := "snaptarget"
	task := models.SyncdiffConfigModel{
		FilterClauseTidb: "id < 199",
		FilterClauseOra:  "id < 199",
		IndexFields:      "id,age",
		OracleHint:       "/* hint for oracle */",
		TidbHint:         "/* hint for tidb */",
	}
	cfg := config.InitOTOConfig("../dev/oto.dev")
	var err error
	err = os.MkdirAll(cfg.SyncDiffControl.ConfDir, 0755)
	if err != nil {
		log.Error(err)
		t.Error(err)
	}

	err = generateSyncDiffConfig(tableSchema, tableName, tableSchemaTarget, ignoreCols,
		confDir, chunkSize, checkThreadCount, snapSource, snapTarget, task, cfg)
	if err != nil {

		t.Error(err)
	}
}

func TestRunSyncDiffControl(t *testing.T) {

	cfg := config.InitOTOConfig("../dev/oto.dev")
	err := database.InitDB(cfg.TiDBConfig)
	if err != nil {
		t.Error(err)
	}
	var records []models.SyncdiffConfigModel
	res := database.DB.Model(&models.SyncdiffConfigModel{}).Where("sync_status in (?,?)", SyncWaiting, SyncRunning).Scan(&records)
	if res.Error != nil {
		t.Error(res.Error)
	}

	for _, synctask := range records {
		tableSchema := synctask.TableSchema
		tableName := synctask.TableNameTidb
		tableSchemaTarget := synctask.TableSchemaOracle
		ignoreCols := synctask.IgnoreColumns
		confDir := cfg.SyncDiffControl.ConfDir
		chunkSize := synctask.ChunkSize
		checkThreadCount := synctask.CheckThreadCount
		snapSource := synctask.SnapshotSource
		snapTarget := synctask.SnapshotTarget
		var err error
		err = os.MkdirAll(cfg.SyncDiffControl.ConfDir, 0755)
		if err != nil {
			log.Error(err)
			t.Error(err)
		}

		err = generateSyncDiffConfig(tableSchema, tableName, tableSchemaTarget, ignoreCols,
			confDir, chunkSize, checkThreadCount, snapSource, snapTarget,
			synctask, cfg)
		if err != nil {
			t.Error(err)
		}
	}
}
