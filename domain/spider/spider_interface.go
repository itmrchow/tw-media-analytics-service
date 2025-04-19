package spider

import "context"

type Spider interface {
	GetNews(newsID string) (*NewsData, error)
	GetNewsList(newsIDList []string) ([]*NewsData, error)
	GetNewsIdList() ([]string, error)
	ArticleListScrapingHandle(ctx context.Context, msg []byte) error
	ArticleContentScrapingHandle(ctx context.Context, msg []byte) error
}
