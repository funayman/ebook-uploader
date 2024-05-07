// Package mux sets the mux def
package mux

import (
	"net/http"
	"os"

	"go.uber.org/zap"

	"github.com/funayman/simple-file-uploader/upload"
	"github.com/funayman/simple-file-uploader/web"
	"github.com/funayman/simple-file-uploader/web/mid"
)

type Config struct {
	Build         string
	ShutdownCh    chan os.Signal
	Log           *zap.SugaredLogger
	UploadCore    *upload.Core
	MaxUploadSize int64
}

// RouteAdder defines behavior that sets routes to bind to an instance
type RouteAdder interface {
	Add(*web.App, Config)
}

func Web(cfg Config, corsOrigins []string, routeAdder RouteAdder) http.Handler {
	globalMids := []web.Middleware{
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Panics(),
	}

	app := web.NewApp(cfg.ShutdownCh, web.NewCorsOptions(corsOrigins...), globalMids...)

	routeAdder.Add(app, cfg)

	return app
}
