package config

import "time"

type Links struct {
	Links        []Link    `json:"links"`
	LastModified time.Time `json:"last_modified"`
}

type Link struct {
	Url              string `json:"url"`
	Version          string `json:"version"`
	Type             string `json:"type"`
	DecompressFlat   bool   `json:"decompress_flat"`
	DecompressFilter string `json:"decompress_filter"`
}
