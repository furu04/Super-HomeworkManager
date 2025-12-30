package database

import (
	"fmt"

	"homework-manager/internal/config"
	"homework-manager/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(dbConfig config.DatabaseConfig, debug bool) error {
	var logMode logger.LogLevel
	if debug {
		logMode = logger.Info
	} else {
		logMode = logger.Silent
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	}

	var db *gorm.DB
	var err error

	switch dbConfig.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbConfig.User,
			dbConfig.Password,
			dbConfig.Host,
			dbConfig.Port,
			dbConfig.Name,
		)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)

	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbConfig.Host,
			dbConfig.Port,
			dbConfig.User,
			dbConfig.Password,
			dbConfig.Name,
		)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)

	case "sqlite":
		fallthrough
	default:
		db, err = gorm.Open(sqlite.Open(dbConfig.Path), gormConfig)
	}

	if err != nil {
		return err
	}

	DB = db
	return nil
}

func Migrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.Assignment{},
		&models.APIKey{},
	)
}

func GetDB() *gorm.DB {
	return DB
}

