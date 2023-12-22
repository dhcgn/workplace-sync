package config

import (
	"net/url"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

type LinksContainer struct {
	MinVersion   string    `json:"min_version"`
	Links        []Link    `json:"links"`
	LastModified time.Time `json:"last_modified"`
}

type Link struct {
	Name                     string            `json:"name"`
	Url                      string            `json:"url"`
	Version                  string            `json:"version"`
	Type                     string            `json:"type"`
	Hash                     string            `json:"hash"`
	DecompressFlat           bool              `json:"decompress_flat"`
	DecompressFilter         string            `json:"decompress_filter"`
	Skipped                  bool              `json:"skipped"`
	OverwriteFilesNames      map[string]string `json:"overwrite_files_names"`
	GithubReleaseAssetFilter string            `json:"github_release_asset_filter"`
}

func (l *Link) GetDisplayName() string {
	if l.Name != "" {
		return l.Name
	}
	splits := strings.Split(l.Url, "/")
	last := splits[len(splits)-1]
	return last
}

func (l *Link) GetHostFromLink() string {
	if l.Url == "" {
		return ""
	}
	u, err := url.Parse(l.Url)
	if err != nil {
		return ""
	}
	return u.Host
}

func (lc *LinksContainer) SortLinks() {
	slices.SortFunc(lc.Links, func(a, b Link) bool {
		return a.GetDisplayName() < b.GetDisplayName()
	})
}

func (lc LinksContainer) GetLinksByDisplayNamePreffix(name string) []Link {
	var match []Link
	for _, l := range lc.Links {
		if strings.HasPrefix(strings.ToLower(l.GetDisplayName()), strings.ToLower(name)) {
			match = append(match, l)
		}
	}

	return match
}
