/*
 * Created: 2021-11-04 23:28:44
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
	"github.com/Win-Man/dbcompare/config"
	"time"
)

// func init() {

// }

func OpenMySQLDB(dbcfg *config.DBConfig) (*sql.DB,error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbcfg.User, dbcfg.Password, dbcfg.Host, dbcfg.Port, dbcfg.Database))
	err = db.Ping()
	if err != nil{
		return nil,err
	}
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetMaxOpenConns(300)
	db.SetMaxIdleConns(100)
	return db,nil
}

func DoQuery(db *sql.DB,sql string)(*sql.Rows,error){
	rows,err := db.Query(sql)
	if err != nil{
		return rows,err
	}
	return rows,nil
}
