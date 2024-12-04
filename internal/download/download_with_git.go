package download

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/sem-hub/ygg-peers-select/internal/mlog"
	git "gopkg.in/src-d/go-git.v4"
)

type DownloadGit struct {
	tempDir tempDir
	workDir string
}

func (p *DownloadGit) WorkDir() string {
	return p.workDir
}

func (p *DownloadGit) Download() error {
	const repoUrl = "https://github.com/yggdrasil-network/public-peers.git"

	logger := mlog.GetLogger()

	var err error
	p.tempDir, err = CreateTempDir()
	if err != nil {
		return err
	}

	logger.Info("Clonning repository from URL: " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(repoUrl))

	_, err = git.PlainClone(p.tempDir.path, false, &git.CloneOptions{
		URL:      repoUrl,
		Progress: os.Stdout,
	})

	if err != nil {
		log.Fatal("git clone error")
	}
	fmt.Println("git clone " + lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("OK"))

	p.workDir = p.tempDir.path
	return nil
}

func (p *DownloadGit) Cleanup() {
	p.tempDir.Delete()
}
