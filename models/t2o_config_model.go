package models

import (
	"time"

	"github.com/Win-Man/dbcompare/database"
)

type T2OConfigModel struct {
	Id                      int       `json:"id" gorm:"type:int;autoIncrement,primaryKey"`
	TableSchemaTidb         string    `json:"table_schema_tidb" gorm:"type:varchar(64);uniqueIndex:uk_tab,priority:1;not null"`
	TableNameTidb           string    `json:"table_name_tidb" gorm:"type:varchar(64);uniqueIndex:uk_tab,priority:2;not null"`
	TableSchemaOracle       string    `json:"table_schema_oracle" gorm:"type:varchar(64)"`
	DumpStatus              string    `json:"dump_status" gorm:"type:varchar(20);not null"`
	DumpDuration            int       `json:"dump_duration" gorm:"type:int"`
	LastDumpTime        time.Time `json:"last_dump_time" gorm:"type:datetime"`
	GenerateCtlStatus       string    `json:"generate_ctl_status" gorm:"type:varchar(20);not null"`
	GenerateCtlDuration     int       `json:"generate_ctl_duration" gorm:"type:int"`
	LastGenerateCtlTime time.Time `json:"last_generate_ctl_time" gorm:"type:datetime"`
	LoadStatus              string    `json:"load_status" gorm:"type:varchar(20);not null"`
	LoadDuration            int       `json:"load_duration" gorm:"type:int"`
	LastLoadTime        time.Time `json:"last_load_time" gorm:"type:datetime"`
}

//TableName of GORM model
func (it *T2OConfigModel) TableName() string {
	return "t2o_config"
}

func (it *T2OConfigModel) SelectById(id int) {
	query := database.DB.First(it, id)
	if query.Error != nil {
		panic(query.Error)
	}
}

func (it *T2OConfigModel) SelectAll(id int) {
	query := database.DB.Find(it)
	if query.Error != nil {
		panic(query.Error)
	}
}

func (it *T2OConfigModel) Insert() (err error) {
	query := database.DB.Create(&it)
	if query.Error != nil {
		return err
	}
	return nil
}

func (it *T2OConfigModel) Update() (err error) {
	query := database.DB.Save(&it)
	if query.Error != nil {
		return err
	}
	return nil
}

func (it *T2OConfigModel) Delete() (err error) {
	query := database.DB.Delete(&it)
	if query.Error != nil {
		return err
	}
	return nil
}

func (it *T2OConfigModel) GetListPage(pageNum int, pageSize int) (err error, o2tConfigList []O2TConfigModel, count int64) {
	err = database.DB.Order("id asc").Offset(pageNum * pageSize).Limit(pageSize).Find(&o2tConfigList).Error
	database.DB.Table(it.TableName()).Count(&count)
	if err != nil {
		return err, o2tConfigList, 0
	}
	return nil, o2tConfigList, count
}
