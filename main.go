package main

import (
	"fmt"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/sony/sonyflake"
	"golang.org/x/sync/singleflight"

	"github.com/xiantank/url-shortener/config"
	db2 "github.com/xiantank/url-shortener/db"
	"github.com/xiantank/url-shortener/repositories"
	"github.com/xiantank/url-shortener/rest"
	"github.com/xiantank/url-shortener/services"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	db, err := db2.New() // TODO: use noSQL is better in this system
	if err != nil {
		panic(err)
	}
	redisCli := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
		Password: "", // no password set
		DB:       0,  // use default DB
	}) // TODO: mv and check already connect

	//gin.SetMode(gin.ReleaseMode)
	db = db.Debug()

	r := gin.Default()
	sf := sonyflake.NewSonyflake(sonyflake.Settings{})
	sfg := &singleflight.Group{}
	repoOp := repositories.New(db)
	serviceOp := services.New(repoOp, sf, sfg, redisCli, logger)
	rest.RegisterHandler(r, serviceOp, logger)

	endless.ListenAndServe(":3000", r)

}
