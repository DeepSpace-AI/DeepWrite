package database

import (
	"database/sql"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB
var SQLDB *sql.DB

func Connect(dsn string, _logger gormlogger.Interface) (err error) {

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: _logger,
	})

	if err != nil {
		panic("failed to connect database")
	}

	SQLDB, err = DB.DB()

	if err != nil {
		panic("failed to get sql db from gorm db")
	}

	return err
}
