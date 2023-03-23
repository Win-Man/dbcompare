package cmd

import (
	"testing"

	"github.com/Win-Man/dbcompare/config"
)

func TestGenerateSyncDiffConfig(t *testing.T) {
	// generateSyncDiffConfig(tableSchema string, tableName string, tableSchemaTarget string, ignoreCols string,
	// 	confDir string, chunkSize int, checkThreadCount int, snapSource string, snapTarget string,
	// 	filterClauseTidb string, filterClauseOra string, cfg config.OTOConfig)
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
	cfg := config.InitOTOConfig("../config/oto.dev")
	err := generateSyncDiffConfig(tableSchema, tableName, tableSchemaTarget, ignoreCols,
		confDir, chunkSize, checkThreadCount, snapSource, snapTarget,
		filterClauseTidb, filterClauseOra, cfg)
	if err != nil {

		t.Error(err)
	}
}
