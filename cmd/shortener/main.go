package main

import (
	"github.com/gin-gonic/gin"

	"github.com/ruslantos/go-shortener-service/internal/config"
	"github.com/ruslantos/go-shortener-service/internal/handlers/getlink"
	"github.com/ruslantos/go-shortener-service/internal/handlers/postlink"
	"github.com/ruslantos/go-shortener-service/internal/storage"
)

func main() {
	l := storage.NewLinksStorage()
	r := gin.New()

	r.POST("/", postlink.New(l).Handle)
	r.GET("/:link", getlink.New(l).Handle)

	config.ParseFlags()
	err := r.Run(config.FlagServerPort)
	if err != nil {
		panic(err)
	}
}
