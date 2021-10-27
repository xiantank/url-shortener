package services

import redisbloom "github.com/RedisBloom/redisbloom-go"

type BloomFilterService interface {
	Add(item string) error
	Exists(item string) (bool, error)
	//Dump()
	//Restore()
}

type bloomFilterServiceImpl struct {
	redisCli   *redisbloom.Client
	filterName string
}

func (b bloomFilterServiceImpl) Add(item string) error {
	_, err := b.redisCli.Add(b.filterName, item)
	return err
}

func (b bloomFilterServiceImpl) Exists(item string) (bool, error) {
	return b.redisCli.Exists(b.filterName, item)
}

func NewBloomFilterService(client *redisbloom.Client, filterName string) BloomFilterService {
	return &bloomFilterServiceImpl{
		redisCli:   client,
		filterName: filterName,
	}
}
