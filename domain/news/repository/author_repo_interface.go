package repository

import "itmrchow/tw-media-analytics-service/domain/news/entity"

type AuthorRepository interface {
	BaseRepository[AuthorRepository]
	FirstOrCreate(author *entity.Author) error
}
