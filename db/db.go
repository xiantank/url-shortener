package db

import (
	"fmt"

	"github.com/xiantank/url-shortener/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const dsnTemplate = "%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local"

func New() (*gorm.DB, error) {
	dsn := fmt.Sprintf(dsnTemplate, config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
