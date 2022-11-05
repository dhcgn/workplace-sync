package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/c-bata/go-prompt"
	"github.com/dhcgn/workplace-sync/config"
	"github.com/dhcgn/workplace-sync/downloader"
	"github.com/pterm/pterm"
	"golang.org/x/exp/slices"
)

const (
	folder = `C:\ws\`
)

var (
	hostFlag    = flag.String("host", "", "The host which TXT record is set to an url of links")
	localSource = flag.String("local", "", "The local source of links")
	allFlag     = flag.Bool("all", false, "Download all links")
)

func main() {
	fmt.Printf("Workplace Sync %v %v\n", buildInfoCommitID, buildInfoTime)
	fmt.Println("https://github.com/dhcgn/workplace-sync")
	fmt.Println()

	flag.Parse()
	if *hostFlag == "" && *localSource == "" {
		fmt.Println("host or localSource is required")
		flag.PrintDefaults()
		return
	}

	if *hostFlag != "" && *localSource != "" {
		fmt.Println("host and localSource are mutually exclusive")
		flag.PrintDefaults()
		return
	}

	var linksContainer config.LinksContainer
	if *hostFlag != "" {
		l, err := getLinks(*hostFlag)
		if err != nil {
			fmt.Println(err)
			return
		}
		linksContainer = l
	} else {
		l, err := getLinksLocal(*localSource)
		if err != nil {
			fmt.Println(err)
			return
		}
		linksContainer = l
	}

	pterm.Info.Printfln("Got %v links", len(linksContainer.Links))

	pterm.Info.Printfln("Use download folder %v", folder)
	err := createDownloadFolder(folder)
	if err != nil {
		fmt.Println(err)
		return
	}

	if *allFlag {
		pterm.Info.Printfln("All links will be downloaded:")
		for i, l := range linksContainer.Links {
			pterm.Info.Printfln("%2v. %v (%v)", (i + 1), l.GetDisplayName(), l.Version)
		}

		for _, l := range linksContainer.Links {
			err := downloader.Get(l, folder)
			if err != nil {
				pterm.Error.Printfln("link %v, folder: %v, error: %v", l.Url, folder, err)
				continue
			}
		}
		return
	}

	interaction := interaction{
		lc: linksContainer,
	}

	fmt.Println("Please select file to download:")
	t := prompt.Input("> ", interaction.completer)

	if t == "" {
		fmt.Println("No file selected")
		return
	}

	i := slices.IndexFunc(linksContainer.Links, func(l config.Link) bool {
		return l.GetDisplayName() == t
	})

	if i == -1 {
		fmt.Println("No file found, please complete the whole name.")
		return
	}

	err = downloader.Get(linksContainer.Links[i], folder)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println()
	fmt.Println("Done")
}

type interaction struct {
	lc config.LinksContainer
}

func (i interaction) completer(d prompt.Document) []prompt.Suggest {
	var suggestions []prompt.Suggest
	for _, l := range i.lc.Links {
		s := prompt.Suggest{}
		s.Text = l.GetDisplayName()
		suggestions = append(suggestions, s)
	}

	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}

func createDownloadFolder(f string) error {
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		err := os.Mkdir(f, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func getLinksLocal(f string) (config.LinksContainer, error) {
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

func getLinks(host string) (config.LinksContainer, error) {
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
