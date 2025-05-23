package infra

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"itmrchow/tw-media-analytics-service/domain/news/entity"
)

func InitMysqlDb() *gorm.DB {
	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s",
		viper.GetString("MYSQL_DB_ACCOUNT"),
		viper.GetString("MYSQL_DB_PASSWORD"),
		viper.GetString("MYSQL_DB_HOST"),
		viper.GetString("MYSQL_DB_PORT"),
		viper.GetString("MYSQL_DB_NAME"),
		viper.GetString("MYSQL_URL_SUFFIX"),
	)

	db, err := initDB(mysql.Open(dns), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init mysql db")
	}

	return db
}

func InitSqliteDb() *gorm.DB {

	db, err := initDB(sqlite.Open("./database.db"), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init sqlite db")
	}

	return db
}

func initDB(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {

	db, err := gorm.Open(dialector, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Minute * 30)
	sqlDB.SetConnMaxIdleTime(15 * time.Minute)

	// Auto migrate all entities
	err = db.AutoMigrate(
		&entity.Media{},
		&entity.Author{},
		&entity.News{},
		&entity.Analysis{},
		&entity.AnalysisMetric{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
