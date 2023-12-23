package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/dhcgn/workplace-sync/config"
	"github.com/dhcgn/workplace-sync/linkscontainer"
	"github.com/pterm/pterm"
)

func checkAndFilterLinksContainer(lc *config.LinksContainer, forceHashCheck, checkLink bool, infos, warns, errors chan<- string) (err error) {
	wg := sync.WaitGroup{}
	if forceHashCheck {
		var filteredLinksContainer config.LinksContainer
		for _, l := range lc.Links {
			if l.Hash == "" {
				warns <- fmt.Sprintf("Link %v (%v) has no hash, will be removed because ForceHashCheck is active", l.GetDisplayName(), l.Version)
			} else {
				filteredLinksContainer.Links = append(filteredLinksContainer.Links, l)
			}
		}
		lc = &filteredLinksContainer
	}

	for _, l := range lc.Links {
		if l.GithubReleaseAssetFilter != "" && ((l.Version != "" && l.Version != "latest") || l.Hash != "") {
			warns <- (fmt.Sprintf("Link %v (%v) has GithubReleaseAssetFilter so and Version or Hash will be ignored", l.GetDisplayName(), l.Version))
		}
		if strings.HasPrefix(l.Url, "https://github.com/") && l.GithubReleaseAssetFilter == "" {
			warns <- (fmt.Sprintf("Link %v (%v) has no GithubReleaseAssetFilter, so download is pinned and not latest release is used", l.GetDisplayName(), l.Version))
		}
	}

	validLinks := make(chan config.Link, len(lc.Links))
	if checkLink {
		// var filteredLinksContainer config.LinksContainer
		for _, l := range lc.Links {
			wg.Add(1)
			go func(link config.Link, wg *sync.WaitGroup) {
				err := checkLinkFunc(link.Url)
				if err != nil {
					errors <- fmt.Sprintf("Link %v (%v) is invalid: %v, url: %v", link.GetDisplayName(), link.Version, err, link.Url)
				} else {
					infos <- fmt.Sprintf("Link %v (%v) is valid, url: %v", link.GetDisplayName(), link.Version, link.Url)
					validLinks <- link
				}
				wg.Done()
			}(l, &wg)
		}
		wg.Wait()
		close(validLinks)

		var filteredLinksContainer config.LinksContainer

		for l := range validLinks {
			filteredLinksContainer.Links = append(filteredLinksContainer.Links, l)
		}
		lc = &filteredLinksContainer
	}

	return nil
}

func getLinksContainer(host, path, url string) (config.LinksContainer, error) {
	if host != "" {
		pterm.Info.Printfln("Obtain links from DNS TXT record of %v", host)
		l, err := linkscontainer.GetLinksDNS(host)
		if err != nil {
			fmt.Println(err)
			return config.LinksContainer{}, err
		}
		return l, nil
	}

	if path != "" {
		pterm.Info.Printfln("Obtain links from local file %v", path)
		l, err := linkscontainer.GetLinksLocal(path)
		if err != nil {
			return config.LinksContainer{}, err
		}
		return l, nil
	}

	if url != "" {
		pterm.Info.Printfln("Obtain links from url %v", url)
		l, err := linkscontainer.GetLinksURL(url)
		if err != nil {
			return config.LinksContainer{}, err
		}
		return l, nil
	}
	return config.LinksContainer{}, fmt.Errorf("host, path or url is required")
}
