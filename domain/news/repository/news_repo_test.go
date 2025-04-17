package repository

import (
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

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

	log := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)

	db, err := infra.InitSqliteDb()
	s.Require().NoError(err)
	sqlDB, err := db.DB()
	s.Require().NoError(err)

	// init test data
	fixtures, _ := testfixtures.New(
		testfixtures.Database(sqlDB),       // You database connection
		testfixtures.Dialect("sqlite"),     // Available: "postgresql", "timescaledb", "mysql", "mariadb", "sqlite" and "sqlserver"
		testfixtures.Directory("testdata"), // The directory containing the YAML files
		testfixtures.DangerousSkipTestDatabaseCheck(),
	)
	err = fixtures.Load()
	s.Require().NoError(err)

	s.newsRepo = NewNewsRepositoryImpl(&log, db)
}

func (s *NewsTestSuite) TestFindNonExistingNewsIDs() {
	mediaID := uint(1)
	newsIDList := []string{"1", "2", "3"}

	nonExistingNewsIDs, err := s.newsRepo.FindNonExistingNewsIDs(mediaID, newsIDList)
	s.NoError(err)
	s.Equal(nonExistingNewsIDs, []string{"2", "3"})
}
