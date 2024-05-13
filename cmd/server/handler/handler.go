// Package handler binds mux/routes for endpoints in thie application
package handler

import (
	"net/http"
	"os"

	"go.uber.org/zap"

	"github.com/funayman/ebook-uploader/cmd/server/handler/uploadgrp"
	"github.com/funayman/ebook-uploader/upload"
	"github.com/funayman/ebook-uploader/web"
	"github.com/funayman/ebook-uploader/web/mid"
)

type Config struct {
	ShutdownCh    chan os.Signal
	CORSOrigins   []string
	Log           *zap.SugaredLogger
	UploadCore    *upload.Core
	MaxUploadSize int64
}

func Mux(config Config) http.Handler {
	globalMids := []web.Middleware{
		mid.Logger(config.Log),
		mid.Errors(config.Log),
		mid.Panics(),
	}

	app := web.NewApp(
		config.ShutdownCh,
		web.NewCorsOptions(config.CORSOrigins...),
		globalMids...,
	)

	uploadgrp.Bind(app, uploadgrp.Config{
		Log:           config.Log,
		UploadCore:    config.UploadCore,
		MaxUploadSize: config.MaxUploadSize,
	})

	return app
}
