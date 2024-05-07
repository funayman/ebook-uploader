// Package uploadfs is an upload file store to a local file system
package uploadfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/inhies/go-bytesize"
	"go.uber.org/zap"
)

var (
	ErrInvalidDirectory = errors.New("invalid directory")
	ErrNotWritable      = errors.New("cannot write to directory")
)

type Store struct {
	log *zap.SugaredLogger
	dir string
}

func NewStore(log *zap.SugaredLogger, dir string) (*Store, error) {
	fi, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("os.Stat: %w", err)
	}
	if !fi.IsDir() {
		return nil, ErrInvalidDirectory
	}
	// check write bit for user
	// https: //stackoverflow.com/a/49148866
	if fi.Mode().Perm()&(1<<(uint(7))) == 0 {
		return nil, ErrNotWritable
	}

	return &Store{log: log, dir: dir}, nil
}

// Save copies the source reader contents to a new file on the system using the
// directory within the Store and the name provided in the function as the full
// path. The source file is closed upon return
func (s *Store) Save(name string, src io.ReadCloser) error {
	// defer src.Close()

	fn := filepath.Join(s.dir, name)

	dst, err := os.OpenFile(fn, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	// used for debugging and time taken to copy to disk
	t := time.Now()

	n, err := io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	s.log.Infow("copied file to disk", "bytes", bytesize.ByteSize(n).String(), "filename", fn, "since", time.Since(t))

	return nil
}
