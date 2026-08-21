package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type SearchResponse struct {
	Items []Repo
}
type Repo struct {
	Name          string
	Description   string
	FullName      string `json:"full_name"`
	ArchiveURL    string `json:"archive_url"`
	DefaultBranch string `json:"default_branch"`
}
type ModInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

type Mod struct {
	ModInfo
	DownloadURL string `json:"download_url"`
}

func main() {
	topic := "tlc-mod"
	baseURL := "https://api.github.com/search/repositories"
	rawContent, err := url.Parse("https://raw.githubusercontent.com")
	if err != nil {
		log.Fatal(err)
	}
	filename := "mods.json"
	archiveFormat := "zipball"

	u, err := url.Parse(baseURL)
	if err != nil {
		log.Fatal(err)
	}
	params := u.Query()
	params.Add("q", "topic:"+topic)
	u.RawQuery = params.Encode()

	log.Printf("Fetching %s", u.String())
	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var data SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	repos := data.Items
	log.Printf("Processing %d repos", len(repos))
	mods := make([]Mod, 0, len(repos))
	for _, repo := range repos {
		rawContent.Path = repo.FullName + "/refs/heads/" + repo.DefaultBranch + "/info.json"
		log.Printf("Fetching %s", rawContent.String())
		resp, err := http.Get(rawContent.String())
		if err != nil {
			log.Printf("Failed fetching: %s", err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("Returned error: %s", resp.Status)
			continue
		}
		var modInfo ModInfo
		err = json.NewDecoder(resp.Body).Decode(&modInfo)
		if err != nil {
			log.Printf("Failed parsing: %s", err)
			continue
		}

		var mod Mod
		mod.ModInfo = modInfo
		u := repo.ArchiveURL
		u = strings.ReplaceAll(u, "{archive_format}", archiveFormat)
		u = strings.ReplaceAll(u, "{/ref}", "/"+repo.DefaultBranch)
		mod.DownloadURL = u
		mods = append(mods, mod)
		log.Printf("Added %s as mod", repo.FullName)
	}
	log.Printf("Saving %d mods to %s", len(mods), filename)

	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	e := json.NewEncoder(file)
	err = e.Encode(mods)
	if err != nil {
		log.Fatal(err)
	}
}
