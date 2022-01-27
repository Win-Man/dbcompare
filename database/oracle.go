/*
 * Created: 2022-01-26 15:12:28
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
	"database/sql"
	"fmt"
	"time"

	"github.com/Win-Man/dbcompare/config"

	"github.com/godror/godror"
	_ "github.com/godror/godror"
)

// 创建 oracle 数据库引擎
func OpenOracleDB(oraCfg *config.OracleDBConfig) (*sql.DB, error) {

	//oralInfo := fmt.Sprintf("%s/%s@%s:%d/%s", oraCfg.User, oraCfg.Password, oraCfg.Host, oraCfg.Port, oraCfg.ServiceName)

	var p godror.ConnectionParams
	p.Username = oraCfg.User
	p.Password = godror.NewPassword(oraCfg.Password)
	p.ConnectString = fmt.Sprintf("%s:%d/%s", oraCfg.Host, oraCfg.Port, oraCfg.ServiceName)
	//TODO 设置时区
	p.Timezone = time.Now().UTC().Location()
	//p.SetSessionParamOnInit("time_zone", "+0:00")
	//db, _ := sql.Open("godror", oralInfo)
	//fmt.Println(p.StringWithPassword())
	db := sql.OpenDB(godror.NewConnector(p))
	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(0)
	db.SetConnMaxLifetime(0)

	err := db.Ping()
	if err != nil {
		return db, fmt.Errorf("error on ping oracle database connection:%v", err)
	}
	return db, nil
}
