package foxy

import (
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func PackageRoot(t *testing.T) string {
	t.Helper()
	_, appdir, _, ok := runtime.Caller(0)
	if ok {
		return path.Clean(path.Dir(appdir))
	}
	t.Fatal("testutil failure: could not determine package root")
	return ""
}

func DataDir(t *testing.T) string {
	t.Helper()
	return path.Join(PackageRoot(t), "testdata")
}

func Test_CauseIAmStupidlyParanoidandSometimesDontTrustTheDocsandCommonSenseandamCompulsive(t *testing.T) {
	assert.NotEmpty(t, PackageRoot(t))
	assert.NotEmpty(t, DataDir(t))
}
