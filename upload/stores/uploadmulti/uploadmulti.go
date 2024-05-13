// Package uploadmulti allows for sending the source to multiple Stores
package uploadmulti

import (
	"errors"
	"io"

	"go.uber.org/zap"

	"github.com/funayman/ebook-uploader/upload"
)

var (
	ErrNoStores = errors.New("no stores provided")
)

type Store struct {
	log    *zap.SugaredLogger
	stores []upload.Storer
}

func NewStore(log *zap.SugaredLogger, stores ...upload.Storer) (*Store, error) {
	if len(stores) == 0 {
		return nil, ErrNoStores
	}
	return &Store{log: log, stores: stores}, nil
}

// Save copies the source reader contents to a new file in the bucket defined
// within the Store and the name provided in the function as the full path
func (s *Store) Save(name string, src io.ReadCloser) error {
	n := len(s.stores)

	wcs := make([]io.WriteCloser, n)
	results := make(chan error)
	errs := make([]error, 0, n+1)
	defer close(results)

	for i := 0; i < len(s.stores); i++ {
		store := s.stores[i]
		pr, pw := io.Pipe()
		wcs[i] = pw

		go func(store upload.Storer) {
			results <- store.Save(name, pr)
		}(store)
	}

	go func(wc io.WriteCloser) {
		defer wc.Close()
		_, err := io.Copy(wc, src)
		results <- err
	}(MultiWriteCloser(wcs...))

	for i := 0; i < n+1; i++ {
		if err := <-results; err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
