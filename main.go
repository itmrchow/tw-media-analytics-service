package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"itmrchow/tw-media-analytics-service/domain/ai_model"
	"itmrchow/tw-media-analytics-service/domain/cron_job"
	"itmrchow/tw-media-analytics-service/domain/news/entity"
	"itmrchow/tw-media-analytics-service/domain/queue"
	"itmrchow/tw-media-analytics-service/domain/spider"
)

func main() {

	initConfig()
	logger := initLogger()

	// test part
	ctx := context.Background()
	model := ai_model.NewGemini(logger, ctx)

	// Spider
	ctiSpider := spider.NewCtiNewsSpider(logger) // 中天
	setnSpider := spider.NewSetnSpider(logger)   // 三立

	// db
	_, err := InitMysqlDb()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init mysql db")
	}

	// cron
	initCron(logger)

	// queue
	q := initQueue(ctx, logger)

	// set subscription
	var g errgroup.Group
	g.Go(func() error {
		return q.Consume(ctx, queue.TopicArticleScraping, ctiSpider.ArticleScrapingHandle)
	})

	g.Go(func() error {
		return q.Consume(ctx, queue.TopicArticleScraping, setnSpider.ArticleScrapingHandle)
	})

	// try publish message
	msg := spider.GetNewsEvent{}
	q.Publish(ctx, queue.TopicArticleScraping, msg)

	defer func() {
		q.CloseClient()
		model.CloseClient()

	}()
	// select {}
	if err := g.Wait(); err != nil {
		log.Fatal().Err(err).Msg("Error occurred in goroutines")
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

func initConfig() {
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("config init error")
	}

	log.Info().Msgf("config init success")
}

func initLogger() *zerolog.Logger {
	// TODO: setting log level
	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel)
	logger = logger.With().Str("service", "tw-media-analytics-service").Logger()
	return &logger
}

func initCron(logger *zerolog.Logger) {

	jobs := cron_job.NewCronJob(logger)

	c := cron.New()
	_, err := c.AddFunc("0 * * * *", jobs.ArticleScrapingJob)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to add cron job")
	}
	c.Start()
	log.Info().Msg("cron job started")
}

func initQueue(ctx context.Context, logger *zerolog.Logger) queue.Queue {

	// create q obj
	q := queue.NewGcpPubSub(ctx, logger)

	// create topic
	err := q.CreateTopic()
	if err == nil {
		log.Info().Msg("Queue topic created")
	} else {
		log.Fatal().Err(err).Msg("failed to create topic")
	}

	// try publish message

	// log.Debug().Str("topic", string(queue.TopicArticleScraping+"_dev")).Msg("try publish message")
	// err = q.Publish(ctx, queue.TopicArticleScraping+"_dev", "test")
	// if err == nil {
	// 	log.Info().Msg("Queue message published")
	// } else {
	// 	log.Fatal().Err(err).Msg("failed to publish message")
	// }

	// create consumer

	return q
}

func InitMysqlDb() (*gorm.DB, error) {
	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s",
		viper.GetString("MYSQL_DB_ACCOUNT"),
		viper.GetString("MYSQL_DB_PASSWORD"),
		viper.GetString("MYSQL_DB_HOST"),
		viper.GetString("MYSQL_DB_PORT"),
		viper.GetString("MYSQL_DB_NAME"),
		viper.GetString("MYSQL_URL_SUFFIX"),
	)

	log.Info().Msgf("dsn: %s", dns)

	db, err := gorm.Open(mysql.Open(dns), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Minute * 30)

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
