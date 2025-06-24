package db

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	"itmrchow/tw-media-analytics-service/domain/news/entity"
)

// InitMysqlDB 初始化 mysql db.
func InitMysqlDB(ctx context.Context, logger *zerolog.Logger, tracer trace.Tracer) *gorm.DB {
	// Trace
	ctx, span := tracer.Start(ctx, "domain/utils/db/InitMysqlDB: Init MysqlDB")
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
func InitSqliteDB(ctx context.Context, logger *zerolog.Logger, tracer trace.Tracer) *gorm.DB {
	// Trace
	ctx, span := tracer.Start(ctx, "domain/utils/db/InitSqliteDB: Init SqliteDB")
	logger.Info().Ctx(ctx).Msg("InitSqliteDb: start")
	defer func() {
		span.End()
		logger.Info().Ctx(ctx).Msg("InitSqliteDb end")
	}()

	db, err := initDB(ctx, sqlite.Open("./database.db"), &gorm.Config{}) // TODO: CTX
	if err != nil {
		logger.Fatal().Err(err).Ctx(ctx).Msg("failed to init sqlite db")
	}

	return db
}

// initDB 初始化 db.
func initDB(ctx context.Context, dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {

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

	return db, nil
}

// PingDB 呼叫 db.Ping() , 於初始化後呼叫.
func PingDB(ctx context.Context, logger *zerolog.Logger, tracer trace.Tracer, db *gorm.DB) error {
	// Trace
	ctx, span := tracer.Start(ctx, "domain/utils/db/PingDB: Ping DB")
	logger.Info().Ctx(ctx).Msg("PingDB: start")
	defer func() {
		span.End()
		logger.Info().Ctx(ctx).Msg("PingDB end")
	}()

	sqlDB, err := db.DB()
	if err != nil {
		logger.Err(err).Ctx(ctx).Msg("failed to get sql db")
		return err
	}
	logger.Info().Ctx(ctx).Msg("db initialized")
	err = sqlDB.Ping()
	if err != nil {
		logger.Err(err).Ctx(ctx).Msg("failed to ping db")
		return err
	}

	logger.Info().Ctx(ctx).Msg("db pinged")
	return nil
}
