package repository

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"itmrchow/tw-media-analytics-service/domain/news/entity"
)

var _ AuthorRepository = &AuthorRepositoryImpl{}

type AuthorRepositoryImpl struct {
	log *zerolog.Logger
	db  *gorm.DB
}

func NewAuthorRepositoryImpl(log *zerolog.Logger, db *gorm.DB) *AuthorRepositoryImpl {
	return &AuthorRepositoryImpl{
		log: log,
		db:  db,
	}
}

func (r *AuthorRepositoryImpl) WithTransaction(tx *gorm.DB) AuthorRepository {
	r.db = tx
	return r
}

func (r *AuthorRepositoryImpl) FirstOrCreate(author *entity.Author) error {
	return r.db.FirstOrCreate(author, author).Error
}
