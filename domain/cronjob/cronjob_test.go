package cronjob

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"itmrchow/tw-media-analytics-service/domain/utils"
	"itmrchow/tw-media-analytics-service/infra"
	mMock "itmrchow/tw-media-analytics-service/mock"
)

func TestCronJobSuite(t *testing.T) {
	suite.Run(t, new(CronJobTestSuite))
}

type CronJobTestSuite struct {
	suite.Suite
	cronJob       *CronJob
	mockPublisher *mMock.MockPublisher
	logger        *zerolog.Logger
	tracer        trace.Tracer
}

func (s *CronJobTestSuite) SetupTest() {
	viper.Set("ENV", "test")

	// 初始化 mock 物件
	s.logger = infra.InitLogger()
	infra.SetupOTelSDK(context.Background(), s.logger)

	s.mockPublisher = mMock.NewMockPublisher(s.T())
	s.tracer = otel.Tracer("tw-media-analytics-service")

	// 初始化 cronJob
	s.cronJob = NewCronJob(s.logger, s.tracer, s.mockPublisher)
}

func (s *CronJobTestSuite) TestArticleScrapingJob() {
	// input

	// mock
	s.mockPublisher.EXPECT().
		Publish("article_list_scraping", mock.Anything).
		Return(nil).
		Once()

	// expect
	s.cronJob.ArticleScrapingJob()
}

func (s *CronJobTestSuite) TestAnalyzeNewsJob() {
	// input

	// mock
	s.mockPublisher.EXPECT().
		Publish("analysis_get", mock.MatchedBy(func(msg interface{}) bool {
			// 先將 interface{} 轉換為 *message.Message
			messages, ok := msg.([]*message.Message)
			if !ok {
				s.T().Error("Failed to convert to *message.Message")
				return false
			}

			// 確保至少有一個訊息
			if len(messages) == 0 {
				s.T().Error("No messages provided")
				return false
			}

			// 取第一個訊息進行驗證
			var event utils.EventNewsAnalysis
			if err := json.Unmarshal(messages[0].Payload, &event); err != nil {
				s.T().Errorf("Failed to unmarshal message: %v", err)
				return false
			}

			if event.AnalysisNum != 2 {
				s.T().Errorf("Expected AnalysisNum to be 2, got %d", event.AnalysisNum)
				return false
			}

			return true
		})).
		Return(nil).
		Once()

	// expect
	s.cronJob.AnalyzeNewsJob()
}
