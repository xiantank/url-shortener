package services

import (
	"strconv"

	"github.com/sony/sonyflake"
)

type GlobalUniqueIDService interface {
	GetID() (string, error)
}

type globalUniqueIDServiceSonyFlakeImpl struct {
	sf *sonyflake.Sonyflake
}

func NewGlobalUniqueIDServiceBySonyFlake(sf *sonyflake.Sonyflake) GlobalUniqueIDService {
	return &globalUniqueIDServiceSonyFlakeImpl{
		sf: sf,
	}
}

func (g globalUniqueIDServiceSonyFlakeImpl) GetID() (string, error) {
	id, err := g.sf.NextID()
	if err != nil {
		return "", err
	}

	return strconv.FormatUint(id, 10), nil
}
