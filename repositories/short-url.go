package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/xiantank/url-shortener/errors"
	"github.com/xiantank/url-shortener/models"
)

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
	db := s.db.WithContext(ctx)

	err := db.Where("uid = ?", pathID).Take(&shortUrl).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.ErrNotFound
	}

	return shortUrl, nil
}

func (s shortUrlImpl) Create(ctx context.Context, shortUrl *models.ShortUrl) error {
	db := s.db.WithContext(ctx)

	return db.Create(shortUrl).Error
}
