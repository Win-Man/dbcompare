/*
 * Created: 2022-01-26 15:52:52
 * Author : Win-Man
 * Email : gang.shen0423@gmail.com
 * -----
 * Last Modified:
 * Modified By:
 * -----
 * Description:
 */

package database

import (
	"fmt"
	"testing"

	"github.com/Win-Man/dbcompare/compare"
	"github.com/Win-Man/dbcompare/config"
	_ "github.com/go-sql-driver/mysql"
)

func TestMySQLConn(t *testing.T) {
	dbConfig := &config.DBConfig{
		Host:     "127.0.0.1",
		User:     "root",
		Password: "",
		Port:     4000,
		Database: "test",
	}
	_, err := OpenMySQLDB(dbConfig)
	if err != nil {
		t.Errorf("connnect mysql db error:%v", err)
	}

}

func TestMySQLSelect(t *testing.T) {
	dbConfig := &config.DBConfig{
		Host:     "127.0.0.1",
		User:     "root",
		Password: "",
		Port:     4000,
		Database: "test",
	}
	db, err := OpenMySQLDB(dbConfig)
	if err != nil {
		t.Errorf("connnect mysql db error:%v", err)
	}

	sqlStr := "select * from test.t"

	mysqlTable, err := compare.OutPrint(db, sqlStr)

	if err != nil {
		t.Errorf("execute sql error during query:%v", err)
	}

	fmt.Printf("%s", mysqlTable)
}
