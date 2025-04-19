package repository

import "gorm.io/gorm"

type BaseRepository[T any] interface {
	WithTransaction(tx *gorm.DB) T
}
