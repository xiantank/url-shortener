package services

import (
	"context"
	"crypto/md5"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nicksnyder/basen"
	log "github.com/sirupsen/logrus"
	"github.com/sony/sonyflake"
	"github.com/xiantank/url-shortener/services/models"
	"gorm.io/gorm"
)

const shortUrlCacheTemplate = "url::%s"
const defaultCacheTTL = time.Hour * 24

type UrlShorterService interface {
	Get(ctx context.Context, uid string) (string, error)
	Set(ctx context.Context, url string, expireAt int64) (*models.ShortUrl, error)
}

type urlShorterImpl struct {
	uniqueIDService GlobalUniqueIDService

	cache  *redis.Client
	db     *gorm.DB
	logger *log.Logger
}

var _ UrlShorterService = (*urlShorterImpl)(nil)

func NewURLShorterService(uniqueIDService GlobalUniqueIDService, redisCli *redis.Client, db *gorm.DB, logger *log.Logger) UrlShorterService {
	rand.Seed(time.Now().UnixNano())
	_ = sonyflake.NewSonyflake(sonyflake.Settings{})
	return &urlShorterImpl{
		uniqueIDService: uniqueIDService,
		cache:           redisCli,
		db:              db,
		logger:          logger,
	}
}

func (u urlShorterImpl) Get(ctx context.Context, pathID string) (string, error) {
	cacheKey := fmt.Sprintf(shortUrlCacheTemplate, pathID)
	v, err := u.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		return v, nil
	}

	if err != redis.Nil {
		return "", err
	}

	// TODO: use bloom(or xor/ribbon) filter to check should find in db
	// TODO: limit concurrent find in db
	// TODO: need to handle high concurrent request cache issue()

	shortUrl := &models.ShortUrl{}
	db := u.db.WithContext(ctx)

	err = db.Where("uid = ?", pathID).Find(shortUrl).Error
	return shortUrl.Url, err
}

func (u urlShorterImpl) Set(ctx context.Context, url string, expireAt int64) (*models.ShortUrl, error) {
	uniqueId, err := u.uniqueIDService.GetID()
	if err != nil {
		return nil, err
	}

	pathID := hash(uniqueId + url) // TODO: padWithZero or not?

	// TODO: if pathID MAYBE exists (use bloom filter to check) generate new one(maybe at most 3 times?)

	shortUrl := &models.ShortUrl{
		UID:      pathID,
		Url:      url,
		ExpireAt: expireAt,
	}
	// TODO: handle expired in db(maybe rm data or handle ttl in app/cache)
	result := u.db.Create(shortUrl)
	if result.Error != nil {
		return nil, result.Error
	}

	// TODO: update to bloom filter
	cacheKey := fmt.Sprintf(shortUrlCacheTemplate, pathID)

	// TODO: update to cache with default + random ttl
	u.cache.SetNX(ctx, cacheKey, url, defaultCacheTTL)

	return shortUrl, nil
}

func hash(s string) string {
	b := md5.Sum([]byte(s))
	return basen.Base62Encoding.EncodeToString(b[0:5])
}
