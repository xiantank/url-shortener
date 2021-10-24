package models

type ShortUrl struct {
	ID       uint32
	UID      string
	Url      string
	ExpireAt int64
}
