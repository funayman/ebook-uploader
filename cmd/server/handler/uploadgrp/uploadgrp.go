// Package uploadgrp houses all handlers related to uploading
package uploadgrp

import (
	"context"
	"errors"
	"html/template"
	"net/http"

	"github.com/funayman/ebook-uploader/upload"
	"github.com/funayman/ebook-uploader/web"
)

const (
	defaultFormUploadID = "btn-upload"
)

var (
	ErrMissingFormField = errors.New("missing required form field")
)

type handler struct {
	uploadCore    *upload.Core
	maxUploadSize int64
	formUploadID  string
}

func newHandler(uploadCore *upload.Core, maxUploadSize int64) *handler {
	return &handler{
		uploadCore:    uploadCore,
		maxUploadSize: maxUploadSize,
		formUploadID:  defaultFormUploadID,
	}
}

func (h *handler) uploadForm(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	form := `
<html>
	<head>
		<title>Simple eBook Uploader</title>
	</head>
	<body>
		<form action="/upload" method="POST" enctype="multipart/form-data">
			<input
				id="{{ .InputID }}"
				name="{{ .InputID }}"
				type="file"
				accept=".prc,.cbr,.lit,.doc,.djvu,.opus,.html,.odt,.ogg,.cbz,.rtf,.mobi,.mp3,.wav,.m4b,.fb2,.epub,.azw3,.pdf,.mp4,.m4a,.azw,.docx,.kepub,.txt,.cbt,.flac"
				multiple />
			<button type="submit">SUBMIT!</button>
		</form>
	</body>
</html>
`
	data := struct {
		InputID string
	}{
		InputID: h.formUploadID,
	}

	t := template.Must(template.New("").Parse(form))
	return web.RespondHTMLTemplate(ctx, t, w, data, http.StatusOK)
}

func (h *handler) uploadFile(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		return err
	}
	defer r.MultipartForm.RemoveAll()

	if _, ok := r.MultipartForm.File[h.formUploadID]; !ok {
		return ErrMissingFormField
	}

	for _, mpf := range r.MultipartForm.File[h.formUploadID] {
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

	data := struct {
		Location string `json:"location"`
	}{
		Location: "/",
	}
	return web.RespondJSON(ctx, w, data, http.StatusOK)
}
