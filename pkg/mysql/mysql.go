/*
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-05-08 09:34:59
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-05-13 13:49:32
 * @FilePath: /pkg/sql/mysql.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */

package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	"golang.org/x/net/proxy"

	"github.com/go-sql-driver/mysql"
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

func DbConnect(dsn string, isproxy bool) *sql.DB {
	dialer, err := NewSocksDialer("127.0.0.1:2223", "", "")
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}

	if isproxy {
		db, err := InitDB(dsn, SocksProxy(dialer)) //返回sql.DB结构体指针类型对象
		if err != nil {
			panic("db连接发生错误")
		}
		return db
	} else {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			panic("db连接发生错误")
		}
		return db
	}
}

type Option func(*sql.DB)

func InitDB(dsn string, opts ...Option) (*sql.DB, error) {

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}

	for _, opt := range opts {
		opt(db)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}

	return db, err
}

func SocksProxy(dialer proxy.Dialer) Option {
	return func(d *sql.DB) {
		mysql.RegisterDialContext("tcp", func(ctx context.Context, addr string) (net.Conn, error) {
			return dialer.Dial("tcp", addr)
		})
	}
}

func NewSocksDialer(addr, user, password string) (proxy.Dialer, error) {
	return proxy.SOCKS5("tcp", addr, &proxy.Auth{User: user, Password: password}, proxy.Direct)
}
