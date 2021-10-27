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
	"golang.org/x/sync/singleflight"

	"github.com/xiantank/url-shortener/errors"
	"github.com/xiantank/url-shortener/models"
	"github.com/xiantank/url-shortener/repositories"
)

const shortUrlCacheTemplate = "url::%s"
const cacheTTLInSeconds = int64(86400)

type UrlShorterService interface {
	Get(ctx context.Context, uid string) (string, error)
	Set(ctx context.Context, url string, expireAt int64) (*models.ShortUrl, error)
}

type urlShorterImpl struct {
	uniqueIDService    GlobalUniqueIDService
	bloomFilterService BloomFilterService
	sfg                *singleflight.Group

	cache  *redis.Client
	repoOp repositories.RepositoryOp
	logger *log.Logger
}

var _ UrlShorterService = (*urlShorterImpl)(nil)

func NewURLShorterService(sfg *singleflight.Group, uniqueIDService GlobalUniqueIDService, bloomFilterService BloomFilterService, redisCli *redis.Client, repoOp repositories.RepositoryOp, logger *log.Logger) UrlShorterService {
	rand.Seed(time.Now().UnixNano())
	_ = sonyflake.NewSonyflake(sonyflake.Settings{})
	return &urlShorterImpl{
		sfg:                sfg,
		uniqueIDService:    uniqueIDService,
		bloomFilterService: bloomFilterService,
		cache:              redisCli,
		repoOp:             repoOp,
		logger:             logger,
	}
}

func (u urlShorterImpl) Get(ctx context.Context, pathID string) (string, error) {
	cacheKey := fmt.Sprintf(shortUrlCacheTemplate, pathID)
	// NOTE: value in cache: empty string: expired; others: full url
	//TODO: not found but pass by bloom-filter also should put empty string to cache
	v, err := u.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		if v == "" {
			return "", errors.ErrExpired
		}

		return v, nil
	}

	if err != redis.Nil {
		return "", err
	}

	exists, err := u.bloomFilterService.Exists(pathID)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.ErrNotFound
	}

	resChan := u.sfg.DoChan(pathID, func() (interface{}, error) {
		ttl := u.getCacheTTL(cacheTTLInSeconds)
		shortUrl, err := u.repoOp.ShortUrl.GetByPathID(ctx, pathID)
		if err == errors.ErrNotFound {
			u.cache.SetNX(ctx, cacheKey, "", ttl)
		}
		if err != nil {
			return "", err
		}

		now := time.Now()
		expireAt := time.Unix(shortUrl.ExpireAt, 0)

		// expired
		if now.After(expireAt) {
			u.cache.SetNX(ctx, cacheKey, "", ttl)

			return "", errors.ErrExpired
		}

		// expired_at < now + ttl
		if now.Add(time.Duration(cacheTTLInSeconds) * time.Second).After(expireAt) {
			ttl = expireAt.Sub(now)
		}

		u.cache.SetNX(ctx, cacheKey, shortUrl.Url, ttl)

		return shortUrl.Url, err
	})

	select {
	case res := <-resChan:
		if res.Err != nil {
			return "", res.Err
		}

		url := res.Val.(string)
		return url, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (u urlShorterImpl) Set(ctx context.Context, url string, expireAt int64) (*models.ShortUrl, error) {
	uniqueId, err := u.uniqueIDService.GetID()
	if err != nil {
		return nil, err
	}

	pathID := hash(uniqueId + url)

	shortUrl := &models.ShortUrl{
		UID:      pathID,
		Url:      url,
		ExpireAt: expireAt,
	}

	if err := u.repoOp.ShortUrl.Create(ctx, shortUrl); err != nil {
		return nil, err
	}
	if err := u.bloomFilterService.Add(pathID); err != nil {
		return nil, err
	}

	return shortUrl, nil
}

// getCacheTTL: ttl 1~1.1 * timeInSeconds
func (u urlShorterImpl) getCacheTTL(timeInSeconds int64) time.Duration {
	return time.Duration(timeInSeconds)*time.Second + time.Second*time.Duration(float64(timeInSeconds)*0.1*rand.Float64())
}

func hash(s string) string {
	b := md5.Sum([]byte(s))
	return basen.Base62Encoding.EncodeToString(b[0:5])
}
