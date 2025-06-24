package repository

import (
	"context"
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"

	"itmrchow/tw-media-analytics-service/domain/utils/db"
	"itmrchow/tw-media-analytics-service/infra"
)

func TestNewsRepoSuite(t *testing.T) {
	suite.Run(t, new(NewsTestSuite))
}

type NewsTestSuite struct {
	suite.Suite
	newsRepo NewsRepository
	// db       *gorm.DB
}

func (s *NewsTestSuite) SetupTest() {
	tracer := otel.Tracer("tw-media-analytics-service_test")
	infra.SetInfraTracer(tracer)

	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	infra.SetInfraLogger(&logger)

	db := db.NewSqliteDB(context.Background(), &logger, tracer)

	sqlDB, err := db.DB()
	s.Require().NoError(err)

	// init test data
	fixtures, _ := testfixtures.New(
		testfixtures.Database(sqlDB), // You database connection
		testfixtures.Dialect(
			"sqlite",
		), // Available: "postgresql", "timescaledb", "mysql", "mariadb", "sqlite" and "sqlserver"
		testfixtures.Directory("testdata"), // The directory containing the YAML files
		testfixtures.DangerousSkipTestDatabaseCheck(),
	)
	err = fixtures.Load()
	s.Require().NoError(err)

	s.newsRepo = NewNewsRepositoryImpl(&logger, db)
}

func (s *NewsTestSuite) TestFindNonExistingNewsIDs() {
	mediaID := uint(1)
	newsIDList := []string{"1", "2", "3"}

	nonExistingNewsIDs, err := s.newsRepo.FindNonExistingNewsIDs(mediaID, newsIDList)
	s.NoError(err)
	s.Equal(nonExistingNewsIDs, []string{"2", "3"})
}
