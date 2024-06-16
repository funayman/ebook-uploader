package web

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

var (
	ErrBadTemplate = errors.New("bad template")
)

func RespondJSON(ctx context.Context, w http.ResponseWriter, data any, statusCode int) error {
	setStatusCode(ctx, statusCode)

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}

func RespondText(ctx context.Context, w http.ResponseWriter, data string, statusCode int) error {
	setStatusCode(ctx, statusCode)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	if _, err := w.Write([]byte(data)); err != nil {
		return err
	}

	return nil
}

func RespondHTMLTemplate(ctx context.Context, t *template.Template, w http.ResponseWriter, data any, statusCode int) error {
	if t == nil {
		return ErrBadTemplate
	}

	setStatusCode(ctx, statusCode)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	pr, pw := io.Pipe()
	go func(pw *io.PipeWriter) {
		err := t.Execute(pw, data)
		pw.CloseWithError(err)
	}(pw)

	m := minify.New()
	m.Add("text/html", &html.Minifier{
		KeepDefaultAttrVals: true,
		KeepEndTags:         true,
		KeepDocumentTags:    true,
	})
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	if err := m.Minify("text/html", w, pr); err != nil {
		return err
	}

	return nil
}
