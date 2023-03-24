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
	filterClauseTidb := "id < 199"
	filterClauseOra := "id < 199"
	indexFields := "id,age"
	cfg := config.InitOTOConfig("../dev/oto.dev")
	var err error
	err = os.MkdirAll(cfg.SyncDiffControl.ConfDir, 0755)
	if err != nil {
		log.Error(err)
		t.Error(err)
	}

	err = generateSyncDiffConfig(tableSchema, tableName, tableSchemaTarget, ignoreCols,
		confDir, chunkSize, checkThreadCount, snapSource, snapTarget,
		filterClauseTidb, filterClauseOra, indexFields, cfg)
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
		filterClauseTidb := synctask.FilterClauseTidb
		filterClauseOra := synctask.FilterClauseOra
		indexFields := synctask.IndexFields
		var err error
		err = os.MkdirAll(cfg.SyncDiffControl.ConfDir, 0755)
		if err != nil {
			log.Error(err)
			t.Error(err)
		}

		err = generateSyncDiffConfig(tableSchema, tableName, tableSchemaTarget, ignoreCols,
			confDir, chunkSize, checkThreadCount, snapSource, snapTarget,
			filterClauseTidb, filterClauseOra, indexFields, cfg)
		if err != nil {
			t.Error(err)
		}
	}
}
