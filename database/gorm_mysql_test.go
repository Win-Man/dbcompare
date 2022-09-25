package database

import (
	"testing"

	"github.com/Win-Man/dbcompare/config"
	_ "github.com/go-sql-driver/mysql"
)

func TestGORMConn(t *testing.T) {
	dbConfig := &config.DBConfig{
		Host:     "127.0.0.1",
		User:     "root",
		Password: "",
		Port:     4000,
		Database: "test",
	}
	InitDB(*dbConfig)

}
