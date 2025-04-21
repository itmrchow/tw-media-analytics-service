package repository

import (
	"context"

	"itmrchow/tw-media-analytics-service/domain/news/entity"
)

type AuthorRepository interface {
	BaseRepository[AuthorRepository]
	FirstOrCreate(ctx context.Context, author *entity.Author) error
}
