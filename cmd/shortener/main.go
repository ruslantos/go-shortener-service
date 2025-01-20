package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	"github.com/ruslantos/go-shortener-service/internal/handlers/getlink"
	"github.com/ruslantos/go-shortener-service/internal/handlers/postlink"
	log "github.com/ruslantos/go-shortener-service/internal/logger"
	"github.com/ruslantos/go-shortener-service/internal/storage"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	l := storage.NewLinksStorage()
	r := gin.New()
	r.Use(log.LoggerMiddleware(logger))

	r.POST("/", postlink.New(l).Handle)
	r.GET("/:link", getlink.New(l).Handle)

	config.ParseFlags()
	err = r.Run(config.FlagServerPort)
	if err != nil {
		panic(err)
	}
}
