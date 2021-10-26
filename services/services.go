package services

import (
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/sony/sonyflake"
	"golang.org/x/sync/singleflight"

	"github.com/xiantank/url-shortener/repositories"
)

type ServiceOp struct {
	uniqueIDService GlobalUniqueIDService
	UrlShorter      UrlShorterService
}

func New(repoOp repositories.RepositoryOp, sf *sonyflake.Sonyflake, sfg *singleflight.Group, redisCli *redis.Client, logger *logrus.Logger) ServiceOp {
	uniqueIDService := NewGlobalUniqueIDServiceBySonyFlake(sf)
	return ServiceOp{
		uniqueIDService: uniqueIDService,
		UrlShorter:      NewURLShorterService(sfg, uniqueIDService, redisCli, repoOp, logger),
	}
}
