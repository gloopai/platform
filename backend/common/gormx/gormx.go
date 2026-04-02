// Package gormx provides small MySQL/GORM helpers shared by product services.
package gormx

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MustOpenMySQL opens a MySQL connection via GORM and panics if open or ping fails.
func MustOpenMySQL(dsn string) *gorm.DB {
	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		panic(err)
	}
	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}
	return gdb
}
