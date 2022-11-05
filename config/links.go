package config

import (
	"strings"
	"time"
)

type LinksContainer struct {
	Links        []Link    `json:"links"`
	LastModified time.Time `json:"last_modified"`
}

type Link struct {
	Name             string `json:"name"`
	Url              string `json:"url"`
	Version          string `json:"version"`
	Type             string `json:"type"`
	DecompressFlat   bool   `json:"decompress_flat"`
	DecompressFilter string `json:"decompress_filter"`
}

func (l *Link) GetDisplayName() string {
	if l.Name != "" {
		return l.Name
	}
	splits := strings.Split(l.Url, "/")
	last := splits[len(splits)-1]
	return last
}
