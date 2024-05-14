package uploadmulti

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/funayman/logger"
	"go.uber.org/zap"
)

var (
	log *zap.SugaredLogger
)

func init() {
	log, _ = logger.New("TESTS", logger.WithLevel("FATAL"))
}

type testUploader struct {
	dir string
}

func (tu testUploader) Save(name string, src io.ReadCloser) error {
	f, err := os.CreateTemp(tu.dir, fmt.Sprintf("*%s", name))
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, src); err != nil {
		return err
	}

	return nil
}

func TestMultiStore(t *testing.T) {
	dir, err := os.MkdirTemp("", "uploadmultiTest")
	if err != nil {
		t.Fatalf("os.MkdirTemp: %v", err)
	}

	t.Logf("dir: %s", dir)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	tu1 := testUploader{dir: dir}
	tu2 := testUploader{dir: dir}
	tu3 := testUploader{dir: dir}

	s, err := NewStore(log, tu1, tu2, tu3)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	// create a random buffer to act as reader
	buf := make([]byte, 5<<20) // 5MB buffer
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("cannot create buffer: %v", err)
	}
	r := io.NopCloser(bytes.NewReader(buf))

	if err := s.Save("test-delme.txt", r); err != nil {
		t.Fatalf("store.Save: %v", err)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("os.ReadDir: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("incorrect number of files; expected: 3; got: %d", len(files))
	}
}
