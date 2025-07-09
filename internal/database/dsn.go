package database

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func createDSN(dbType SupportedDBType, host, port, user, password, dbname string) (string, error) {
	switch dbType {
	case PostgreSQLDB:
		return "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname + "?sslmode=disable", nil
	case MySQLDB, MariaDB:
		return "mysql://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname, nil
	default:
		return "", &UnsupportedDBError{DBType: dbType}
	}
}

func Connect(dbType SupportedDBType, host, port, user, password, dbname string) (string, error) {
	if !IsSupportedDBType(dbType) {
		// 지원하지 않는 데이터베이스 유형입니다.
		return "", &UnsupportedDBError{DBType: dbType}
	}

	dsn, err := createDSN(dbType, host, port, user, password, dbname)
	if err != nil {
		return "", err
	}
	conn, err := sql.Open(string(dbType), dsn)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	if err = conn.Ping(); err != nil {
		return "", err
	}
	return dsn, nil
}

func IsSupportedDBType(dbType SupportedDBType) bool {
	switch dbType {
	case PostgreSQLDB, MySQLDB, MariaDB:
		return true
	default:
		return false
	}
}
