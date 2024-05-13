package uploadmulti

import (
	"errors"
	"io"
)

type multiWriteCloser struct {
	w  io.Writer
	cs []io.Closer
}

func MultiWriteCloser(writeClosers ...io.WriteCloser) io.WriteCloser {
	m := &multiWriteCloser{
		cs: make([]io.Closer, len(writeClosers)),
	}
	writers := make([]io.Writer, len(writeClosers))

	for i, w := range writeClosers {
		m.cs[i] = w
		writers[i] = w
	}
	m.w = io.MultiWriter(writers...)

	return m
}

func (m *multiWriteCloser) Write(p []byte) (n int, err error) {
	return m.w.Write(p)
}

func (m *multiWriteCloser) Close() error {
	errs := make([]error, 0, len(m.cs))

	for _, c := range m.cs {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
