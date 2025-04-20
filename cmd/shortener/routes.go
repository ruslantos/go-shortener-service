package main

//func setupRouter(linkService service.LinkService, log *zap.Logger) *chi.Mux {
//	postLinkHandler := postlink.New(&linkService)
//	getLinkHandler := getlink.New(&linkService)
//	shortenHandler := shorten.New(&linkService)
//	pingHandler := ping.New(&linkService)
//	shortenBatchHandler := shortenbatch.New(&linkService)
//	getUserUrlsHandler := getuserurls.New(&linkService)
//	deleteUserUrlsHandler := deleteuserurls.New(&linkService)
//
//	r := chi.NewRouter()
//
//	r.Use(compress.GzipMiddlewareWriter,
//		compress.GzipMiddlewareReader,
//		logger.LoggerChi(log),
//		authMiddlware.CookieMiddleware)
//
//	r.Post("/", postLinkHandler.Handle)
//	r.Get("/{link}", getLinkHandler.Handle)
//	r.Post("/api/shorten", shortenHandler.Handle)
//	r.Get("/ping", pingHandler.Handle)
//	r.Post("/api/shorten/batch", shortenBatchHandler.Handle)
//	r.Get("/api/user/urls", getUserUrlsHandler.Handle)
//	r.Delete("/api/user/urls", deleteUserUrlsHandler.Handle)
//
//	return r
//}
