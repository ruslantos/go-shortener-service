package main

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/handlers/deleteuserurls"
	"github.com/ruslantos/go-shortener-service/internal/handlers/getlink"
	"github.com/ruslantos/go-shortener-service/internal/handlers/getuserurls"
	"github.com/ruslantos/go-shortener-service/internal/handlers/ping"
	"github.com/ruslantos/go-shortener-service/internal/handlers/postlink"
	"github.com/ruslantos/go-shortener-service/internal/handlers/shorten"
	"github.com/ruslantos/go-shortener-service/internal/handlers/shortenbatch"
	authMiddlware "github.com/ruslantos/go-shortener-service/internal/middleware/auth"
	"github.com/ruslantos/go-shortener-service/internal/middleware/compress"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/service"
)

func setupRouter(linkService service.LinkService, log *zap.Logger) *chi.Mux {
	postLinkHandler := postlink.New(&linkService)
	getLinkHandler := getlink.New(&linkService)
	shortenHandler := shorten.New(&linkService)
	pingHandler := ping.New(&linkService)
	shortenBatchHandler := shortenbatch.New(&linkService)
	getUserUrlsHandler := getuserurls.New(&linkService)
	deleteUserUrlsHandler := deleteuserurls.New(&linkService)

	r := chi.NewRouter()

	r.Use(compress.GzipMiddlewareWriter,
		compress.GzipMiddlewareReader,
		logger.LoggerChi(log),
		authMiddlware.CookieMiddleware)

	r.Post("/", postLinkHandler.Handle)
	r.Get("/{link}", getLinkHandler.Handle)
	r.Post("/api/shorten", shortenHandler.Handle)
	r.Get("/ping", pingHandler.Handle)
	r.Post("/api/shorten/batch", shortenBatchHandler.Handle)
	r.Get("/api/user/urls", getUserUrlsHandler.Handle)
	r.Delete("/api/user/urls", deleteUserUrlsHandler.Handle)

	return r
}
