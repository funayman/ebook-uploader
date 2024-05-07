package mid

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/funayman/simple-file-uploader/web"
)

var (
	ErrBasicUnauthorized = errors.New("Unauthorized")
)

func BasicAuth(username, password string) web.Middleware {
	expectedUsernameHash := sha256.Sum256([]byte(username))
	expectedPasswordHash := sha256.Sum256([]byte(password))

	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			baUsername, baPassword, ok := r.BasicAuth()
			if ok {
				baUsernameHash := sha256.Sum256([]byte(baUsername))
				baPasswordHash := sha256.Sum256([]byte(baPassword))

				usernameMatch := (subtle.ConstantTimeCompare(baUsernameHash[:], expectedUsernameHash[:]) == 1)
				passwordMatch := (subtle.ConstantTimeCompare(baPasswordHash[:], expectedPasswordHash[:]) == 1)

				if usernameMatch && passwordMatch {
					return next(ctx, w, r)
				}
			}

			return ErrBasicUnauthorized
		}
	}
}
