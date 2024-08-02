package database

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB(dataSourceName string) (*sql.DB, error) {
	var err error
	DB, err = sql.Open("mysql", dataSourceName)
	if err != nil {
			return nil, err
	}

	// 測試連接
	err = DB.Ping()
	if err != nil {
			return nil, err
	}

	return DB, nil
}
// 其他資料庫相關函數...