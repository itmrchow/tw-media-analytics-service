package spider

import "context"

type Spider interface {
	GetNews(newsID string) (*NewsData, error)
	GetNewsList(newsIDList []string) ([]*NewsData, error)
	GetNewsIdList() ([]string, error)
	ArticleScrapingHandle(ctx context.Context, msg []byte) error
}

type GetNewsEvent struct {
}
