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

	"github.com/xiantank/url-shortener/models"
	"github.com/xiantank/url-shortener/repositories"
)

const shortUrlCacheTemplate = "url::%s"
const defaultCacheTTL = time.Hour * 24

type UrlShorterService interface {
	Get(ctx context.Context, uid string) (string, error)
	Set(ctx context.Context, url string, expireAt int64) (*models.ShortUrl, error)
}

type urlShorterImpl struct {
	uniqueIDService GlobalUniqueIDService
	sfg             *singleflight.Group

	cache  *redis.Client
	repoOp repositories.RepositoryOp
	logger *log.Logger
}

var _ UrlShorterService = (*urlShorterImpl)(nil)

func NewURLShorterService(sfg *singleflight.Group, uniqueIDService GlobalUniqueIDService, redisCli *redis.Client, repoOp repositories.RepositoryOp, logger *log.Logger) UrlShorterService {
	rand.Seed(time.Now().UnixNano())
	_ = sonyflake.NewSonyflake(sonyflake.Settings{})
	return &urlShorterImpl{
		sfg:             sfg,
		uniqueIDService: uniqueIDService,
		cache:           redisCli,
		repoOp:          repoOp,
		logger:          logger,
	}
}

func (u urlShorterImpl) Get(ctx context.Context, pathID string) (string, error) {
	cacheKey := fmt.Sprintf(shortUrlCacheTemplate, pathID)
	// NOTE: value in cache: empty string: expired; others: full url
	//TODO: not found but pass by bloom-filter also should put empty string to cache
	v, err := u.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		if v == "" {
			return "", ExpiredError
		}

		return v, nil
	}

	if err != redis.Nil {
		return "", err
	}

	// TODO: use bloom(or xor/ribbon) filter to check should find in db
	// TODO: need to handle high concurrent request cache issue()

	resChan := u.sfg.DoChan(pathID, func() (interface{}, error) {
		shortUrl, err := u.repoOp.ShortUrl.GetByPathID(ctx, pathID) // TODO: handle gorm.ErrRecordNotFound
		if err != nil {
			return "", err
		}

		now := time.Now()
		expireAt := time.Unix(shortUrl.ExpireAt, 0)
		ttl := defaultCacheTTL

		// expired
		if now.After(expireAt) {
			u.cache.SetNX(ctx, cacheKey, "", ttl)

			return "", ExpiredError
		}

		// expired_at < now + ttl
		if now.Add(defaultCacheTTL).After(expireAt) {
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

	pathID := hash(uniqueId + url) // TODO: padWithZero or not?

	// TODO: if pathID MAYBE exists (use bloom filter to check) generate new one(maybe at most 3 times?)

	shortUrl := &models.ShortUrl{
		UID:      pathID,
		Url:      url,
		ExpireAt: expireAt,
	}
	// TODO: handle expired in db(maybe rm data or handle ttl in app/cache)

	if err := u.repoOp.ShortUrl.Create(ctx, shortUrl); err != nil {
		return nil, err
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
