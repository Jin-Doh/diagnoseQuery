package database

import (
	"database/sql"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// NormalizeDBType normalizes database type string to supported type
func NormalizeDBType(dbTypeStr string) SupportedDBType {
	normalized := strings.ToLower(strings.TrimSpace(dbTypeStr))

	switch normalized {
	case "postgres", "postgresql", "pgsql":
		return PostgreSQLDB
	case "mysql":
		return MySQLDB
	case "mariadb":
		return MariaDB
	default:
		return SupportedDBType(normalized)
	}
}

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
	// Normalize the database type
	normalizedType := NormalizeDBType(string(dbType))

	if !IsSupportedDBType(normalizedType) {
		// 지원하지 않는 데이터베이스 유형입니다.
		return "", &UnsupportedDBError{DBType: dbType}
	}

	dsn, err := createDSN(normalizedType, host, port, user, password, dbname)
	if err != nil {
		return "", err
	}

	// Use the correct driver name for sql.Open
	var driverName string
	switch normalizedType {
	case PostgreSQLDB:
		driverName = "postgres"
	case MySQLDB, MariaDB:
		driverName = "mysql"
	default:
		driverName = string(normalizedType)
	}

	conn, err := sql.Open(driverName, dsn)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	if err = conn.Ping(); err != nil {
		return "", err
	}
	return dsn, nil
}

// ConnectWithString connects to database using string type (convenience function)
func ConnectWithString(dbTypeStr, host, port, user, password, dbname string) (string, error) {
	dbType := NormalizeDBType(dbTypeStr)
	return Connect(dbType, host, port, user, password, dbname)
}

func IsSupportedDBType(dbType SupportedDBType) bool {
	switch dbType {
	case PostgreSQLDB, MySQLDB, MariaDB:
		return true
	default:
		return false
	}
}
