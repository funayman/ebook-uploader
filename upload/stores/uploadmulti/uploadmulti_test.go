package uploadmulti

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/funayman/logger"
	"go.uber.org/zap"
)

var (
	log *zap.SugaredLogger
)

func init() {
	log, _ = logger.New("TESTS")
}

type testUploader struct{}

func (tu testUploader) Save(name string, src io.ReadCloser) error {
	f, err := os.CreateTemp("", "test")
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
	tu1 := testUploader{}
	tu2 := testUploader{}
	tu3 := testUploader{}

	s, err := NewStore(log, tu1, tu2, tu3)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	r := strings.NewReader("this is a test reader!")

	if err := s.Save("test-delme.txt", io.NopCloser(r)); err != nil {
		t.Fatalf("store.Save: %v", err)
	}
}
