package models

import (
	"testing"

	"github.com/Win-Man/dbcompare/config"
	"github.com/Win-Man/dbcompare/database"
)

func TestGORMConn(t *testing.T) {
	dbConfig := &config.DBConfig{
		Host:     "127.0.0.1",
		User:     "root",
		Password: "",
		Port:     4000,
		Database: "test",
	}
	database.InitDB(*dbConfig)

}

func TestGORMCreate(t *testing.T) {
	dbConfig := &config.DBConfig{
		Host:     "127.0.0.1",
		User:     "root",
		Password: "",
		Port:     4000,
		Database: "test",
	}
	database.InitDB(*dbConfig)

	if database.DB.Migrator().HasTable(&O2TConfigModel{}) {
		database.DB.Migrator().DropTable(&O2TConfigModel{})
	}
	database.DB.Migrator().CreateTable(&O2TConfigModel{})

	if database.DB.Migrator().HasTable(&T2OConfigModel{}) {
		database.DB.Migrator().DropTable(&T2OConfigModel{})
	}
	database.DB.Migrator().CreateTable(&T2OConfigModel{})

	if database.DB.Migrator().HasTable(&SyncdiffConfigModel{}) {
		database.DB.Migrator().DropTable(&SyncdiffConfigModel{})
	}
	database.DB.Migrator().CreateTable(&SyncdiffConfigModel{})

}

// func TestStructCreate(t *testing.T) {
// 	err := converter.NewTable2Struct().
// 		SavePath("./models.go").
// 		Dsn("root:@tcp(127.0.0.1:4000)/test?charset=utf8mb4").
// 		TagKey("gorm").
// 		EnableJsonTag(true).
// 		Table("syncdiff_config_result").
// 		Run()
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// }
