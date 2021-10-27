package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/xiantank/url-shortener/errors"
	"github.com/xiantank/url-shortener/services"
)

type restImpl struct {
	engine          *gin.Engine
	serviceOp       services.ServiceOp
	uniqueIDService services.GlobalUniqueIDService
	logger          *logrus.Logger
}

type PostShortUrlRequest struct {
	Url      string    `json:"url" binding:"required,url"`
	ExpireAt time.Time `json:"expireAt" binding:"required"`
}

func RegisterHandler(engine *gin.Engine, serviceOp services.ServiceOp, logger *logrus.Logger) {
	// TODO: add log/error handling middleware
	ri := &restImpl{
		engine:    engine,
		serviceOp: serviceOp,
		logger:    logger,
	}

	engine.GET("/:uid", ri.GetShorts)

	api := engine.Group("/api/v1")

	api.GET("/", ri.ok)
	api.POST("/urls", ri.PostShorts)
}

func (ri restImpl) ok(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "ok")
}

func (ri restImpl) GetShorts(ctx *gin.Context) {
	uid := ctx.Param("uid")
	url, err := ri.serviceOp.UrlShorter.Get(ctx.Request.Context(), uid)
	if err != nil {
		switch err {
		case errors.ErrExpired:
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		case errors.ErrNotFound:
			ri.logger.Infof("not exists pathID: %v", uid)
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, url)
}

func (ri restImpl) PostShorts(ctx *gin.Context) {
	// TODO: add request validation
	req := PostShortUrlRequest{}
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	shortUrl, err := ri.serviceOp.UrlShorter.Set(ctx.Request.Context(), req.Url, req.ExpireAt.Unix())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":       shortUrl.UID,
		"shortUrl": fmt.Sprintf("http://%s/%s", ctx.Request.Host, shortUrl.UID), // ignore determined http or https
	})
}
