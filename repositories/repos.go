package repositories

import "gorm.io/gorm"

type RepositoryOp struct {
	ShortUrl ShortUrlRepo
}

func New(db *gorm.DB) RepositoryOp {
	return RepositoryOp{
		ShortUrl: NewShortUrlRepo(db),
	}
}
