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
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/dhcgn/workplace-sync/config"
	"github.com/dhcgn/workplace-sync/downloader"
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

	// temp := links{
	// 	Links: []link{
	// 		{
	// 			Url:     "https://download.sysinternals.com/files/SysinternalsSuite.zip",
	// 			Version: "latest",
	// 		},
	// 	},
	// 	LastModified: time.Now().UTC(),
	// }
	// d, _ := json.MarshalIndent(temp, "", "  ")
	// fmt.Println(string(d))

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

	var links config.LinksContainer
	if *hostFlag != "" {
		l, err := getLinks(*hostFlag)
		if err != nil {
			fmt.Println(err)
			return
		}
		links = l
	} else {
		l, err := getLinksLocal(*localSource)
		if err != nil {
			fmt.Println(err)
			return
		}
		links = l
	}

	if *allFlag {
		err := createFolder(folder)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, l := range links.Links {
			err := downloader.Get(l, folder)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		return
	}

	interaction := interaction{
		lc: links,
	}

	fmt.Println("Please select file to download.")
	t := prompt.Input("> ", interaction.completer)

	if t == "" {
		fmt.Println("No file selected")
		return
	}

	i := slices.IndexFunc(links.Links, func(l config.Link) bool {
		if l.Name != "" && l.Name == t {
			return true
		}
		splits := strings.Split(l.Url, "/")
		last := splits[len(splits)-1]
		return last == t
	})

	if i == -1 {
		fmt.Println("No file found")
		return
	}

	err := downloader.Get(links.Links[i], folder)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Done")
}

type interaction struct {
	lc config.LinksContainer
}

func (i interaction) completer(d prompt.Document) []prompt.Suggest {
	var suggestions []prompt.Suggest
	for _, l := range i.lc.Links {
		s := prompt.Suggest{}
		if l.Name != "" {
			s.Text = l.Name
		} else {
			splits := strings.Split(l.Url, "/")
			s.Text = splits[len(splits)-1]
		}

		suggestions = append(suggestions, s)
	}

	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}

func createFolder(f string) error {
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
