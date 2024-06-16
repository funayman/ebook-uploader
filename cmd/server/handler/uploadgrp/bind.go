package uploadgrp

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/funayman/ebook-uploader/upload"
	"github.com/funayman/ebook-uploader/web"
	"github.com/funayman/ebook-uploader/web/mid"
)

type Config struct {
	Log           *zap.SugaredLogger
	UploadCore    *upload.Core
	MaxUploadSize int64
}

func Bind(app *web.App, config Config) {
	h := newHandler(config.UploadCore, config.MaxUploadSize)

	app.Handle(http.MethodGet, "/upload", h.uploadForm)
	app.Handle(http.MethodPost, "/upload", h.uploadFile, mid.LimitBodySize(config.MaxUploadSize))
	app.Handle(http.MethodGet, "/upload/complete", h.uploadSuccessError)
}
