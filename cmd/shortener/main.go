package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/acme/autocert"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	"github.com/ruslantos/go-shortener-service/internal/handlers/deleteuserurls"
	"github.com/ruslantos/go-shortener-service/internal/handlers/getlink"
	"github.com/ruslantos/go-shortener-service/internal/handlers/getstats"
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

	cfg := config.ParseFlags()

	linkStorage := storage.Get(cfg)
	defer linkStorage.Close()

	linkService := *service.NewLinkService(linkStorage)

	r := setupRouter(linkService, log)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	go linkService.StartDeleteWorker(ctx)

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	go func() {
		var err error
		if cfg.EnableHTTPS {
			certManager := &autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist("shortener.com"),
				Cache:      autocert.DirCache("certs"),
			}
			tlsConfig := &tls.Config{
				GetCertificate: certManager.GetCertificate,
				MinVersion:     tls.VersionTLS12,
			}
			srv.Addr = ":443"
			srv.TLSConfig = tlsConfig

			logger.GetLogger().Info("Starting HTTPS server on :443")
			err = srv.ListenAndServeTLS("", "")
		} else {
			logger.GetLogger().Info("Starting HTTP server on :80")
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			logger.GetLogger().Fatal("cannot start server", zap.Error(err))
		}
	}()

	<-ctx.Done()

	logger.GetLogger().Info("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.GetLogger().Error("Server forced to shutdown", zap.Error(err))
	}

	logger.GetLogger().Info("Server exited properly")
}

func setupRouter(linkService service.LinkService, log *zap.Logger) *chi.Mux {
	postLinkHandler := postlink.New(&linkService)
	getLinkHandler := getlink.New(&linkService)
	shortenHandler := shorten.New(&linkService)
	pingHandler := ping.New(&linkService)
	shortenBatchHandler := shortenbatch.New(&linkService)
	getUserUrlsHandler := getuserurls.New(&linkService)
	deleteUserUrlsHandler := deleteuserurls.New(&linkService)
	getStatsHandler := getstats.New(&linkService)

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
	r.Get("/api/internal/stats", getStatsHandler.Handle)

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
