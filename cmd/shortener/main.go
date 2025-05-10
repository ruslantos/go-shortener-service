package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileClient "github.com/ruslantos/go-shortener-service/internal/files"
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
	"github.com/ruslantos/go-shortener-service/internal/storage"
	"github.com/ruslantos/go-shortener-service/internal/storage/filestorage"
	"github.com/ruslantos/go-shortener-service/internal/storage/mapstorage"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	log, zapErr := zap.NewDevelopment()
	if zapErr != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	config.ParseFlags()

	var linkService service.LinkService
	var linkStorage service.LinksStorage

	cfg := storage.Load()
	switch cfg.StorageType {
	case "map":
		linkStorage = mapstorage.NewMapStorage()
	case "file":
		fileProducer, err := fileClient.NewProducer(config.FileStoragePath)
		if err != nil {
			logger.GetLogger().Fatal("cannot create file producer", zap.Error(err))
		}
		fileConsumer, err := fileClient.NewConsumer(config.FileStoragePath)
		if err != nil {
			logger.GetLogger().Fatal("cannot create file consumer", zap.Error(err))
		}

		linkStorage = filestorage.NewFileStorage(fileConsumer, fileProducer)
		err = linkStorage.InitStorage()
		if err != nil {
			logger.GetLogger().Fatal("cannot initialize file storage", zap.Error(err))
		}
	case "postgres":
		db, err := sqlx.Open("pgx", config.DatabaseDsn)
		if err != nil {
			logger.GetLogger().Fatal("cannot connect to database", zap.Error(err))
		}
		defer db.Close()

		linkStorage = storage.NewLinksStorage(db)
		err = linkStorage.InitStorage()
		if err != nil {
			logger.GetLogger().Fatal("cannot initialize database", zap.Error(err))
		}
	default:
		logger.GetLogger().Fatal("unknown storage type", zap.String("storageType", cfg.StorageType))
	}

	linkService = *service.NewLinkService(linkStorage)

	r := setupRouter(linkService, log)

	go linkService.StartDeleteWorker(context.Background())

	crt, key := config.GetCerts()
	fmt.Println(crt, key)

	var err error
	if config.EnableHTTPS {
		err = http.ListenAndServeTLS(":443", crt, key, r)
	} else {
		err = http.ListenAndServe(config.FlagServerPort, r)
	}
	if err != nil {
		logger.GetLogger().Fatal("cannot start server", zap.Error(err))
	}
}

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
	r.Mount("/debug/pprof", pprofHandler())

	return r
}

func pprofHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", pprof.Index)
	mux.HandleFunc("/cmdline", pprof.Cmdline)
	mux.HandleFunc("/profile", pprof.Profile)
	mux.HandleFunc("/symbol", pprof.Symbol)
	mux.HandleFunc("/trace", pprof.Trace)
	mux.Handle("/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/heap", pprof.Handler("heap"))
	mux.Handle("/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle("/block", pprof.Handler("block"))
	mux.Handle("/mutex", pprof.Handler("mutex"))
	return mux
}
