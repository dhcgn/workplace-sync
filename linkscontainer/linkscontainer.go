package linkscontainer

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"

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

	resp, err := http.Get(u)
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
	return l, nil
}
