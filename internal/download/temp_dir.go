package download

import (
	"os"

	"github.com/sem-hub/ygg-peers-select/internal/mlog"
)

type tempDir struct {
	path string
}

func CreateTempDir() (tempDir, error) {
	d := new(tempDir)
	var err error
	d.path, err = os.MkdirTemp(os.TempDir(), "ygg-ping")

	mlog.GetLogger().Debug("Temp DIR: " + d.path)

	return *d, err
}

func (d *tempDir) Delete() {
	mlog.GetLogger().Debug("Remove DIR: " + d.path)
	os.RemoveAll(d.path)
}
