package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	"itmrchow/tw-media-analytics-service/domain/news/entity"
)

// InitMysqlDB 初始化 mysql db.
func InitMysqlDB(ctx context.Context) *gorm.DB {
	// Trace
	ctx, span := tracer.Start(ctx, "infra/InitMysqlDB: Init MysqlDB")
	logger.Info().Ctx(ctx).Msg("InitMysqlDb: start")
	defer func() {
		span.End()
		logger.Info().Ctx(ctx).Msg("InitMysqlDb end")
	}()

	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s",
		viper.GetString("MYSQL_DB_ACCOUNT"),
		viper.GetString("MYSQL_DB_PASSWORD"),
		viper.GetString("MYSQL_DB_HOST"),
		viper.GetString("MYSQL_DB_PORT"),
		viper.GetString("MYSQL_DB_NAME"),
		viper.GetString("MYSQL_URL_SUFFIX"),
	)

	db, err := initDB(ctx, mysql.Open(dns), &gorm.Config{})
	if err != nil {
		logger.Fatal().Err(err).Ctx(ctx).Msg("failed to init mysql db")
	}

	return db
}

// InitSqliteDB 初始化 sqlLite db.
func InitSqliteDB() *gorm.DB {
	db, err := initDB(context.Background(), sqlite.Open("./database.db"), &gorm.Config{}) // TODO: CTX
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init sqlite db")
	}

	return db
}

// initDB 初始化 db.
func initDB(ctx context.Context, dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
	// Trace
	ctx, span := tracer.Start(ctx, "infra/initDB: Init DB")
	logger.Info().Ctx(ctx).Msg("InitDB: start")
	defer func() {
		span.End()
		logger.Info().Ctx(ctx).Msg("InitDB end")
	}()

	db, err := gorm.Open(dialector, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Use OpenTelemetry plugin
	if err = db.Use(tracing.NewPlugin()); err != nil {
		return nil, fmt.Errorf("failed to use OpenTelemetry plugin: %w", err)
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
