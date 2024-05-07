package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type validator interface {
	Validate() error
}

func URLParam(r *http.Request, param string) string {
	return chi.URLParam(r, param)
}

func DecodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}

	if v, ok := dst.(validator); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("unable to validate payload: %w", err)
		}
	}

	return nil
}
