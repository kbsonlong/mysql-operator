/*
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-05-08 09:34:59
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-05-08 22:59:10
 * @FilePath: /mysql-operator/Users/zengshenglong/Code/GoWorkSpace/operators/mysql-operator/pkg/sql/mysql.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */

package sql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// dsn := GetDsn(map[string]interface{}{"user_name":"root","password":"88888888"})
func GetDsn(dbConfig map[string]interface{}) string {
	user_name, ok := dbConfig["username"]
	if !ok {
		user_name = "root"
	}
	password, ok := dbConfig["password"]
	if !ok {
		password = ""
	}
	host, ok := dbConfig["host"]
	if !ok {
		host = "localhost"
	}
	port, ok := dbConfig["port"]
	if !ok {
		port = "3306"
	}
	dbname, ok := dbConfig["dbname"]
	if !ok {
		dbname = ""
	}
	charset, ok := dbConfig["charset"]
	if !ok {
		charset = "utf8"
	}

	param, ok := dbConfig["param"]
	if ok && param != "" {
		param = fmt.Sprintf("&%s", param)
	} else {
		param = ""
	}
	dbDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s%s", user_name, password, host, port, dbname, charset, param)
	return dbDsn
}

func DbConnect(dsn string) *sql.DB {
	db, err := sql.Open("mysql", dsn) //返回sql.DB结构体指针类型对象
	if err != nil {
		panic("db连接发生错误")
	}
	// err = db.Ping()
	// if err != nil {
	// 	panic("db连接失败")
	// }
	return db
}
