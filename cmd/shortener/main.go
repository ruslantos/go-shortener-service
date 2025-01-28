package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileClient "github.com/ruslantos/go-shortener-service/internal/file"
	"github.com/ruslantos/go-shortener-service/internal/handlers/getlink"
	"github.com/ruslantos/go-shortener-service/internal/handlers/postlink"
	"github.com/ruslantos/go-shortener-service/internal/handlers/shorten"
	"github.com/ruslantos/go-shortener-service/internal/middleware"
	"github.com/ruslantos/go-shortener-service/internal/middleware/compress"
	"github.com/ruslantos/go-shortener-service/internal/storage"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()
	config.ParseFlags()

	fileProducer, err := fileClient.NewProducer(config.FileStoragePath + "links")
	fileConsumer, err := fileClient.NewConsumer(config.FileStoragePath + "links")

	l := storage.NewLinksStorage(fileConsumer)
	err = l.InitLinkMap()
	if err != nil {
		panic(err)
	}
	r := gin.New()
	//r.Use(middleware.Logger(logger), middleware.Gzip())
	//r.Use(middleware.Gzip())

	placeholderHandler := func(w http.ResponseWriter, r *http.Request) {}
	wrappedHandler := compress.GzipMiddleware(placeholderHandler)

	r.Use(middleware.Logger(logger), wrapHTTPHandlerFunc(wrappedHandler))

	r.POST("/", postlink.New(l, fileProducer).Handle)
	r.POST("/api/shorten", shorten.New(l, fileProducer).Handle)
	r.GET("/:link", getlink.New(l).Handle)

	err = r.Run(config.FlagServerPort)
	if err != nil {
		panic(err)
	}
}

func wrapHTTPHandlerFunc(h http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
