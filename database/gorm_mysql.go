package database

import (
	"fmt"

	"github.com/Win-Man/dbcompare/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(cfg config.DBConfig) error {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	// DB, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
	// 	cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database))
	return err

}
