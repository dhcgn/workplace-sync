package linkscontainer

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/dhcgn/workplace-sync/config"
)

func GetLinksLocal(f string) (config.LinksContainer, error) {
	data, err := os.ReadFile(f)
	if err != nil {
		return config.LinksContainer{}, err
	}

	var l config.LinksContainer
	err = json.Unmarshal(data, &l)
	if err != nil {
		return config.LinksContainer{}, err
	}
	l.SortLinks()
	return l, nil
}

func GetLinksDNS(host string) (config.LinksContainer, error) {
	txts, err := net.LookupTXT(host)
	if err != nil {
		return config.LinksContainer{}, err
	}

	if len(txts) == 0 {
		return config.LinksContainer{}, fmt.Errorf("no link")
	}

	if len(txts) > 1 {
		return config.LinksContainer{}, fmt.Errorf("too many links")
	}

	u := txts[0]
	if _, err := url.Parse(u); err != nil {
		return config.LinksContainer{}, err
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	url, err := url.Parse(u)
	if err != nil {
		return config.LinksContainer{}, err
	}

	r := &http.Request{
		Method: "GET",
		URL:    url,
	}
	r.Header = map[string][]string{
		"Accept-Encoding": {"gzip, deflate"},
	}

	resp, err := client.Do(r)
	if err != nil {
		return config.LinksContainer{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return config.LinksContainer{}, err
	}

	var l config.LinksContainer
	err = json.Unmarshal(body, &l)
	if err != nil {
		return config.LinksContainer{}, err
	}
	l.SortLinks()
	return l, nil
}
