package download

import (
	"os"

	"github.com/sem-hub/ygg-peers-select/internal/mlog"
	"github.com/sem-hub/ygg-peers-select/internal/options"
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
	if options.Opts.DoNotRemove {
		mlog.GetLogger().Debug("Do not remove temp DIR: " + d.path)
		return
	}
	mlog.GetLogger().Debug("Remove DIR: " + d.path)
	os.RemoveAll(d.path)
}
