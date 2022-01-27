/*
 * Created: 2022-01-26 15:25:34
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
	_ "github.com/godror/godror"
)

func TestOracleConn(t *testing.T) {
	oraConfig := config.OracleDBConfig{}
	oraConfig.Host = "172.16.6.203"
	oraConfig.User = "c##gangshen"
	oraConfig.Password = "123456"
	oraConfig.Port = 1539
	oraConfig.ServiceName = "orcl"

	_, err := OpenOracleDB(&oraConfig)
	if err != nil {
		t.Errorf("%v", err)
	}

}

func TestOracleSelect(t *testing.T) {
	oraConfig := config.OracleDBConfig{}
	oraConfig.Host = "172.16.6.203"
	oraConfig.User = "c##gangshen"
	oraConfig.Password = "123456"
	oraConfig.Port = 1539
	oraConfig.ServiceName = "orcl"

	db, err := OpenOracleDB(&oraConfig)
	if err != nil {
		t.Errorf("%v", err)
	}

	sqlStr := "select * from t"

	mysqlTable, err := compare.OutPrintOracle(db, sqlStr)

	if err != nil {
		t.Errorf("execute sql error during query:%v", err)
	}

	fmt.Printf("%s\n", mysqlTable)

	// defer db.Close()
	// rows, err := db.Query(sqlStr)
	// if err != nil {
	// 	t.Errorf("execute sql error during query:%v", err)
	// }
	// cols,_ := rows.Columns()
	// fmt.Printf("Result columns:%s\n",cols)
	// var id int
	// var name string
	// defer rows.Close()
	// for rows.Next(){
	// 	rows.Scan(&id,&name)
	// 	fmt.Printf("id:%d name:%s\n",id,name)
	// }

}
