package download

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/sem-hub/ygg-peers-select/internal/mlog"
	"github.com/sem-hub/ygg-peers-select/internal/utils"
)

type DownloadZip struct {
	tempDir tempDir
	workDir string
}

func (p *DownloadZip) WorkDir() string {
	return p.workDir
}

func (p *DownloadZip) Download() error {
	const fileUrl = "https://github.com/yggdrasil-network/public-peers/archive/refs/heads/master.zip"

	logger := mlog.GetLogger()

	logger.Info("Downloading peers from: " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(fileUrl))

	var err error
	p.tempDir, err = CreateTempDir()
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(p.tempDir.path, "master.zip"))
	if err != nil {
		log.Fatal("Can't create file: " + err.Error())
	}
	defer file.Close()

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(fileUrl)
	if err != nil {
		log.Fatal("Can't fetch file: " + err.Error())
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)

	logger.Debug("Downloaded a file with size: " + strconv.FormatInt(size, 10))

	utils.Unzip(filepath.Join(p.tempDir.path, "master.zip"), p.tempDir.path)

	logger.Debug("Unzipped")

	p.workDir = filepath.Join(p.tempDir.path, "public-peers-master")
	return err
}

func (p *DownloadZip) Cleanup() {
	p.tempDir.Delete()
}
