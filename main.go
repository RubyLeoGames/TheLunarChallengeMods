package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
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

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

var rawContent = mustParseURL("https://raw.githubusercontent.com")

const archiveFormat = "zipball"
const infoFile = "info.json"

func main() {
	topic := "tlc-mod"
	baseURL := "https://api.github.com/search/repositories?per_page=100"
	filename := "mods.json"
	nextPagePattern, err := regexp.Compile(`<(\S*)>; rel=\"next\"`)
	if err != nil {
		log.Fatal(err)
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		log.Fatal(err)
	}
	params := parsed.Query()
	params.Add("q", "topic:"+topic)
	parsed.RawQuery = params.Encode()
	u := parsed.String()
	morePages := true

	mods := make([]Mod, 0)
	for morePages {
		log.Printf("Fetching %s", u)
		resp, err := http.Get(u)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		var data SearchResponse
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			log.Fatal(err)
		}
		var link = resp.Header["Link"]
		if len(link) > 0 {
			matches := nextPagePattern.FindStringSubmatch(link[0])
			if len(matches) > 1 {
				u = matches[1]
			} else {
				morePages = false
			}
		} else {
			morePages = false
		}

		repos := data.Items
		log.Printf("Processing %d repos", len(repos))
		for _, repo := range repos {
			mod, err := ProcessRepo(repo)
			if err != nil {
				log.Printf("Skipping %s: %s", repo.FullName, err)
				continue
			}
			mods = append(mods, mod)
			log.Printf("Added %s as mod", repo.FullName)
		}
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

func ProcessRepo(repo Repo) (Mod, error) {
	rawContent.Path = repo.FullName + "/refs/heads/" + repo.DefaultBranch + "/" + infoFile
	log.Printf("Fetching %s", rawContent.String())
	resp, err := http.Get(rawContent.String())
	var mod Mod
	if err != nil {
		return mod, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return mod, fmt.Errorf("Couldn't fetch %s: %s", infoFile, resp.Status)
	}
	var modInfo ModInfo
	err = json.NewDecoder(resp.Body).Decode(&modInfo)
	if err != nil {
		return mod, err
	}

	mod.ModInfo = modInfo
	u := repo.ArchiveURL
	u = strings.ReplaceAll(u, "{archive_format}", archiveFormat)
	u = strings.ReplaceAll(u, "{/ref}", "/"+repo.DefaultBranch)
	mod.DownloadURL = u
	return mod, nil
}
