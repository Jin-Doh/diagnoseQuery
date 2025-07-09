package database

type SupportedDBType string

const (
	PostgreSQLDB SupportedDBType = "postgresql"
	MySQLDB      SupportedDBType = "mysql"
	MariaDB      SupportedDBType = "mariadb"
)

type PostgreSQL struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type MySQL struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type DBConnection struct {
	DBType     SupportedDBType
	PostgreSQL *PostgreSQL
	MySQL      *MySQL
}

type UnsupportedDBError struct {
	DBType SupportedDBType
}

func (e *UnsupportedDBError) Error() string {
	return "지원하지 않는 데이터베이스 유형: " + string(e.DBType)
}
func (e *UnsupportedDBError) Is(target error) bool {
	if _, ok := target.(*UnsupportedDBError); ok {
		return true
	}
	return false
}
