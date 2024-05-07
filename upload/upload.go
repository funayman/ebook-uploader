// Package upload creates a new core for storing uploaded files
package upload

import (
	"fmt"
	"io"

	"go.uber.org/zap"
)

type Storer interface {
	Save(string, io.ReadCloser) error
}

type Core struct {
	log    *zap.SugaredLogger
	storer Storer
}

func NewCore(log *zap.SugaredLogger, storer Storer) *Core {
	return &Core{
		log:    log,
		storer: storer,
	}
}

func (c *Core) Save(name string, src io.ReadCloser) error {
	if err := c.storer.Save(name, src); err != nil {
		return fmt.Errorf("storer: %w", err)
	}

	return nil
}
