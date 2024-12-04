package guesscountry

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sem-hub/ygg-peers-select/internal/interactive"
	"github.com/sem-hub/ygg-peers-select/internal/mlog"
)

type apiData struct {
	Query     string `json:"query"`
	Status    string `json:"status"`
	Continent string `json:"continent"`
	Country   string `json:"country"`
}

func GetCountryByIP(workDir string, dontAsk bool) (string, error) {
	const apiUrl = "http://ip-api.com/json/?fields=query,status,continent,country"

	logger := mlog.GetLogger()

	logger.Info("Get country with " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render("ip-api.com"))

	// Ask API for country guess
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(apiUrl)
	if err != nil {
		logger.Error("Can't fetch file: " + err.Error())
		return "", err
	}
	defer resp.Body.Close()

	var data apiData
	json.NewDecoder(resp.Body).Decode(&data)
	if data.Status != "success" {
		logger.Error("Can't get IP info: " + data.Status)
		return "", err
	}

	countryGuessed := filepath.Join(strings.ToLower(data.Continent), strings.ToLower(data.Country))

	path := filepath.Join(workDir, countryGuessed+".md")

	// Check if file exists
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	f.Close()

	if dontAsk {
		logger.Info("Use country: " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("211")).Render(countryGuessed))
		return path, nil
	}

	confirmed, err := interactive.Ask("Is it correct: ", countryGuessed)
	if err != nil {
		logger.Error(err.Error())
		return "", err

	}

	// User answered "no"
	if !confirmed {
		return "", nil
	}

	return path, nil
}
