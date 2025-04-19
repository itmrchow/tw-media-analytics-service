package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"itmrchow/tw-media-analytics-service/domain/ai_model"
	"itmrchow/tw-media-analytics-service/domain/cron_job"
	news "itmrchow/tw-media-analytics-service/domain/news/delivery"
	"itmrchow/tw-media-analytics-service/domain/news/repository"
	"itmrchow/tw-media-analytics-service/domain/news/service"
	"itmrchow/tw-media-analytics-service/domain/queue"
	spider "itmrchow/tw-media-analytics-service/domain/spider/delivery"
	"itmrchow/tw-media-analytics-service/domain/utils"
	"itmrchow/tw-media-analytics-service/infra"
)

func main() {
	// 系統信號處理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// config
	infra.InitConfig()

	// logger
	logger := initLogger()

	ctx, cancel := context.WithCancel(context.Background())

	// ai model
	model := ai_model.NewGemini(logger, ctx)

	// db
	db := infra.InitMysqlDb()

	// queue
	q := initQueue(ctx, logger)

	// cron
	initCron(logger, q)

	newsRepo := repository.NewNewsRepositoryImpl(logger, db)
	authorRepo := repository.NewAuthorRepositoryImpl(logger, db)

	// service
	newsService := service.NewNewsServiceImpl(logger, newsRepo, authorRepo, q, db)

	// handler
	// - Spider handler
	ctiSpiderEventHandler := spider.NewCtiNewsNewsSpiderEventHandler(logger, q) // 中天
	setnSpiderEventHandler := spider.NewSetnNewsSpiderEventHandler(logger, q)   // 三立

	// - news handler
	newsHandler := news.NewNewsEventHandler(logger, q, db, newsService)

	// consumer
	go func() {
		if err := initConsumer(ctx, q, []*spider.SpiderEventHandler{ctiSpiderEventHandler, setnSpiderEventHandler}, newsHandler); err != nil {
			log.Err(err).Msg("failed to init consumer")
			cancel()
		}
	}()

	// try publish message
	msg := utils.EventArticleListScraping{}

	q.Publish(ctx, queue.TopicArticleListScraping, msg)

	defer func() {
		q.CloseClient()
		model.CloseClient()

		logger.Info().Msg("Client closed")

	}()

	select {
	case sig := <-sigChan:
		logger.Info().Msgf("收到系統信號: %v, 開始關閉服務", sig)
		cancel()
	case <-ctx.Done():
		logger.Info().Msg("服務開始關閉")
	}
}

func tryAnalyzeNews(model *ai_model.Gemini) {

	// 	title := "不爽舞藝差「只要是韓籍就可當啦啦隊」？ 慈妹4字回應"
	// 	content := `
	// 	富邦啦啦隊Fubon Angels人氣成員慈妹（本名：彭翊慈），去年底發了黑白低潮文，自曝「忍耐到極限」。最近她上沈玉琳的YouTube節目《威廉沈歡樂送》，被問到是否韓籍成員無論舞藝好壞，都可以加入啦啦隊，讓她感覺壓力？她表示「不予置評」，且補充有些台籍啦啦隊成員可能也舞藝不佳，但重點是大家都很努力。
	//  慈妹最近現身沈玉琳的節目，開頭沈玉琳便關心慈妹心情，打趣說：「她（慈妹）受不了台灣各大球團，現在怎麼樣？只要是韓國人就可以來當啦啦隊是不是？聽說有的球團到韓國街頭（拉人）來做啦啦隊，『我不會跳舞捏』、『韓國人就好』，聽說現在只要是妳的身分證是韓國人就可以被拉來台灣當啦啦隊？」
	//  聞此，慈妹回應：「這個喔，我真的不予置評耶」，沈玉琳追問，是否真的有韓國籍但不太會跳舞卻在當啦啦隊？慈妹回應：「我覺得其實多多少少，也是蠻多本土的啦啦隊，他們也其實不太會跳舞」，「但是大家都很努力，我覺得這是一個重點」。
	//  沈玉琳接著追問那去年慈妹為何低潮？她回應可能開始上節目或做直播後，有時候為了節目效果，就會比較誇大，直播間就會因此湧入很多她不認識的人說她工作上的事情，「我可能就會有點被影響到了吧。」然後剛開始她也不太喜歡打字發文，就是不太喜歡評論。沈玉琳說：「我跟妳講這是演藝必經之路，每個人都這樣走過來的，千萬不要為了一點人家的留言評論，然後在那裡尋死尋活，妳就繼續做妳自己。」他也幫慈妹打氣，做這行要付出的成本就是會被別人酸言酸語，要養成強大的內心才走得下去。
	// 	`

	// 	newsAnalytics, err := model.AnalyzeNews(title, content)
	// 	if err != nil {
	// 		log.Fatal().Err(err).Msg("analyze news error")
	// 	}

	// fmt.Println(newsAnalytics)
}

func initLogger() *zerolog.Logger {
	// TODO: setting log level
	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel)
	logger = logger.With().Str("service", "tw-media-analytics-service").Logger()
	return &logger
}

func initCron(logger *zerolog.Logger, queue queue.Queue) {

	jobs := cron_job.NewCronJob(logger, queue)

	c := cron.New()
	_, err := c.AddFunc("0 * * * *", jobs.ArticleScrapingJob)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to add cron job")
	}
	c.Start()
	logger.Info().Msg("cron job started")
}

func initQueue(ctx context.Context, logger *zerolog.Logger) queue.Queue {

	// create q obj
	q := queue.NewGcpPubSub(ctx, logger)

	// init topic
	err := q.InitTopic()
	if err == nil {
		logger.Info().Msg("Queue topic created")
	} else {
		logger.Fatal().Err(err).Msg("failed to create topic")
	}

	return q
}

func initConsumer(ctx context.Context, q queue.Queue,
	spiderHandlerList []*spider.SpiderEventHandler,
	newsHandler *news.NewsEventHandler,
) (err error) {
	// set subscription
	var group errgroup.Group

	// - ArticleListScraping
	for mediaID, s := range spiderHandlerList {
		group.Go(func() error {
			mediaID++
			subID := fmt.Sprintf("%s_%s_%v_sub", string(queue.TopicArticleListScraping), viper.GetString("ENV"), mediaID)
			return q.Consume(ctx, queue.TopicArticleListScraping, subID, s.ArticleListScrapingHandle)
		})
	}

	// - CheckNewsExist
	group.Go(func() error {
		return q.Consume(ctx, queue.TopicNewsCheck, "", newsHandler.CheckNewsExistHandle)
	})

	// - SaveNews
	group.Go(func() error {
		return q.Consume(ctx, queue.TopicNewsSave, "", newsHandler.SaveNewsHandle)
	})

	return group.Wait()
}
