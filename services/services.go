package services

import (
	redisbloom "github.com/RedisBloom/redisbloom-go"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/sony/sonyflake"
	"golang.org/x/sync/singleflight"

	"github.com/xiantank/url-shortener/config"
	"github.com/xiantank/url-shortener/repositories"
)

type ServiceOp struct {
	uniqueIDService GlobalUniqueIDService
	UrlShorter      UrlShorterService
}

func New(repoOp repositories.RepositoryOp, sf *sonyflake.Sonyflake, sfg *singleflight.Group, redisCli *redis.Client, redisBloomCli *redisbloom.Client, logger *logrus.Logger) ServiceOp {
	uniqueIDService := NewGlobalUniqueIDServiceBySonyFlake(sf)
	filterService := NewBloomFilterService(redisBloomCli, config.BloomFilterName)
	return ServiceOp{
		uniqueIDService: uniqueIDService,
		UrlShorter:      NewURLShorterService(sfg, uniqueIDService, filterService, redisCli, repoOp, logger),
	}
}
