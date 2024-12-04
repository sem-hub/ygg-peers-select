package download

import (
	"os"
	"path/filepath"

	"github.com/sem-hub/ygg-peers-select/internal/options"
)

type Downloader interface {
	Download() error
	Cleanup()
	WorkDir() string
}

var (
	p Downloader = nil
)

func Download(o *options.Options) (string, error) {
	if o.WithGit {
		p = &DownloadGit{}
	} else {
		p = &DownloadZip{}
	}

	err := p.Download()
	os.Remove(filepath.Join(p.WorkDir(), "README.md"))
	return p.WorkDir(), err
}

func Cleanup() {
	p.Cleanup()
}
