package main

import (
	"fmt"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/sony/sonyflake"
	"github.com/xiantank/url-shortener/config"
	db2 "github.com/xiantank/url-shortener/db"

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
	uniqueIDService := services.NewGlobalUniqueIDServiceBySonyFlake(sf)
	shortenerService := services.NewURLShorterService(uniqueIDService, redisCli, db, logger) // TODO: db use repository
	rest.RegisterHandler(r, shortenerService, logger)

	endless.ListenAndServe(":3000", r)

}
