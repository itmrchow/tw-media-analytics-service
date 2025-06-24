package repository

import (
	"context"
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"

	"itmrchow/tw-media-analytics-service/domain/news/entity"
	"itmrchow/tw-media-analytics-service/domain/utils/db"
	"itmrchow/tw-media-analytics-service/infra"
)

func TestAuthorRepoSuite(t *testing.T) {
	suite.Run(t, new(AuthorTestSuite))
}

type AuthorTestSuite struct {
	suite.Suite
	authorRepo AuthorRepository
}

func (s *AuthorTestSuite) SetupTest() {
	tracer := otel.Tracer("tw-media-analytics-service_test")
	infra.SetInfraTracer(tracer)

	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	infra.SetInfraLogger(&logger)

	db := db.NewSqliteDB(context.Background(), &logger, tracer)

	sqlDB, err := db.DB()
	s.Require().NoError(err)

	// 初始化測試資料
	fixtures, _ := testfixtures.New(
		testfixtures.Database(sqlDB),
		testfixtures.Dialect("sqlite"),
		testfixtures.Directory("testdata"),
		testfixtures.DangerousSkipTestDatabaseCheck(),
	)
	err = fixtures.Load()
	s.Require().NoError(err)

	s.authorRepo = NewAuthorRepositoryImpl(&logger, db)
}

func (s *AuthorTestSuite) TestFirstOrCreate_ExistingAuthor() {
	// 準備測試資料
	existingAuthor := &entity.Author{
		Name: "test author 1", // 根據 testdata/authors.yaml 的資料
	}

	// 執行測試
	err := s.authorRepo.FirstOrCreate(context.Background(), existingAuthor)

	// 驗證結果
	s.NoError(err)
	s.NotZero(existingAuthor.ID) // 確認有取得 ID
	s.Equal("test author 1", existingAuthor.Name)
}

func (s *AuthorTestSuite) TestFirstOrCreate_NewAuthor() {
	// 準備測試資料
	newAuthor := &entity.Author{
		Name: "test author 2",
	}

	// 執行測試
	err := s.authorRepo.FirstOrCreate(context.Background(), newAuthor)

	// 驗證結果
	s.NoError(err)
	s.NotZero(newAuthor.ID) // 確認有產生新的 ID
	s.Equal("test author 2", newAuthor.Name)

	// 再次查詢確認資料已建立
	checkAuthor := &entity.Author{
		Name: "test author 2",
	}
	err = s.authorRepo.FirstOrCreate(context.Background(), checkAuthor)
	s.NoError(err)
	s.Equal(newAuthor.ID, checkAuthor.ID)
}
