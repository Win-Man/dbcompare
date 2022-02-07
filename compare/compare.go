/*
 * Created: 2021-11-04 22:48:23
 * Author : Win-Man
 * Email : gang.shen0423@gmail.com
 * -----
 * Last Modified:
 * Modified By:
 * -----
 * Description:
 */

package compare

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/Win-Man/dbcompare/config"
	log "github.com/sirupsen/logrus"
)

func init() {

}

const NILVALUE = "NULL"

type Table struct {
	SQLStr       string
	ColumnHeader []string
	RecordList   [][]string
}

func (t *Table) AddRecord(slist []string) {
	t.RecordList = append(t.RecordList, slist)
}

func (t *Table) String() string {
	resStr := fmt.Sprintf("%s\n%s", t.SQLStr, strings.Join(t.ColumnHeader, "\t"))
	for _, rlist := range t.RecordList {
		resStr = fmt.Sprintf("%s\n%s", resStr, strings.Join(rlist, "\t"))
	}
	return resStr
}

func (t *Table) ResultString() string {
	resStr := fmt.Sprintf("%s", strings.Join(t.ColumnHeader, "\t"))
	for _, rlist := range t.RecordList {
		resStr = fmt.Sprintf("%s\n%s", resStr, strings.Join(rlist, "\t"))
	}
	return resStr
}

func CompareSelect(sourcedb *sql.DB, destdb *sql.DB, SourceSqls []string, DestSqls []string, cfg *config.Config) error {
	log.Debug(fmt.Sprintf("SourceSqls:%v\nDestSqls:%v", SourceSqls, DestSqls))
	log.Debug(fmt.Sprintf("source type:%s dest type:%s", cfg.CompareConfig.SourceType, cfg.CompareConfig.DestType))
	// mysqlRows,err := DoQuery(mysqldb,sql)
	// if err != nil{
	// 	log.Error(err)
	// }
	// tidbRows,err := DoQuery(tidbdb,sql)
	if len(SourceSqls) != len(DestSqls) {
		log.Fatal(fmt.Scanf("The number of sqls in sqlfile-source is different with number of sqls in sqlfile-dest:%d vs %d", len(SourceSqls), len(DestSqls)))
	}
	timeStr := time.Now().Format("20060102150504")
	sourceFilePath := fmt.Sprintf("%s_%s_%s", cfg.CompareConfig.OutputPrefix, timeStr, strings.ToLower(cfg.CompareConfig.SourceType))
	destFilePath := fmt.Sprintf("%s_%s_%s", cfg.CompareConfig.OutputPrefix, timeStr, strings.ToLower(cfg.CompareConfig.DestType))
	diffFilePath := fmt.Sprintf("%s_%s_diff", cfg.CompareConfig.OutputPrefix, timeStr)
	sourceTypeStr := cfg.CompareConfig.SourceType
	destTypeStr := cfg.CompareConfig.DestType
	var diffSQLList []string
	for i, sql := range SourceSqls {
		var sourceTable, destTable *Table
		var sql1 = sql
		var sql2 = DestSqls[i]
		var err error
		if strings.ToLower(sourceTypeStr) == "mysql" || strings.ToLower(sourceTypeStr) == "tidb" {
			sourceTable, err = OutPrint(sourcedb, sql1)
		} else {
			sourceTable, err = OutPrintOracle(sourcedb, sql1)
		}

		if err != nil {
			diffSQLList = append(diffSQLList, fmt.Sprintf("ERROR SOURCE SQL:\n%s\n%s", sql1, strings.Repeat("-", 20)))
			continue
		}
		if strings.ToLower(destTypeStr) == "mysql" || strings.ToLower(destTypeStr) == "tidb" {
			destTable, err = OutPrint(destdb, sql2)
		} else {
			destTable, err = OutPrintOracle(destdb, sql2)
		}
		if err != nil {
			diffSQLList = append(diffSQLList, fmt.Sprintf("ERROR DEST SQL:\n%s\n%s", sql2, strings.Repeat("-", 20)))
			continue
		}

		if sourceTable.ResultString() != destTable.ResultString() {
			diffSQLList = append(diffSQLList, fmt.Sprintf("DIFF SOURCE SQL:\n%s\nDIFF DEST SQL:\n%s\n%s", sql1, sql2, strings.Repeat("-", 20)))
			log.Warn("SOURCE result is different with DEST returns")
			log.Warn(sourceTable.String())
			log.Warn(destTable.String())
		}

		if cfg.CompareConfig.Output == "print" {
			fmt.Printf("SOUECE output:\n")
			fmt.Println(sourceTable)
			fmt.Printf("DEST output:\n")
			fmt.Println(destTable)
		} else if cfg.CompareConfig.Output == "file" {
			WriteTableFile(sourceFilePath, sourceTable)
			WriteTableFile(destFilePath, destTable)
		}
		log.Info(fmt.Sprintf("Compare No.%d SQL Done.", i))
		log.Debug(fmt.Sprintf("Compare Source SQL Done.%s", sql1))
		log.Debug(fmt.Sprintf("Compare Dest SQL Done.%s", sql2))
	}
	if len(diffSQLList) != 0 {
		if cfg.CompareConfig.Output == "print" {
			fmt.Printf("Diff SQLs:\n")
			for _, sql := range diffSQLList {
				fmt.Printf("%s\n", sql)
			}
		} else if cfg.CompareConfig.Output == "file" {
			for _, sql := range diffSQLList {
				WriteFile(diffFilePath, sql)
			}
		}

	}
	return nil
}

func WriteTableFile(filePath string, t *Table) error {
	_, err := os.Stat(filePath)
	if err != nil {
		os.Create(filePath)
	}
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	w.WriteString(t.String())
	w.Flush()
	return nil
}

