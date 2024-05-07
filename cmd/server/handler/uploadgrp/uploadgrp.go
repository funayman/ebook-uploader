// Package uploadgrp houses all handlers related to uploading
package uploadgrp

import (
	"context"
	"errors"
	"html/template"
	"net/http"

	"github.com/funayman/simple-file-uploader/upload"
	"github.com/funayman/simple-file-uploader/web"
)

var (
	ErrMissingFormField = errors.New("missing required form field")
)

type handler struct {
	uploadCore    *upload.Core
	maxUploadSize int64
}

func newHandler(uploadCore *upload.Core, maxUploadSize int64) *handler {
	return &handler{
		uploadCore:    uploadCore,
		maxUploadSize: maxUploadSize,
	}
}

func (h *handler) uploadForm(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	form := `<html>
<body>
<form action="/upload" method="POST" enctype="multipart/form-data">
<input type="file" name="btn-upload" id="btn-upload" multiple/>
<button type="submit">SUBMIT!</button>
</form>
</body>
</html>
`
	t := template.Must(template.New("").Parse(form))
	return web.RespondHTMLTemplate(ctx, t, w, nil, 200)
}

func (h *handler) uploadFile(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		return err
	}
	defer r.MultipartForm.RemoveAll()

	if _, ok := r.MultipartForm.File["btn-upload"]; !ok {
		return ErrMissingFormField
	}

	for _, mpf := range r.MultipartForm.File["btn-upload"] {
		err := func() error {
			src, err := mpf.Open()
			if err != nil {
				return err
			}
			defer src.Close()

			return h.uploadCore.Save(mpf.Filename, src)
		}()

		if err != nil {
			return err
		}
	}

	return web.RespondJSON(ctx, w, "upload complete", 200)
}

// func (h *handler) doUpload(r *http.Request) error {
// 	var isUploadFieldPresent bool
//
// 	mpr, err := r.MultipartReader()
// 	if err != nil {
// 		return err
// 	}
//
// 	var part *multipart.Part
// 	for part, err = mpr.NextPart(); err == nil; part, err = mpr.NextPart() {
// 		if part.FormName() != "btn-upload" {
// 			continue
// 		}
//
// 		isUploadFieldPresent = true
// 		if err := h.uploadCore.Save(part.FileName(), part); err != nil {
// 			fmt.Errorf("upload core: %w", err)
// 		}
// 	}
//
// 	if !errors.Is(err, io.EOF) {
// 		return fmt.Errorf("doUpload: %w", err)
// 	}
//
// 	if !isUploadFieldPresent {
// 		return ErrMissingFormField
// 	}
//
// 	return nil
// }
