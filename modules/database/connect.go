package database

import (
	"fmt"
	"solana/utilities"
	"strings"
	"time"

	"github.com/fatih/color"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func parse() (map[string]string, error) {
	connStr := utilities.Get[string]("db_conn")
	if connStr == "" {
		return nil, fmt.Errorf("CONNECTION_URL environment variable is not set")
	}

	connStr = strings.ReplaceAll(connStr, "*", "")

	params := make(map[string]string)
	for _, pair := range strings.Split(connStr, ";") {
		if parts := strings.Split(pair, "="); len(parts) == 2 {
			params[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return params, nil
}

func Init() (*gorm.DB, error) {
	params, err := parse()
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		params["Host"], params["Port"], params["Username"], params["Password"], params["Database"])

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: time.Now().UTC,
		Logger:  logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	utilities.LogWithTimestamp(color.GreenString, "Connected to database")
	DB = db
	return db, nil
}

func Get() *gorm.DB {
	return DB
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get database instance: %v", err)
		}
		return sqlDB.Close()
	}
	return nil
}
