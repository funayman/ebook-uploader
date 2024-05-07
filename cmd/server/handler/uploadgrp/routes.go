package uploadgrp

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/funayman/simple-file-uploader/upload"
	"github.com/funayman/simple-file-uploader/web"
	"github.com/funayman/simple-file-uploader/web/mid"
)

type Config struct {
	Log           *zap.SugaredLogger
	UploadCore    *upload.Core
	MaxUploadSize int64
}

func Routes(app *web.App, cfg Config) {
	h := newHandler(cfg.UploadCore, cfg.MaxUploadSize)

	app.Handle(http.MethodGet, "/upload", h.uploadForm)
	app.Handle(http.MethodPost, "/upload", h.uploadFile, mid.LimitBodySize(cfg.MaxUploadSize))
}