func WriteFile(filePath string, content string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		os.Create(filePath)
	}
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	w.WriteString(fmt.Sprintf("%s\n", content))
	w.Flush()
	return nil
}

func OutPrint(db *sql.DB, sql string) (*Table, error) {
	var outTable Table
	rows, err := db.Query(sql)
	if err != nil {
		log.Error(err)
		return &outTable, err
	}
	columnTypes, _ := rows.ColumnTypes()
	//fmt.Printf("columnTypes:%v\n", columnTypes)
	var rowParam = make([]interface{}, len(columnTypes)) // 传入到 rows.Scan 的参数 数组
	var rowValue = make([]interface{}, len(columnTypes)) // 接收数据一行列的数组

	var columnHeader []string
	for i, colType := range columnTypes {
		columnHeader = append(columnHeader, strings.ToUpper(colType.Name()))
		rowValue[i] = reflect.New(colType.ScanType())           // 跟据数据库参数类型，创建默认值 和类型
		rowParam[i] = reflect.ValueOf(&rowValue[i]).Interface() // 跟据接收的数据的类型反射出值的地址
	}
	var list []map[string]string
	outTable.ColumnHeader = columnHeader
	outTable.SQLStr = sql
	for rows.Next() {
		_ = rows.Scan(rowParam...)
		item := make(map[string]string)
		var valueList []string
		for i, colType := range columnTypes {
			//fmt.Printf("colType:%+v colscanType:%s\n", colType, colType.ScanType().String())
			typeName := colType.DatabaseTypeName()
			if rowValue[i] == nil {
				item[colType.Name()] = NILVALUE
				valueList = append(valueList, NILVALUE)
			} else {
				switch typeName {
				case "VARCHAR", "CHAR", "DATETIME", "TIMESTAMP", "INT":
					//item[colType.Name()] = reflect.ValueOf(rowValue[i]).String()
					item[colType.Name()] = string(rowValue[i].([]byte))
				// case "FLOAT":
				// 	item[colType.Name()], _ = strconv.ParseFloat(string(rowValue[i].([]byte)), 64)
				// 	// item[colType.Name()], _ = rowValue[i].(float64)
				default:
					item[colType.Name()] = string(rowValue[i].([]byte))
				}
				valueList = append(valueList, item[colType.Name()])
			}
		}
		list = append(list, item)
		outTable.AddRecord(valueList)
	}
	rows.Close()
	return &outTable, nil
}

func OutPrintOracle(db *sql.DB, sql string) (*Table, error) {
	var outTable Table
	rows, err := db.Query(sql)
	if err != nil {
		log.Error(err)
		return &outTable, err
	}
	columnTypes, _ := rows.ColumnTypes()
	var rowParam = make([]interface{}, len(columnTypes)) // 传入到 rows.Scan 的参数 数组
	var rowValue = make([]interface{}, len(columnTypes)) // 接收数据一行列的数组

	var columnHeader []string
	for i, colType := range columnTypes {
		columnHeader = append(columnHeader, colType.Name())
		rowValue[i] = reflect.New(colType.ScanType())           // 跟据数据库参数类型，创建默认值 和类型
		rowParam[i] = reflect.ValueOf(&rowValue[i]).Interface() // 跟据接收的数据的类型反射出值的地址
	}
	var list []map[string]string
	outTable.ColumnHeader = columnHeader
	outTable.SQLStr = sql
	for rows.Next() {
		_ = rows.Scan(rowParam...)
		item := make(map[string]string)
		var valueList []string
		for i, colType := range columnTypes {
			typeName := colType.DatabaseTypeName()
			if rowValue[i] == nil {
				item[colType.Name()] = NILVALUE
				valueList = append(valueList, NILVALUE)
			} else {
				//https://github.com/godror/godror/blob/00248a71fb884addb85dd0956178afd35469bfce/rows.go#L136
				switch typeName {
				case "VARCHAR2", "NVARCHAR2", "CHAR", "NCHAR", "LONG", "NUMBER", "INTERVAL YEAR TO MONTH",
					"CLOB", "NCLOB", "OBJECT":
					item[colType.Name()] = reflect.ValueOf(rowValue[i]).String()
				case "RAW", "ROWID", "LONG RAW", "BLOB", "BFILE":
					item[colType.Name()] = string(rowValue[i].([]byte))
				case "FLOAT", "DOUBLE":
					item[colType.Name()] = fmt.Sprintf("%f", reflect.ValueOf(rowValue[i]).Float())
				case "BINARY_INTEGER":
					//TODO: uint
					item[colType.Name()] = fmt.Sprintf("%d", reflect.ValueOf(rowValue[i]).Int())
				case "BOOLEAN":
					item[colType.Name()] = fmt.Sprintf("%t", reflect.ValueOf(rowValue[i]).Bool())
				case "TIMESTAMP", "TIMESTAMP WITH TIME ZONE", "TIMESTAMP WITH LOCAL TIME ZONE", "DATE":
					item[colType.Name()] = reflect.ValueOf(rowValue[i]).String()
				case "INTERVAL DAY TO SECOND":
					item[colType.Name()] = reflect.ValueOf(rowValue[i]).String()
				case "JSON":
					item[colType.Name()] = reflect.ValueOf(rowValue[i]).String()
				default:
					return nil, errors.New(fmt.Sprintf("Unknow column type:%s column name:%s", typeName, colType.Name()))
				}
				valueList = append(valueList, item[colType.Name()])
			}

		}
		list = append(list, item)
		outTable.AddRecord(valueList)
	}
	rows.Close()
	return &outTable, nil
}
