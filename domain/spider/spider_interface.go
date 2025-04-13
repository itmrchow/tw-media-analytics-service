package spider

type Spider interface {
	GetNews(newsID int) (*NewsData, error)
}
