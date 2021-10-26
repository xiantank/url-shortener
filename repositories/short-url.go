package repositories

import (
	"context"

	"github.com/afex/hystrix-go/hystrix"
	"gorm.io/gorm"

	"github.com/xiantank/url-shortener/models"
)

const dbAccess = "db-access"

type ShortUrlRepo interface {
	GetByPathID(ctx context.Context, pathID string) (*models.ShortUrl, error)
	Create(ctx context.Context, shortUrl *models.ShortUrl) error
}

type shortUrlImpl struct {
	db *gorm.DB
}

func NewShortUrlRepo(db *gorm.DB) ShortUrlRepo {
	return &shortUrlImpl{db: db}
}

func (s shortUrlImpl) GetByPathID(ctx context.Context, pathID string) (*models.ShortUrl, error) {
	shortUrl := &models.ShortUrl{}
	err := hystrix.Do(dbAccess, func() error {
		db := s.db.WithContext(ctx)

		// TODO: different gorm.ErrRecordNotFound and other error, or will trigger circuit breaker
		return db.Where("uid = ?", pathID).Take(&shortUrl).Error
	}, nil)

	if err != nil {
		return nil, err
	}

	return shortUrl, nil
}

func (s shortUrlImpl) Create(ctx context.Context, shortUrl *models.ShortUrl) error {
	return hystrix.Do(dbAccess, func() error {
		db := s.db.WithContext(ctx)

		return db.Create(shortUrl).Error
	}, nil)
}
