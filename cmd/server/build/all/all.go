// Package all defines ALL routes to be added to the web.App including such
// endpoints as crud, metrics, reporting, etc
package all

import (
	"github.com/funayman/ebook-uploader/cmd/server/handler/uploadgrp"
	"github.com/funayman/ebook-uploader/web"
	"github.com/funayman/ebook-uploader/web/mux"
)

func Routes() all {
	return all{}
}

type all struct{}

func (a all) Add(app *web.App, cfg mux.Config) {
	uploadgrp.Routes(app, uploadgrp.Config{
		Log:           cfg.Log,
		UploadCore:    cfg.UploadCore,
		MaxUploadSize: cfg.MaxUploadSize,
	})
}
