package models

import (
	"time"

	"github.com/Win-Man/dbcompare/database"
)

type SyncdiffConfigModel struct {
	Id                int       `json:"id" gorm:"type:int;autoIncrement,primaryKey"`
	TableSchema       string    `json:"table_schema" gorm:"type:varchar(64);uniqueIndex:uk_tab,priority:1;not null"`
	TableNameTidb     string    `json:"table_name" gorm:"column:table_name;type:varchar(64);uniqueIndex:uk_tab,priority:2;not null"`
	TableSchemaOracle string    `json:"table_schema_oracle" gorm:"type:varchar(64)"`
	Batchid           string    `json:"batchid" gorm:"type:varchar(128)"`
	TableCount        int       `json:"table_count" gorm:"type:int"`
	SyncStatus        string    `json:"sync_status" gorm:"type:varchar(32)"`
	SyncStarttime     time.Time `json:"sync_starttime" gorm:"type:datetime"`
	SyncEndtime       time.Time `json:"sync_endtime" gorm:"type:datetime"`
	SyncDuration      int       `json:"sync_duration" gorm:"type:int"`
	SyncMessages      string    `json:"sync_messages" gorm:"type:varchar(1000)"`
	JobStarttime      time.Time `json:"job_starttime" gorm:"type:datetime"`
	ChunkNum          int       `json:"chunk_num" gorm:"type:int"`
	CheckSuccessNum   int       `json:"check_success_num" gorm:"type:int"`
	CheckFailedNum    int       `json:"check_failed_num" gorm:"type:int"`
	CheckIgnoreNum    int       `json:"check_ignore_num" gorm:"type:int"`
	State             string    `json:"state" gorm:"type:varchar(32)"`
	ConfigHash        string    `gorm:"config_hash;type:varchar(50)" json:"config_hash"`
	UpdateTime        time.Time `gorm:"update_time;type:datetime" json:"update_time"`
	IgnoreColumns     string    `gorm:"ignore_columns;type:varchar(128)" json:"ignore_columns"`
	FilterClauseTidb  string    `gorm:"filter_clause_tidb;type:varchar(128)" json:"filter_clause_tidb"`
	FilterClauseOra   string    `gorm:"filter_clause_ora;type:varchar(128)" json:"filter_clause_ora"`
	ChunkSize         int       `gorm:"chunk_size;type:int;default:1000" json:"chunk_size"`
	CheckThreadCount  int       `gorm:"check_thread_count;type:int;default:10" json:"check_thread_count"`
	UseSnapshot       string    `gorm:"use_snapshot;type:varchar(10)" json:"use_snapshot"`
	SnapshotSource    string    `gorm:"snapshot_source;type:varchar(100)" json:"snapshot_source"`
	SnapshotTarget    string    `gorm:"snapshot_target;type:varchar(100)" json:"snapshot_target"`
	UseTso            string    `gorm:"use_tso;type:varchar(10)" json:"use_tso"`
	TsoInfo           string    `gorm:"tso_info;type:varchar(100)" json:"tso_info"`
	SourceInfo        string    `gorm:"source_info;type:varchar(256)" json:"source_info"`
	TargetInfo        string    `gorm:"target_info;type:varchar(256)" json:"target_info"`
	ContainDatatypes  string    `gorm:"contain_datatypes;type:varchar(256)" json:"contain_datatypes"`
	TableLabel        string    `gorm:"table_label;type:varchar(128)" json:"table_label"`
	Remark            string    `gorm:"remark;type:varchar(128)" json:"remark"`
}

//TableName of GORM model
func (it *SyncdiffConfigModel) TableName() string {
	return "syncdiff_config_result"
}

func (it *SyncdiffConfigModel) SelectById(id int) {
	query := database.DB.First(it, id)
	if query.Error != nil {
		panic(query.Error)
	}
}

func (it *SyncdiffConfigModel) SelectAll(id int) {
	query := database.DB.Find(it)
	if query.Error != nil {
		panic(query.Error)
	}
}

func (it *SyncdiffConfigModel) Insert() (err error) {
	query := database.DB.Create(&it)
	if query.Error != nil {
		return err
	}
	return nil
}

func (it *SyncdiffConfigModel) Update() (err error) {
	query := database.DB.Save(&it)
	if query.Error != nil {
		return err
	}
	return nil
}

func (it *SyncdiffConfigModel) Delete() (err error) {
	query := database.DB.Delete(&it)
	if query.Error != nil {
		return err
	}
	return nil
}

func (it *SyncdiffConfigModel) GetListPage(pageNum int, pageSize int) (err error, o2tConfigList []O2TConfigModel, count int64) {
	err = database.DB.Order("id asc").Offset(pageNum * pageSize).Limit(pageSize).Find(&o2tConfigList).Error
	database.DB.Table(it.TableName()).Count(&count)
	if err != nil {
		return err, o2tConfigList, 0
	}
	return nil, o2tConfigList, count
}
