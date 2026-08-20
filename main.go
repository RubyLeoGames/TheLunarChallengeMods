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
type Mod struct {
	Name        string
	Description string
	DownloadURL string `json:"download_url"`
}

func main() {
	topic := "tlc-mod"
	base_url := "https://api.github.com/search/repositories"
	filename := "mods.json"
	archiveFormat := "zipball"

	u, err := url.Parse(base_url)
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
	var data SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	repos := data.Items
	log.Printf("Processing %d repos", len(repos))
	mods := make([]Mod, len(repos))
	for i, repo := range repos {
		mods[i].Name = repo.Name
		mods[i].Description = repo.Description
		u := repo.ArchiveURL
		u = strings.ReplaceAll(u, "{archive_format}", archiveFormat)
		u = strings.ReplaceAll(u, "{/ref}", "/"+repo.DefaultBranch)
		mods[i].DownloadURL = u
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
