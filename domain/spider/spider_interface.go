package spider

type Spider interface {
	GetNews(newsID string) (*NewsData, error)
	GetNewsList(newsIDList []string) ([]*NewsData, error)
	GetNewsIdList() ([]string, error)
}
