package foxy

import (
	"io"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/mkfoss/foxy/internal/mockfiler"
	"github.com/stretchr/testify/assert"
)

type BytesOpener struct {
	data []byte
}

func (bop *BytesOpener) OpenFile(name string, flag int, perm os.FileMode) (io.ReadSeekCloser, error) {
	fl := mockfiler.NewMockFiler()
	fl.Data = bop.data
	return fl, nil
}

func PackageRoot(t *testing.T) string {
	t.Helper()
	_, appdir, _, ok := runtime.Caller(0)
	if ok {
		return path.Clean(path.Dir(appdir))
	}
	t.Fatal("testutil failure: could not determine package root")
	return ""
}

func TesrDataDir(t *testing.T) string {
	t.Helper()
	return path.Join(PackageRoot(t), "testdata")
}

func Test_CauseIAmStupidlyParanoidandSometimesDontTrustTheDocsandCommonSenseandamCompulsive(t *testing.T) {
	assert.NotEmpty(t, PackageRoot(t))
	assert.NotEmpty(t, TesrDataDir(t))
}
